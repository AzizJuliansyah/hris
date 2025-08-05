package controllers

import (
	"database/sql"
	"fmt"
	"hris/config"
	"hris/entities"
	"hris/helpers"
	"hris/models"
	"hris/services/sessiondata"
	"hris/views"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type KPIController struct {
	db *sql.DB
}

func NewKPIController(db *sql.DB) *KPIController {
	return &KPIController{db: db}
}

func (controller *KPIController) FindAllKPI(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/kpi/kpi.html",
	))
	data := make(map[string]interface{})
	
	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

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

	todayKPI := request.URL.Query().Get("today_kpi") == "true"
	data["todayKPI"] = todayKPI

	kpiModel := models.NewKPIModel(controller.db)
	kpis, err := kpiModel.FindKPIList("", selectedMonth, todayKPI)
	if err != nil {
		data["error"] = "Gagal mengambil data kpi: " + err.Error()
		log.Println("error find all employee: ", err.Error())
	} else {
		data["kpis"] = kpis
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func (controller *KPIController) AddKPI(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/kpi/add-kpi.html",
	))
	data := make(map[string]interface{})

	session, _ := config.Store.Get(request, config.SESSION_ID)
	sessionNIK := session.Values["nik"].(string)
	
	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	salaryModel := models.NewSalaryModel(controller.db)
	employees, err := salaryModel.GetEmployeeNameandNIK()
	if err != nil {
		log.Println("Error Getting Employee NIK and Name", err)
		return
	}
	data["employees"] = employees

	data["kpi"] = entities.AddKPI{}

	if request.Method == http.MethodPost {
		request.ParseForm()

		target_valueStr := request.Form.Get("target_value")
		target_value, _ := strconv.ParseInt(target_valueStr, 10, 64)

		deadlineInput := request.Form.Get("deadline")
		deadline, _ := time.Parse("2006-01-02T15:04", deadlineInput)

		kpi := entities.AddKPI{
			Employee_NIK: request.Form.Get("employee_nik"),
			Admin_NIK: sessionNIK,
			Title: request.Form.Get("title"),
			Description: request.Form.Get("description"),
			Target_Value: target_value,
			Unit: request.Form.Get("unit"),
			Deadline: deadline,
		}
		errorMessages := helpers.NewValidation().Struct(kpi)
		if errorMessages != nil {
			data["validation"] = errorMessages
			data["kpi"] = kpi
			data["currentPath"] = request.URL.Path
			templateLayout.ExecuteTemplate(httpWriter, "base", data)
			return
		}

		kpiModel := models.NewKPIModel(controller.db)
		err = kpiModel.AddKPI(kpi)
		if err != nil {
			data["error"] = "Gagal menambahkan KPI: " + err.Error()
		} else {
			session.AddFlash("Berhasil menambahkan KPI baru.", "success")
			session.Save(request, httpWriter)
			http.Redirect(httpWriter, request, "/kpi", http.StatusSeeOther)
		}
	}
	
	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)

}

func (controller *KPIController) EditKPI(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/kpi/edit-kpi.html",
	))
	data := make(map[string]interface{})

	session, _ := config.Store.Get(request, config.SESSION_ID)
	sessionNIK := session.Values["nik"].(string)

	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	id := request.URL.Query().Get("id")
	int64Id, _ := strconv.ParseInt(id, 10, 64)
	if id == "" {
		http.Error(httpWriter, "ID tidak ditemukan", http.StatusBadRequest)
		return
	}
	
	salaryModel := models.NewSalaryModel(controller.db)
	employees, err := salaryModel.GetEmployeeNameandNIK()
	if err != nil {
		log.Println("Error Getting Employee NIK and Name", err)
		return
	}
	data["employees"] = employees

	kpiModel := models.NewKPIModel(controller.db)
	kpi, err := kpiModel.FindKPIByID(int64Id)
	if err != nil {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Gagal mendapatkan data KPI! " + err.Error(), "error")
		session.Save(request, httpWriter)
	
		http.Redirect(httpWriter, request, "/kpi", http.StatusSeeOther)
	} else {
		data["kpi"] = kpi
	}

	if request.Method == http.MethodPost {
		request.ParseForm()

		target_valueStr := request.Form.Get("target_value")
		target_value, _ := strconv.ParseInt(target_valueStr, 10, 64)

		deadlineInput := request.Form.Get("deadline")
		deadline, _ := time.Parse("2006-01-02T15:04", deadlineInput)

		kpi := entities.EditKPI{
			ID: int64Id,
			Employee_NIK: request.Form.Get("employee_nik"),
			Updated_Admin_NIK: sessionNIK,
			Title: request.Form.Get("title"),
			Description: request.Form.Get("description"),
			Target_Value: target_value,
			Unit: request.Form.Get("unit"),
			Deadline: deadline,
			Status: request.Form.Get("status"),
		}
		errorMessages := helpers.NewValidation().Struct(kpi)
		if errorMessages != nil {
			data["validation"] = errorMessages
			data["kpi"] = kpi
			data["currentPath"] = request.URL.Path
			templateLayout.ExecuteTemplate(httpWriter, "base", data)
			return
		}

		kpiModel := models.NewKPIModel(controller.db)
		err := kpiModel.EditKPI(kpi)

		if err != nil {
			data["error"] = "Edit data gagal: " + err.Error()
		} else {
			session.AddFlash("Berhasil mengubah data kpi.", "success")
			session.Save(request, httpWriter)
			http.Redirect(httpWriter, request, "/kpi/edit-kpi?id="+id, http.StatusSeeOther)
			
		}
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func (controller *KPIController) EvaluasiKPI(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/kpi/evaluasi-kpi.html",
	))
	data := make(map[string]interface{})

	session, _ := config.Store.Get(request, config.SESSION_ID)
	sessionNIK := session.Values["nik"].(string)

	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	id := request.URL.Query().Get("id")
	int64Id, _ := strconv.ParseInt(id, 10, 64)

	if id == "" {
		http.Error(httpWriter, "ID tidak ditemukan", http.StatusBadRequest)
		return
	}

	redirectPath := request.URL.Query().Get("path")
	if redirectPath == "" {
		redirectPath = "kpi"
	}
	data["redirectPath"] = redirectPath
	uuid := request.URL.Query().Get("uuid")
	data["uuid"] = uuid

	kpiModel := models.NewKPIModel(controller.db)
	kpi, err := kpiModel.FindKPIByID(int64Id)
	if err != nil {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Gagal mendapatkan data KPI! " + err.Error(), "error")
		session.Save(request, httpWriter)
	
		http.Redirect(httpWriter, request, "/kpi", http.StatusSeeOther)
	} else {
		data["kpi"] = kpi
	}

	if request.Method == http.MethodPost {
		request.ParseForm()

		scoreStr := request.Form.Get("score")
		score, _ := strconv.ParseInt(scoreStr, 10, 64)

		evaluasi := entities.EvaluasiKPI{
			ID: int64Id,
			Checked_Admin_NIK: sessionNIK,
			Final_Status: request.Form.Get("final_status"),
			Score: float64(score),
			Admin_Notes: request.Form.Get("admin_notes"),
		}
		errorMessages := helpers.NewValidation().Struct(evaluasi)
		if errorMessages != nil {
			data["validation"] = errorMessages
			data["evaluasi"] = evaluasi
			data["currentPath"] = request.URL.Path
			templateLayout.ExecuteTemplate(httpWriter, "base", data)
			return
		}

		kpiModel := models.NewKPIModel(controller.db)
		err := kpiModel.InputEvaluasiKPI(evaluasi)
		if err != nil {
			data["error"] = "Evaluasi kpi gagal: " + err.Error()
		} else {
			session.AddFlash("Berhasil me-evaluasi kpi.", "success")
			session.Save(request, httpWriter)
			http.Redirect(httpWriter, request, "/kpi/evaluasi-kpi?id="+id+"&path="+redirectPath, http.StatusSeeOther)
		}
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func (controller *KPIController) ResetEvaluasiKPI(httpWriter http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("id")
	int64Id, _ := strconv.ParseInt(id, 10, 64)

	if id == "" {
		http.Error(httpWriter, "ID tidak ditemukan", http.StatusBadRequest)
		return
	}

	session, _ := config.Store.Get(request, config.SESSION_ID)
	sessionNIK := session.Values["nik"].(string)

	kpiModel := models.NewKPIModel(controller.db)
	err := kpiModel.ResetEvaluasiKPI(int64Id, sessionNIK)
	if err != nil {
		http.Error(httpWriter, "Gagal mereset data: " + err.Error(), http.StatusInternalServerError)
		return
	} else {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Berhasil mereset evaluasi KPI!", "success")
		session.Save(request, httpWriter)
	}

	http.Redirect(httpWriter, request, "/kpi/evaluasi-kpi?id="+id, http.StatusSeeOther)
}


func (controller *KPIController) ListKPI(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/kpi/list-kpi.html",
	))
	data := make(map[string]interface{})

	session, _ := config.Store.Get(request, config.SESSION_ID)
	sessionNIK := session.Values["nik"].(string)
	
	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

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

	todayKPI := request.URL.Query().Get("today_kpi") == "true"
	data["todayKPI"] = todayKPI

	kpiModel := models.NewKPIModel(controller.db)
	kpis, err := kpiModel.FindKPIList(sessionNIK, selectedMonth, todayKPI)
	if err != nil {
		data["error"] = "Gagal mengambil data kpi: " + err.Error()
		log.Println("error find all employee: ", err.Error())
	} else {
		data["kpis"] = kpis
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func (controller *KPIController) DetailKPI(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/kpi/detail-kpi.html",
	))
	data := make(map[string]interface{})

	session, _ := config.Store.Get(request, config.SESSION_ID)
	sessionNIK := session.Values["nik"].(string)
	
	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	id := request.URL.Query().Get("id")
	int64Id, _ := strconv.ParseInt(id, 10, 64)
	if id == "" {
		http.Error(httpWriter, "ID tidak ditemukan", http.StatusBadRequest)
		return
	}

	kpiModel := models.NewKPIModel(controller.db)
	kpi, err := kpiModel.FindKPIByID(int64Id)
	if err != nil {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Gagal mendapatkan data KPI! " + err.Error(), "error")
		session.Save(request, httpWriter)
	
		http.Redirect(httpWriter, request, "/kpi", http.StatusSeeOther)
	} else {
		data["kpi"] = kpi
	}

	if kpi.Employee_NIK != sessionNIK {
		if session.Values["isAdmin"] == true {
			views.RenderTemplate(httpWriter, "views/static/forbidden/forbidden.html", data)
		} else {
			views.RenderTemplate(httpWriter, "views/static/forbidden/forbidden.html", data)
		}
	}

	if request.Method == http.MethodPost && request.FormValue("submit-progress") == "1"{
		id := request.URL.Query().Get("id")
		int64Id, _ := strconv.ParseInt(id, 10, 64)

		if id == "" {
			http.Error(httpWriter, "ID tidak ditemukan", http.StatusBadRequest)
			return
		}

		result, err := SubmitProgressKPI(int64Id, controller.db, httpWriter, request)
        if err != nil {
			fmt.Println(err)
            for key, value := range result {
                data[key] = value
            }
        } else {
            session.AddFlash("Berhasil submit progress tugas.", "success")
            session.Save(request, httpWriter)
            http.Redirect(httpWriter, request, "/kpi/detail-kpi?id=" + id, http.StatusSeeOther)
			return
        }
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func (controller *KPIController) StartKPI(httpWriter http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("id")
	int64Id, _ := strconv.ParseInt(id, 10, 64)

	if id == "" {
		http.Error(httpWriter, "ID tidak ditemukan", http.StatusBadRequest)
		return
	}

	session, _ := config.Store.Get(request, config.SESSION_ID)
	sessionNIK := session.Values["nik"].(string)

	kpiModel := models.NewKPIModel(controller.db)
	kpi, err := kpiModel.FindKPIByID(int64Id)
	if err != nil {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Gagal mendapatkan data KPI! " + err.Error(), "error")
		session.Save(request, httpWriter)
	
		http.Redirect(httpWriter, request, "/kpi", http.StatusSeeOther)
	}

	if kpi.Employee_NIK != sessionNIK {
		if session.Values["isAdmin"] == true {
			views.RenderTemplate(httpWriter, "views/static/forbidden/forbidden.html", nil)
		} else {
			views.RenderTemplate(httpWriter, "views/static/forbidden/forbidden.html", nil)
		}
	}

	errStart := kpiModel.StartKPI(int64Id)
	if errStart != nil {
		http.Error(httpWriter, "Gagal memulai tugas: " + errStart.Error(), http.StatusInternalServerError)
		return
	} else {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Berhasil memulai tugas!", "success")
		session.Save(request, httpWriter)
	}

	http.Redirect(httpWriter, request, "/kpi/detail-kpi?id="+id, http.StatusSeeOther)
}

func (controller *KPIController) FinishKPI(httpWriter http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("id")
	int64Id, _ := strconv.ParseInt(id, 10, 64)

	if id == "" {
		http.Error(httpWriter, "ID tidak ditemukan", http.StatusBadRequest)
		return
	}

	session, _ := config.Store.Get(request, config.SESSION_ID)
	sessionNIK := session.Values["nik"].(string)

	kpiModel := models.NewKPIModel(controller.db)
	kpi, err := kpiModel.FindKPIByID(int64Id)
	if err != nil {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Gagal mendapatkan data KPI! " + err.Error(), "error")
		session.Save(request, httpWriter)
	
		http.Redirect(httpWriter, request, "/kpi", http.StatusSeeOther)
	}

	if kpi.Employee_NIK != sessionNIK {
		if session.Values["isAdmin"] == true {
			views.RenderTemplate(httpWriter, "views/static/forbidden/forbidden.html", nil)
		} else {
			views.RenderTemplate(httpWriter, "views/static/forbidden/forbidden.html", nil)
		}
	}

	errFinish := kpiModel.FinishKPI(int64Id)
	if errFinish != nil {
		http.Error(httpWriter, "Gagal menyelesaikan tugas: " + errFinish.Error(), http.StatusInternalServerError)
		return
	} else {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Berhasil menyelesaikan tugas!", "success")
		session.Save(request, httpWriter)
	}

	http.Redirect(httpWriter, request, "/kpi/detail-kpi?id="+id, http.StatusSeeOther)
}

func SubmitProgressKPI(id int64, db *sql.DB, httpWriter http.ResponseWriter, request *http.Request) (map[string]interface{}, error) {

	errors := make(map[string]interface{})

	request.ParseMultipartForm(5 << 20)

	photoname := ""
	file, handler, err := request.FormFile("evidence_file")
	if err == nil {
		defer file.Close()

		if handler.Size > 2*1024*1024 {
			return map[string]interface{}{
				"error": "Ukuran file maksimal 2MB",
			}, fmt.Errorf("size validation error")
		}
		ext := strings.ToLower(filepath.Ext(handler.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			return map[string]interface{}{
				"error": "Tipe file harus jpg, jpeg, png, atau webp",
			}, fmt.Errorf("image type validation error")
		}

		kpiModels := models.NewKPIModel(db)
		oldPhoto, err := kpiModels.GetKPIEvidenceFileByID(id)
		if err != nil {
			return map[string]interface{}{
				"error" : "Gagal mendapatkan data Photo.",
			}, err
		}

		// Simpan file ke public/images/kpi_evidence
		filename := fmt.Sprintf("kpi_evidence_%d%s", time.Now().UnixNano(), ext)
		path := filepath.Join("public/images/kpi_evidence", filename)
		if oldPhoto.Valid && oldPhoto.String != "" {
			oldPath := filepath.Join("public/images/kpi_evidence", oldPhoto.String)
			if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
				return map[string]interface{}{
					"error" : "Gagal menghapus foto lama.",
				}, err
			}
		}

		out, err := os.Create(path)
		if err != nil {
			return map[string]interface{}{
				"error" : "Gagal menyimpan foto.",
			}, err
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			return map[string]interface{}{
				"error" : "Gagal menyimpan file.",
			}, err
		}
		photoname = filename
	} else {
		// Tidak ada file baru: pakai foto lama
		kpiModels := models.NewKPIModel(db)
		oldPhoto, err := kpiModels.GetKPIEvidenceFileByID(id)
		if err == nil && oldPhoto.Valid {
			photoname = oldPhoto.String
		}
	}
	if photoname == "" {
		return map[string]interface{}{
			"error": "File bukti (evidence_file) wajib diisi.",
		}, fmt.Errorf("file bukti wajib diisi")
	}

	actual_valueStr := request.Form.Get("actual_value")
	actual_value, _ := strconv.ParseInt(actual_valueStr, 10, 64)

	kpiModels := models.NewKPIModel(db)
	target_valueStr, _ := kpiModels.GetKPITargetValueByID(id)
	target_value, _ := strconv.ParseInt(target_valueStr, 10, 64)

	if actual_value > target_value {
		return map[string]interface{}{
			"error": "Progress value tidak bisa melebihi target value!",
		}, fmt.Errorf("validation Error")
	}

	progress := entities.SubmitProgressKPI{
		ID: id, 
		Actual_Value: int(actual_value),
		Evidence_File: photoname,
	}
	validationResult := helpers.NewValidation().Struct(progress)
	if validationResult != nil {
		errors = validationResult.(map[string]interface{})
		return map[string]interface{}{
			"validation": errors,
			"progress": progress,
		}, fmt.Errorf("validation Error")
	}

	errInsert := kpiModels.SubmitProgressKPI(progress)
	if errInsert != nil {
		return map[string]interface{}{
			"error": "Gagal menambahkan progres tugas" + errInsert.Error(),
		}, errInsert
	}

	return nil, nil
}

func (controller *KPIController) SoftDeleteKPI(httpWriter http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("id")
	int64Id, _ := strconv.ParseInt(id, 10, 64)

	if id == "" {
		http.Error(httpWriter, "ID tidak ditemukan", http.StatusBadRequest)
		return
	}

	kpiModel := models.NewKPIModel(controller.db)
	err := kpiModel.SoftDeleteKPI(int64Id)
	if err != nil {
		http.Error(httpWriter, "Gagal menghapus data: " + err.Error(), http.StatusInternalServerError)
		return
	} else {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Berhasil menghapus KPI!", "success")
		session.Save(request, httpWriter)
	}

	http.Redirect(httpWriter, request, "/kpi", http.StatusSeeOther)
}