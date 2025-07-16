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

type AttendanceController struct {
	db *sql.DB
}

func NewAttendanceController(db *sql.DB) *AttendanceController {
	return &AttendanceController{db: db}
}

func (controller *AttendanceController) SubmitAttendance(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/attendance/attendance-submit.html",
	))
	data := make(map[string]interface{})
	
	
	session, _ := config.Store.Get(request, config.SESSION_ID)
	sessionNIK := session.Values["nik"].(string)
	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	updateAttendanceStatus(controller.db, sessionNIK, data)

	officeModel := models.NewOfficeModel(controller.db)
	office, _ := officeModel.FindAllOffice()
	data["office"] = office

	shiftModel := models.NewShiftModel(controller.db)
	shift, _ := shiftModel.FindAllShift()
	data["shift"] = shift

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

	todayAttendance := request.URL.Query().Get("today_attendance") == "true"
	data["todayAttendance"] = todayAttendance

	getAttendanceList(controller.db, sessionNIK, data, selectedMonth, todayAttendance)

	if request.Method == http.MethodGet {
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	request.ParseForm()
	switch data["status"] {
	case helpers.NOT_CHECKED_IN:
		actionCheckIn(controller.db, httpWriter, request, sessionNIK, data, selectedMonth, todayAttendance)
	case helpers.CHECKED_IN:
		actionCheckOut(controller.db, httpWriter, request, sessionNIK, data, selectedMonth, todayAttendance)
	}
}

func updateAttendanceStatus(db *sql.DB, sessionNik string, data map[string]interface{}) {
	attendanceModel := models.NewAttendanceModel(db)
	lastAtt := attendanceModel.GetLastAttendance(sessionNik)
	data["status"] = lastAtt
}

func getAttendanceList(db *sql.DB, sessionNik string, data map[string]interface{}, selectedMonth string, todayAttendance bool) {
	attendanceModel := models.NewAttendanceModel(db)
	attendedList, err := attendanceModel.GetAttendanceList(sessionNik, selectedMonth, todayAttendance)
	if err != nil {
		data["errorList"] = "Error saat menampilkan list kehadiran: " + err.Error()
	}

	data["attendances"] = attendedList
	data["selectedMonth"] = selectedMonth
}

func actionCheckIn(db *sql.DB, httpWriter http.ResponseWriter, request *http.Request, sessionNIK string, data map[string]interface{}, selectedMonth string, todayAttendance bool) {
	layoutTime := "2006-01-02 15:04:05"
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/attendance/attendance-submit.html",
	))
	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	photo := request.Form.Get("attendance_photo")
	latLongStr := request.FormValue("latlong")
	officeIDStr := request.FormValue("office_id")
	shiftIDStr := request.FormValue("shift_id")
	notes := request.Form.Get("notes")

	checkIn := entities.CheckIn{
		NIK: sessionNIK,
		Photo: photo,
		LatLongStr: latLongStr,
		OfficeID: officeIDStr,
		ShiftID: shiftIDStr,
		Notes: notes,
	}

	errorValidation := helpers.NewValidation().Struct(checkIn)
	if errorValidation != nil {
		data["validation"] = errorValidation
		data["attendance"] = checkIn
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	officeID, _ := strconv.ParseInt(officeIDStr, 10, 64)
	shiftID, _ := strconv.ParseInt(shiftIDStr, 10, 64)

	latLongParts := strings.Split(latLongStr, ",")
	latitude, _ := strconv.ParseFloat(strings.TrimSpace(latLongParts[0]), 64)
	longitude, _ := strconv.ParseFloat(strings.TrimSpace(latLongParts[1]), 64)

	officeModel := models.NewOfficeModel(db)
	findOffice, _ := officeModel.FindOfficeByID(officeID)

	shiftModel := models.NewShiftModel(db)
	findShift, _ := shiftModel.FindShiftByID(shiftID)

	distance := helpers.CalculateDistance(latitude, longitude, findOffice.Latitude, findOffice.Longitude)
	if distance > float64(findOffice.Radius) {
		data["error"] = "Anda berada diluar radius kantor, tidak bisa check-in"
		data["attendance"] = checkIn
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	now := time.Now()
	dateToday := now.Format("2006-01-02")
	shiftStartFull := fmt.Sprintf("%s %s", dateToday, findShift.StartTime)
	shiftEndFull := fmt.Sprintf("%s %s", dateToday, findShift.EndTime)

	loc, _ := time.LoadLocation("Asia/Jakarta")
	shiftStartTime, _ := time.ParseInLocation(layoutTime, shiftStartFull, loc)
	shiftEndTime, _ := time.ParseInLocation(layoutTime, shiftEndFull, loc)

	if now.After(shiftEndTime) {
		data["error"] = "Shift yang anda pilih sudah selesai"
		data["attendance"] = checkIn
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	isLate := now.After(shiftStartTime)
	checkIn.Time = now
	checkIn.Latitude = latitude
	checkIn.Longitude = longitude
	checkIn.IsLate = isLate

	attendanceModel := models.NewAttendanceModel(db)
	errCheckIn := attendanceModel.CheckIn(checkIn)
	if errCheckIn != nil {
		data["error"] = "Error " + errCheckIn.Error()
	} else {
		if isLate {
			data["isLate"] = "Berhasil check in, namun anda terlambat"
			} else {
				data["success"] = "Berhasil check in, selamat bekerja"
			}
		getAttendanceList(db, sessionNIK, data, selectedMonth, todayAttendance)
		updateAttendanceStatus(db, sessionNIK, data)
	}

	data["attendance"] = entities.Attendance{}
	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func actionCheckOut(db *sql.DB, httpWriter http.ResponseWriter, request *http.Request, sessionNIK string, data map[string]interface{}, selectedMonth string, todayAttendance bool) {
	layoutTime := "2006-01-02 15:04:05"
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/attendance/attendance-submit.html",
	))
	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	photo := request.Form.Get("attendance_photo")
	latLongStr := request.FormValue("latlong")
	notes := request.Form.Get("notes")

	checkOut := entities.CheckOut{
		NIK: sessionNIK,
		Photo: photo,
		LatLongStr: latLongStr,
		Notes: notes,
	}

	errorValidation := helpers.NewValidation().Struct(checkOut)
	if errorValidation != nil {
		data["validation"] = errorValidation
		data["attendance"] = checkOut
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	latLongParts := strings.Split(latLongStr, ",")
	latitude, _ := strconv.ParseFloat(strings.TrimSpace(latLongParts[0]), 64)
	longitude, _ := strconv.ParseFloat(strings.TrimSpace(latLongParts[1]), 64)

	attendanceModel := models.NewAttendanceModel(db)
	officeID, shiftID, _ := attendanceModel.GetLatestOfficeAndShift(sessionNIK)

	officeModel := models.NewOfficeModel(db)
	findOffice, _ := officeModel.FindOfficeByID(officeID)
	distance := helpers.CalculateDistance(latitude, longitude, findOffice.Latitude, findOffice.Longitude)
	if distance > float64(findOffice.Radius) {
		data["error"] = "Anda berada diluar radius kantor, tidak bisa check-out"
		data["attendance"] = checkOut
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	shiftModel := models.NewShiftModel(db)
	findShift, _ := shiftModel.FindShiftByID(shiftID)
	now := time.Now()
	dateToday := now.Format("2006-01-02")
	shiftEndFull := fmt.Sprintf("%s %s", dateToday, findShift.EndTime)

	loc, _ := time.LoadLocation("Asia/Jakarta")
	shiftEndTime, _ := time.ParseInLocation(layoutTime, shiftEndFull, loc)

	isEarly := now.Before(shiftEndTime)
	checkOut.Time = now
	checkOut.Latitude = latitude
	checkOut.Longitude = longitude
	checkOut.IsEarly = isEarly

	errCheckOut := attendanceModel.CheckOut(sessionNIK, checkOut)
	if errCheckOut != nil {
		data["error"] = "Error " + errCheckOut.Error()
	} else {
		if isEarly {
			data["isEarly"] = "Berhasil checkout, namun anda pulang lebih awal"
		} else {
			data["success"] = "Berhasil checkout, selamat pulang"
		}
		getAttendanceList(db, sessionNIK, data, selectedMonth, todayAttendance)
		updateAttendanceStatus(db, sessionNIK, data)
	}

	data["attendance"] = entities.Attendance{}
	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func (controller *AttendanceController) ListAttendance(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/attendance/attendance-list.html",
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

	todayAttendance := request.URL.Query().Get("today_attendance") == "true"
	data["todayAttendance"] = todayAttendance

	attendanceModel := models.NewAttendanceModel(controller.db)
	attendaceList, err := attendanceModel.GetAttendanceList("", selectedMonth, todayAttendance)
	if err != nil {
		data["errorList"] = "Gagal menampilkan list kehadiran"
	}

	data["attendances"] = attendaceList
	data["selectedMonth"] = selectedMonth


	if request.Method == http.MethodGet {
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}
}