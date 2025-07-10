package controllers

import (
	"hris/config"
	"hris/models"
	"hris/services/sessiondata"
	"html/template"
	"log"
	"net/http"
	"time"
)

func Home(httpWriter http.ResponseWriter, request *http.Request) {
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
	if flashes := session.Flashes("success"); len(flashes) > 0 {
		data["success"] = flashes[0]
		session.Save(request, httpWriter)
	}
	errSession := sessiondata.SetUserSessionData(request, data)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	newss, err := models.NewNewsModel().FindAllNews()
	if err != nil {
		data["error"] = "Terdapat kesahalan saat menampilkan data news " + err.Error()
		log.Println("error :", err.Error())
	} else {
		data["news"] = newss
	}

	now := time.Now()
	attendanceCount, _ := models.NewAttendanceModel().CountAttendanceThisMonth(sessionNIK, now.Month(), now.Year())
	leaveDays, _ := models.NewLeaveModel().CountLeaveDaysThisMonth(sessionNIK, now.Month(), now.Year())
	data["attendanceCount"] = attendanceCount
	data["leaveDays"] = leaveDays

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func HomeAdmin(httpWriter http.ResponseWriter, request *http.Request) {
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

	session, _ := config.Store.Get(request, config.SESSION_ID)
	if flashes := session.Flashes("success"); len(flashes) > 0 {
		data["success"] = flashes[0]
		session.Save(request, httpWriter)
	}
	errSession := sessiondata.SetUserSessionData(request, data)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	countAllActiveEmployee, errCAAEmployee := models.NewEmployeeModel().CountAllActiveEmployee()
	if errCAAEmployee != nil {
		data["errCAAEmployee"] = "Gagal Mengambil Total Karyawan Active" + errCAAEmployee.Error()
	} else {
		data["totalEmployee"] = countAllActiveEmployee
	}

	countAllLeave, errCALeave := models.NewLeaveModel().CountAllLeave()
	if errCALeave != nil {
		data["errCALeave"] = "Gagal Mengambil Total Pengajuan Cuti" + errCALeave.Error()
	} else {
		data["totalLeave"] = countAllLeave
	}

	countAllAttendance, errCAAttendance := models.NewAttendanceModel().CountAllAttendance()
	if errCAAttendance != nil {
		data["errCAAttendance"] = "Gagal Mengambil Total Absen Karyawan" + errCAAttendance.Error()
	} else {
		data["totalAttendance"] = countAllAttendance
	}

	countAllNews, errCANews := models.NewNewsModel().CountAllNews()
	if errCANews != nil {
		data["errCANews"] = "Gagal Mengambil Total Berita" + errCANews.Error()
	} else {
		data["totalNews"] = countAllNews
	}

	newss, err := models.NewNewsModel().FindAllNews()
	if err != nil {
		data["error"] = "Terdapat kesahalan saat menampilkan data news " + err.Error()
		log.Println("error :", err.Error())
	} else {
		data["news"] = newss
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}