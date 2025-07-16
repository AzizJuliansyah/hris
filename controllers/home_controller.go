package controllers

import (
	"database/sql"
	"fmt"
	"hris/config"
	"hris/models"
	"hris/services/sessiondata"
	"html/template"
	"log"
	"net/http"
	"time"
)

type HomeController struct {
	db *sql.DB
}

func NewHomeController(db *sql.DB) *HomeController {
	return &HomeController{db: db}
}

func (controller *HomeController) Home(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/home/home.html",
	))
	data := make(map[string]interface{})

	session, _ := config.Store.Get(request, config.SESSION_ID)
	sessionNIK := session.Values["nik"].(string)
	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	newsModel := models.NewNewsModel(controller.db)
	newss, err := newsModel.FindAllNews()
	if err != nil {
		data["error"] = "Terdapat kesahalan saat menampilkan data news " + err.Error()
		log.Println("error :", err.Error())
	} else {
		data["news"] = newss
	}

	now := time.Now()
	month := now.Format("January 2006")
	data["month"] = month

	attendanceModel := models.NewAttendanceModel(controller.db)
	totalAttendanceAll, totalAttendanceThisMonth, err := attendanceModel.GetAttendanceCounts(sessionNIK, month)
	if err != nil {
		fmt.Println(err)
	}
	data["totalAttendanceAll"] = totalAttendanceAll
	data["totalAttendanceThisMonth"] = totalAttendanceThisMonth

	leaveModel := models.NewLeaveModel(controller.db)
	totalLeaveAll, totalLeaveThisMonth, err := leaveModel.GetLeaveCounts(sessionNIK, month)
	if err != nil {
		fmt.Println(err)
	}
	data["totalLeaveAll"] = totalLeaveAll
	data["totalLeaveThisMonth"] = totalLeaveThisMonth

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func (controller *HomeController) HomeAdmin(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/home/home-admin.html",
	))
	data := make(map[string]interface{})

	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	EmployeeModel := models.NewEmployeeModel(controller.db)
	countAllActiveEmployee, errCAAEmployee := EmployeeModel.CountAllActiveEmployee()
	if errCAAEmployee != nil {
		data["errCAAEmployee"] = "Gagal Mengambil Total Karyawan Active" + errCAAEmployee.Error()
	} else {
		data["totalEmployee"] = countAllActiveEmployee
	}

	leaveModel := models.NewLeaveModel(controller.db)
	countAllLeave, errCALeave := leaveModel.CountAllLeave()
	if errCALeave != nil {
		data["errCALeave"] = "Gagal Mengambil Total Pengajuan Cuti" + errCALeave.Error()
	} else {
		data["totalLeave"] = countAllLeave
	}

	attendanceModel := models.NewAttendanceModel(controller.db)
	countAllAttendance, errCAAttendance := attendanceModel.CountAllAttendance()
	if errCAAttendance != nil {
		data["errCAAttendance"] = "Gagal Mengambil Total Absen Karyawan" + errCAAttendance.Error()
	} else {
		data["totalAttendance"] = countAllAttendance
	}

	newsModel := models.NewNewsModel(controller.db)
	countAllNews, errCANews := newsModel.CountAllNews()
	if errCANews != nil {
		data["errCANews"] = "Gagal Mengambil Total Berita" + errCANews.Error()
	} else {
		data["totalNews"] = countAllNews
	}

	newss, err := newsModel.FindAllNews()
	if err != nil {
		data["error"] = "Terdapat kesahalan saat menampilkan data news " + err.Error()
		log.Println("error :", err.Error())
	} else {
		data["news"] = newss
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}