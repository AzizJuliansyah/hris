package controllers

import (
	"database/sql"
	"hris/config"
	"hris/entities"
	"hris/helpers"
	"hris/models"
	"hris/services/sessiondata"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

type ShiftController struct {
	db *sql.DB
}

func NewShiftController(db *sql.DB) *ShiftController {
	return &ShiftController{db: db}
}

func (controller *ShiftController) FindAllShift(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/shift/shift.html",
	))

	if request.Method == http.MethodGet {

		var data = make(map[string]interface{})

		shiftModel := models.NewShiftModel(controller.db)
		shifts, err := shiftModel.FindAllShift()
		if err != nil {
			data["error"] = "Terdapat kesahalan saat menampilkan data shift " + err.Error()
			log.Println("error :", err.Error())
		} else {
			data["shift"] = shifts
		}

		errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
		if errSession != nil {
			log.Println("SetUserSessionData error:", errSession.Error())
		}

		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}
}

func (controller *ShiftController) AddShift(httpWriter http.ResponseWriter, request *http.Request) {

	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/shift/add-shift.html",
	))
	var data = make(map[string]interface{})
	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	if request.Method == http.MethodGet {
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}


	request.ParseForm()
	startTimeStr := request.Form.Get("start_time")
	endTimeStr := request.Form.Get("end_time")

	shift := entities.Shift{
		Name:      request.Form.Get("name"),
		StartTime: startTimeStr,
		EndTime:   endTimeStr,
	}

	errorMessages := helpers.NewValidation().Struct(shift)

	if errorMessages != nil {
		data["validation"] = errorMessages
		data["shift"] = shift
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	layout := "15:04"
	startTimeParsed, _ := time.Parse(layout, startTimeStr)
	endTimeParsed, _ := time.Parse(layout, endTimeStr)

	shift.StartTime = startTimeParsed.Format("15:04:05")
	shift.EndTime = endTimeParsed.Format("15:04:05")

	shiftModel := models.NewShiftModel(controller.db)
	err := shiftModel.AddShift(shift)
	if err != nil {
		data["error"] = "Gagal menambahkan shift: " + err.Error()
	} else {
		data["success"] = "Berhasil menambahkan shift"
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)

}

func (controller *ShiftController) EditShift(httpWriter http.ResponseWriter, request *http.Request) {

	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/shift/edit-shift.html",
	))
	

	var data = make(map[string]interface{})
	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	id := request.URL.Query().Get("id")
	int64Id, _ := strconv.ParseInt(id, 10, 64)
	if request.Method == http.MethodGet {
		// Ambil data shift berdasarkan ID
		shiftModel := models.NewShiftModel(controller.db)
		shift, err := shiftModel.FindShiftByID(int64Id)
		if err != nil || id == "" {
			session, _ := config.Store.Get(request, config.SESSION_ID)
			session.AddFlash("Gagal mendapatkan data shift! " + err.Error(), "error")
			session.Save(request, httpWriter)
		
			http.Redirect(httpWriter, request, "/shift", http.StatusSeeOther)
		}
		
		data["shift"] = shift
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	request.ParseForm()

	shift := entities.Shift{
		Id:        int64Id,
		Name:      request.Form.Get("name"),
		StartTime: request.Form.Get("start_time"),
		EndTime:   request.Form.Get("end_time"),
	}

	errorMessages := helpers.NewValidation().Struct(shift)

	if errorMessages != nil {
		data["validation"] = errorMessages
		data["shift"] = shift
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	shiftModel := models.NewShiftModel(controller.db)
	err := shiftModel.EditShift(shift)
	if err != nil {
		data["error"] = "Edit data gagal: " + err.Error()
	} else {
		updatedShift, errFind := shiftModel.FindShiftByID(int64Id)
		if errFind != nil {
			data["error"] = "Data berhasil diubah, tapi gagal menampilkan data terbaru: " + errFind.Error()
		} else {
			data["shift"] = updatedShift
			data["success"] = "Berhasil mengubah data shift."
		}
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func (controller *ShiftController) DeleteShift(httpWriter http.ResponseWriter, request *http.Request) {

	id := request.URL.Query().Get("id")
	int64Id, _ := strconv.ParseInt(id, 10, 64)

	if id == "" {
		http.Error(httpWriter, "ID tidak ditemukan", http.StatusBadRequest)
		return
	}

	shiftModel := models.NewShiftModel(controller.db)
	err := shiftModel.SoftDeleteShift(int64Id)
	if err != nil {
		http.Error(httpWriter, "Gagal menghapus data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(httpWriter, request, "/shift", http.StatusSeeOther)
}