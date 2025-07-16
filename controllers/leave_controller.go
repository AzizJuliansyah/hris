package controllers

import (
	"database/sql"
	"fmt"
	"hris/config"
	"hris/entities"
	"hris/helpers"
	"hris/models"
	"hris/services/sessiondata"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type LeaveController struct {
	db *sql.DB
}

func NewLeaveController(db *sql.DB) *LeaveController {
	return &LeaveController{db: db}
}

func (controller *LeaveController) ListLeave(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/leave/leave-list.html",
	))
	data := make(map[string]interface{})

	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	// Hitung 5 bulan terakhir
	currentDate := time.Now()
	var months []string
	for i := 0; i < 5; i++ {
		previousMonth := currentDate.AddDate(0, -i, 0)
		months = append(months, previousMonth.Format("January 2006"))
	}
	data["months"] = months

	selectedMonth := request.URL.Query().Get("month")
	if selectedMonth == "" {
		selectedMonth = currentDate.Format("January 2006")
	}
	todayLeave := request.URL.Query().Get("today_leave") == "true"
	data["todayLeave"] = todayLeave

	leaveModel := models.NewLeaveModel(controller.db)
	leaveList, err := leaveModel.GetLeaveList("", selectedMonth, todayLeave)
	if err != nil {
		log.Println("Error getting leave list:", err)
		data["error"] = "Failed to get leave list"
	}

	data["leaves"] = leaveList
	data["selectedMonth"] = selectedMonth

	if request.Method == http.MethodGet {
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}
}

func (controller *LeaveController) LeaveType(httpWriter http.ResponseWriter, request *http.Request) {
    templateLayout := template.Must(template.ParseFiles(
        "views/static/layouts/base.html",
        "views/static/layouts/header.html",
        "views/static/layouts/navbar.html",
        "views/static/layouts/sidebar.html",
        "views/static/layouts/footer.html",
        "views/static/layouts/footer_js.html",
        "views/static/leave/leave-type.html",
    ))
    data := make(map[string]interface{})
    session, _ := config.Store.Get(request, config.SESSION_ID)
    errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
    if errSession != nil {
        log.Println("SetUserSessionData error:", errSession.Error())
    }
    
    if request.Method == http.MethodPost && request.FormValue("add-leave-type") == "1" {
        result, err := AddLeaveType(controller.db, httpWriter, request)
        if err != nil {
            for key, value := range result {
                data[key] = value
                fmt.Println("Error data:", key, value)
            }
        } else {
            session.AddFlash("Berhasil menambahkan tipe cuti.", "success")
            session.Save(request, httpWriter)
            http.Redirect(httpWriter, request, request.URL.Path, http.StatusSeeOther)
            return
        }
    }
    
    if request.Method == http.MethodPost && request.FormValue("edit-leave-type") == "1" {
        result, err := EditLeaveType(controller.db, httpWriter, request)
        if err != nil {
			fmt.Println(err)
            for key, value := range result {
                data[key] = value
                fmt.Println("Error data:", key, value)
            }
			data["error"] = err
        } else {
            session.AddFlash("Berhasil mengubah data tipe cuti.", "success")
            session.Save(request, httpWriter)
            http.Redirect(httpWriter, request, request.URL.Path, http.StatusSeeOther)
            return
        }
    }
    
	leaveModel := models.NewLeaveModel(controller.db)
    leaveType, errLeaveType := leaveModel.FindAllLeaveType()
    if errLeaveType != nil {
        data["error"] = "Gagal menampilkan tipe cuti: " + errLeaveType.Error()
    } else {
        data["leaveType"] = leaveType
    }
    
    data["currentPath"] = request.URL.Path
    templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func AddLeaveType(db *sql.DB, httpWriter http.ResponseWriter, request *http.Request) (map[string]interface{}, error)  {
	errors := make(map[string]interface{})

	request.ParseForm()

	addLeave := entities.LeaveType{
		Name: request.Form.Get("name"),
		MaxDay: request.Form.Get("max_day"),
	}

	validationResult := helpers.NewValidation().Struct(addLeave)
	if validationResult != nil {
		errors = validationResult.(map[string]interface{})
		fmt.Println(errors)
		return map[string]interface{}{
			"validationaddLeave": errors,
			"addLeave": addLeave,
		}, fmt.Errorf("validation Error")
	}

	leaveModel := models.NewLeaveModel(db)
	errInsert := leaveModel.AddLeaveType(addLeave)
	if errInsert != nil {
		return map[string]interface{}{
			"error": "Gagal menambahkan tipe cuti" + errInsert.Error(),
		}, errInsert
	}

	return nil, nil
}

func EditLeaveType(db *sql.DB, httpWriter http.ResponseWriter, request *http.Request) (map[string]interface{}, error) {
	errors := make(map[string]interface{})

	request.ParseForm()

	id := request.FormValue("id")
	LeaveTypeId, _ := strconv.ParseInt(id, 10, 64)
	editLeave := entities.LeaveType{
		Id: LeaveTypeId,
		Name: request.Form.Get("name"),
		MaxDay: request.Form.Get("max_day"),
	}

	validationResult := helpers.NewValidation().Struct(editLeave)
	if validationResult != nil {
		errors = validationResult.(map[string]interface{})
		fmt.Println(errors)
		return map[string]interface{}{
			"validationeditLeave": errors,
			"editLeave": editLeave,
		}, fmt.Errorf("validation Error")
	}

	leaveModel := models.NewLeaveModel(db)
	errInsert := leaveModel.EditLeaveType(editLeave)
	if errInsert != nil {
		return map[string]interface{}{
			"error": "Gagal mengubah data tipe cuti" + errInsert.Error(),
		}, errInsert
	}

	return nil, nil
}

func (controller *LeaveController) DeleteLeaveType(httpWriter http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("delete_id")
	int64Id, _ := strconv.ParseInt(id, 10, 64)

	if id == "" {
		http.Error(httpWriter, "ID tidak ditemukan", http.StatusBadRequest)
		return
	}

	leaveModel := models.NewLeaveModel(controller.db)
	err := leaveModel.DeleteLeaveType(int64Id)
	if err != nil {
		http.Error(httpWriter, "Gagal menghapus data: "+err.Error(), http.StatusInternalServerError)
		return
	} else {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Berhasil menghapus tipe cuti!", "success")
		session.Save(request, httpWriter)
	}

	http.Redirect(httpWriter, request, "/leave/leave-type", http.StatusSeeOther)
}

func (controller *LeaveController) SubmitLeave(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/leave/leave-submit.html",
	))
	data := make(map[string]interface{})

	session, _ := config.Store.Get(request, config.SESSION_ID)
	sessionNIK := session.Values["nik"].(string)
	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	leaveModel := models.NewLeaveModel(controller.db)
	leaveType, _ := leaveModel.FindAllLeaveType()
	data["leaveType"] = leaveType

	currentDate := time.Now()
	var months []string
	for i := 0; i < 5; i++ {
		previousMonth := currentDate.AddDate(0, -i, 0)
		months = append(months, previousMonth.Format("January 2006"))
	}
	data["months"] = months
	selectedMonth := request.URL.Query().Get("month")
	if selectedMonth == "" {
		selectedMonth = currentDate.Format("January 2006")
	}
	data["selectedMonth"] = selectedMonth

	todayLeave := request.URL.Query().Get("today_leave") == "true"
	data["todayLeave"] = todayLeave

	getLeaveList(controller.db, data, sessionNIK, selectedMonth, todayLeave)

	if request.Method == http.MethodGet {
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	request.ParseForm()

	var attachment *string
	if attachmentValue := request.Form.Get("attachment_photo"); attachmentValue != "" {
		attachment = &attachmentValue
	}

	rawLeaveDates := request.Form["leave_date[]"]
	leaveDate := cleanLeaveDates(rawLeaveDates)
	submitLeave := entities.SubmitLeave{
		NIK:           sessionNIK,
		LeaveTypeId:   request.Form.Get("leave_type_id"),
		LeaveDate:     leaveDate,
		LeaveDateJoin: strings.Join(leaveDate, ","),
		Reason:        request.Form.Get("reason"),
		Attachment:    attachment,
		Status:        helpers.PENDING_LEAVE,
	}

	validationErrors := helpers.NewValidation().Struct(submitLeave)

	if validationErrors != nil {
		data["validation"] = validationErrors
		data["leave"] = submitLeave
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	isValid, errMsg := helpers.IsLeaveDateValid(leaveDate)
	if !isValid {
		data["error"] = errMsg
		data["leave"] = submitLeave
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	errSubmit := leaveModel.InsertLeave(submitLeave)
	if (errSubmit) != nil {
		data["error"] = "Error " + errSubmit.Error()
	} else {
		data["success"] = "Pengajuan cuti berhasil, silahkan tunggu persetujuan dari Admin"
		getLeaveList(controller.db, data, sessionNIK, selectedMonth, todayLeave)
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func getLeaveList(db *sql.DB, data map[string]interface{}, sessionNIK string, selectedMonth string, todayLeave bool) {
	leaveModel := models.NewLeaveModel(db)
	leaveList, err := leaveModel.GetLeaveList(sessionNIK, selectedMonth, todayLeave)
	if err != nil {
		log.Println("Error getting leave list:", err)
		data["error"] = "Failed to get leave list"
	}
	data["leaves"] = leaveList
}

func cleanLeaveDates(input []string) []string {
	var cleaned []string
	for _, date := range input {
		trimmed := strings.TrimSpace(date)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}

func (controller *LeaveController) ApprovalLeave(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/leave/leave-approval.html",
	))

	idStr := request.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	data := make(map[string]interface{})
	session, _ := config.Store.Get(request, config.SESSION_ID)
	sessionNIK := session.Values["nik"].(string)
	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}
	
	getLeave(controller.db, id, httpWriter, data, request)

	if request.Method == http.MethodGet {
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	request.ParseForm()

	status := request.Form.Get("status")
	reason := request.Form.Get("reason_status")

	idInt64, _ := strconv.ParseInt(idStr, 10, 64)
	statusInt64, _ := strconv.ParseInt(status, 10, 64)
	approval := entities.ApprovalLeave{
		Id:           idInt64,
		AdminNIK:	  sessionNIK,
		Status:       statusInt64,
		ReasonStatus: reason,
		UpdatedAt:    time.Now(),
	}

	errorValidation := helpers.NewValidation().Struct(approval)

	if errorValidation != nil {
		data["validation"] = errorValidation
		data["approval"] = approval
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	leaveModel := models.NewLeaveModel(controller.db)
	errApprove := leaveModel.UpdateLeaveStatus(approval)
	if errApprove != nil {
		data["error"] = "Gagal memproses cuti " + errApprove.Error()
	} else {
		data["success"] = "Cuti berhasil diproses"
		getLeave(controller.db, id, httpWriter, data, request)
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func getLeave(db *sql.DB, id int64, httpWriter http.ResponseWriter, data map[string]interface{}, request *http.Request) {
	leaveModel := models.NewLeaveModel(db)
	leave, err := leaveModel.GetLeaveById(id)
	if err != nil || leave == nil {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Gagal mendapatkan data pengajuan cuti!" + err.Error(), "error")
		session.Save(request, httpWriter)
	
		http.Redirect(httpWriter, request, "/leave-list", http.StatusSeeOther)
	}
	data["leave"] = leave
}