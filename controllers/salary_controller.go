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
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func ListSalary(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/salary/salary-list.html",
	))
	data := make(map[string]interface{})
	session, _ := config.Store.Get(request, config.SESSION_ID)	
	errSession := sessiondata.SetUserSessionData(request, data)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}
	if flashes := session.Flashes("success"); len(flashes) > 0 {
		data["success"] = flashes[0]
		session.Save(request, httpWriter)
	}

	if request.Method == http.MethodGet {
		salaries, err := models.NewSalaryModel().FindAllSalaries()
		
		if err != nil {
			data["error"] = "Failed to retrieve salary data: " + err.Error()
			log.Println("Error retrieving salaries:", err)
		} else {
			data["salaries"] = salaries
		}		

		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}
}

func DetailEmployeeSalary(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.New("base").Funcs(template.FuncMap{
		"formatNumber": func(n int64) string {
			return strconv.FormatInt(n, 10)
		},
		"formatPeriod": func(period string) string {
			t, err := time.Parse("2006-01", period)
			if err != nil {
				return period // fallback jika gagal parse
			}
			return t.Format("January 2006")
		},
	})
	templateLayout = template.Must(templateLayout.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/salary/detail-salary.html",
	))
	

	data := make(map[string]interface{})
	session, _ := config.Store.Get(request, config.SESSION_ID)
	errSession := sessiondata.SetUserSessionData(request, data)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}
	if flashes := session.Flashes("success"); len(flashes) > 0 {
		data["success"] = flashes[0]
		session.Save(request, httpWriter)
	}

	id := request.URL.Query().Get("id")
	int64Id, _ := strconv.ParseInt(id, 10, 64)
	if id == "" {
		http.Error(httpWriter, "ID tidak ditemukan", http.StatusBadRequest)
		return
	}

	salary, err := models.NewSalaryModel().FindSalaryByID(int64Id)
	if err != nil {
		data["error"] = "Failed to retrieve salary data: " + err.Error()
		log.Println("Error retrieving salary:", err)
	} else {
		data["salary"] = salary
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

	nik := salary.NIK
	parsedMonth, _ := time.Parse("January 2006", selectedMonth)
	month := parsedMonth.Month()
	year := parsedMonth.Year()

	daysPresent, _ := models.NewAttendanceModel().CountDaysPresent(nik, month, year)
	leaves, errLeave := models.NewLeaveModel().FindLeavesByNIK(nik)
	if errLeave != nil {
		log.Println("Error get leaves:", errLeave)
	}
	daysLeave := 0
	for _, leave := range leaves {
		daysLeave += countLeaveDaysInMonth(leave.LeaveDate, year, month)
	}

	var (
		monthly   int64
		daily     int64
		meal      int64
		transport int64
		total     int64
	)
	
	if salary.Monthly_Wages.Valid {
		monthly = parseInt64(salary.Monthly_Wages.String)
		total += monthly
	}
	if salary.Daily_Wages.Valid {
		daily = parseInt64(salary.Daily_Wages.String) * int64(daysPresent)
		total += daily
	}
	if salary.Meal_Allowance.Valid {
		meal = parseInt64(salary.Meal_Allowance.String) * int64(daysPresent)
		total += meal
	}
	if salary.Transport_Allowance.Valid {
		transport = parseInt64(salary.Transport_Allowance.String) * int64(daysPresent)
		total += transport
	}
	
	data["monthlyTotal"] = monthly
	data["dailyTotal"] = daily
	data["mealTotal"] = meal
	data["transportTotal"] = transport
	data["salaryTotal"] = total
	data["daysPresent"] = daysPresent
	data["daysLeave"] = daysLeave
	data["salaryId"] = salary.Id
	data["employeeName"] = salary.EmployeeName

	slip, errSlip := models.NewSalaryModel().GetSalarySlipsByNIK(nik)
	if errSlip != nil {
		data["error"] = "Gagal mendapatkan slip gaji: " + errSlip.Error()
	} else {
		data["salarySlips"] = slip
	}


	if request.Method == http.MethodPost {
		exists, err := models.NewSalaryModel().IsSlipExist(nik, year, int(month))
		if err != nil {
			http.Error(httpWriter, "Gagal mengecek slip: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if exists {
			data["error"] = "Slip gaji bulan ini sudah diterbitkan."
			templateLayout.ExecuteTemplate(httpWriter, "base", data)
			return
		}
	
		slip := entities.SalarySlip{
			NIK:                     nik,
			Period:                  fmt.Sprintf("%04d-%02d", year, int(month)),
			Monthly_Wages:           salary.Monthly_Wages,
			Daily_Wages:             salary.Daily_Wages,
			Meal_Allowance:          salary.Meal_Allowance,
			Transport_Allowance:     salary.Transport_Allowance,
			Total_Working_Days:      daysPresent,
			Total_Leave_Days:        daysLeave,
			Monthly_Wages_Received:  sql.NullInt64{Int64: monthly, Valid: salary.Monthly_Wages.Valid},
			Daily_Wages_Received:    sql.NullInt64{Int64: daily, Valid: salary.Daily_Wages.Valid},
			Meal_Allowance_Received: sql.NullInt64{Int64: meal, Valid: salary.Meal_Allowance.Valid},
			Transport_Allowance_Received: sql.NullInt64{Int64: transport, Valid: salary.Transport_Allowance.Valid,},
			Salary_Total: 		   sql.NullInt64{Int64: total, Valid: true},
		}
		err = models.NewSalaryModel().CreateSalarySlip(slip)
		if err != nil {
			data["error"] = "Gagal menerbitkan slip gaji: " + err.Error()
		} else {
			session.AddFlash("Slip gaji berhasil diterbitkan.", "success")
			session.Save(request, httpWriter)
			http.Redirect(httpWriter, request, "/salary/detail-salary?id="+id, http.StatusSeeOther)

		}
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func SlipListEmployeeSalary(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/salary/slip-list.html",
	))

	data := make(map[string]interface{})
	session, _ := config.Store.Get(request, config.SESSION_ID)
	sessionNIK := session.Values["nik"].(string)
	errSession := sessiondata.SetUserSessionData(request, data)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	slip, errSlip := models.NewSalaryModel().GetSalarySlipsByNIK(sessionNIK)
	if errSlip != nil {
		data["error"] = "Gagal mendapatkan slip gaji: " + errSlip.Error()
	} else {
		data["salarySlips"] = slip
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}


func humanizeComma(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var result []string
	for len(s) > 3 {
		result = append([]string{s[len(s)-3:]}, result...)
		s = s[:len(s)-3]
	}
	if s != "" {
		result = append([]string{s}, result...)
	}
	return strings.Join(result, ",")
}


func DownloadEmployeeSlip(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := "views/static/salary/slip-pdf.html"

	id := request.URL.Query().Get("id")
	data := make(map[string]interface{})

	if id == "" {
		http.Error(httpWriter, "ID tidak ditemukan", http.StatusBadRequest)
		return
	}
	int64Id, _ := strconv.ParseInt(id, 10, 64)

	slip, err := models.NewSalaryModel().GetSalarySlipByID(int64Id)
	if err != nil {
		http.Error(httpWriter, "Gagal mengambil data slip: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data["slip"] = slip

	funcMap := template.FuncMap{
		"formatNumber": func(n int64) string {
			return humanizeComma(n)
		},
	}

	tmpl, err := template.New(filepath.Base(templateLayout)).Funcs(funcMap).ParseFiles(templateLayout)
	if err != nil {
		http.Error(httpWriter, "Gagal memuat template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(httpWriter, data)
	if err != nil {
		http.Error(httpWriter, "Gagal render template: "+err.Error(), http.StatusInternalServerError)
	}
}


func countLeaveDaysInMonth(leaveDates string, year int, month time.Month) int {
	count := 0
	dates := strings.Split(leaveDates, ",")
	for _, d := range dates {
		t, err := time.Parse("2006-01-02", d)
		if err == nil && t.Year() == year && t.Month() == month {
			count++
		}
	}
	return count
}
func parseInt64(value string) int64 {
	if value == "" {
		return 0
	}
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Println("Error converting string to int64:", err)
		return 0
	}
	return intValue
}

func InputEmployeeSalary(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/salary/input-salary.html",
	))

	data := make(map[string]interface{})
	errSession := sessiondata.SetUserSessionData(request, data)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	employees, err := models.NewSalaryModel().GetEmployeeNameandNIK()
	if err != nil {
		log.Println("Error Getting Employee NIK and Name", err)
		return
	}
	data["employees"] = employees

	if request.Method == http.MethodPost {
		request.ParseForm()

		salary := entities.EmployeeSalary{
			NIK:                 request.Form.Get("nik"),
			Monthly_Wages:       sql.NullString{String: request.Form.Get("monthly_wages"), Valid: true},
			Daily_Wages:         sql.NullString{String: request.Form.Get("daily_wages"), Valid: true},
			Meal_Allowance:      sql.NullString{String: request.Form.Get("meal_allowance"), Valid: true},
			Transport_Allowance: sql.NullString{String: request.Form.Get("transport_allowance"), Valid: true},
		}

		errorMessages := helpers.NewValidation().Struct(salary)
		if errorMessages != nil {
			data["validation"] = errorMessages
			data["salary"] = salary
			data["currentPath"] = request.URL.Path
			templateLayout.ExecuteTemplate(httpWriter, "base", data)
			return
		}

		nikInput := request.Form.Get("nik")
		exists, errNik := models.NewSalaryModel().IsEmployeeExistByNIK(nikInput)
		if errNik != nil {
			log.Println("Error cek NIK:", errNik)
			data["error"] = "Terjadi kesalahan saat validasi NIK"
			data["validation"] = map[string]string{"NIK": "Terjadi kesalahan saat validasi NIK"}
			data["salary"] = salary
			templateLayout.ExecuteTemplate(httpWriter, "base", data)
			return
		}
		if !exists {
			data["error"] = "NIK tidak ditemukan dalam data karyawan"
			data["validation"] = map[string]string{"NIK": "NIK tidak ditemukan dalam data karyawan"}
			data["salary"] = salary
			templateLayout.ExecuteTemplate(httpWriter, "base", data)
			return
		}

		err := models.NewSalaryModel().InputEmployeeSalary(salary)
		if err != nil {
			log.Println("Error inputting employee salary:", err)
			http.Error(httpWriter, "Failed to input salary", http.StatusInternalServerError)
			return
		}

		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Salary input successful", "success")
		session.Save(request, httpWriter)

		http.Redirect(httpWriter, request, "/salary-list", http.StatusSeeOther)
		return
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}


func EditEmployeeSalary(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/salary/edit-salary.html",
	))

	data := make(map[string]interface{})
	errSession := sessiondata.SetUserSessionData(request, data)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	id := request.URL.Query().Get("id")
	int64Id, _ := strconv.ParseInt(id, 10, 64)
	if id == "" {
		http.Error(httpWriter, "ID tidak ditemukan", http.StatusBadRequest)
		return
	}

	if request.Method == http.MethodGet {
		salary, err := models.NewSalaryModel().FindSalaryByID(int64Id)
		if err != nil {
			data["error"] = "Failed to retrieve salary data: " + err.Error()
			log.Println("Error retrieving salary:", err)
		} else {
			data["salary"] = salary
		}

		session, _ := config.Store.Get(request, config.SESSION_ID)
		if flashes := session.Flashes("success"); len(flashes) > 0 {
			data["success"] = flashes[0]
			session.Save(request, httpWriter)
		}

		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	if request.Method == http.MethodPost {
		request.ParseForm()

		salary := entities.EditEmployeeSalary{
			Id: 				 int64Id,		
			Monthly_Wages:       sql.NullString{String: request.Form.Get("monthly_wages"), Valid: true},
			Daily_Wages:         sql.NullString{String: request.Form.Get("daily_wages"), Valid: true},
			Meal_Allowance:      sql.NullString{String: request.Form.Get("meal_allowance"), Valid: true},
			Transport_Allowance: sql.NullString{String: request.Form.Get("transport_allowance"), Valid: true},
		}

		errorMessages := helpers.NewValidation().Struct(salary)
		if errorMessages != nil {
			data["validation"] = errorMessages
			data["salary"] = salary
			data["currentPath"] = request.URL.Path
			templateLayout.ExecuteTemplate(httpWriter, "base", data)
			return
		}

		err := models.NewSalaryModel().EditEmployeeSalary(salary)
		if err != nil {
			log.Println("Error editting employee salary:", err)
			http.Error(httpWriter, "Failed to edit salary", http.StatusInternalServerError)
			return
		} else {
			salary, errFind := models.NewSalaryModel().FindSalaryByID(int64Id)
			if errFind != nil {
				data["error"] = "Data berhasil diubah, tapi gagal menampilkan data terbaru: " + errFind.Error()
			} else {
				data["salary"] = salary
				data["success"] = "Berhasil mengubah data gaji."
			}
		}
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func DeleteEmployeeSalary(httpWriter http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("id")
	int64Id, _ := strconv.ParseInt(id, 10, 64)
	if id == "" {
		http.Error(httpWriter, "ID tidak ditemukan", http.StatusBadRequest)
		return
	}

	err := models.NewSalaryModel().SoftDeleteSalary(int64Id)
	if err != nil {
		log.Println("Error deleting employee salary:", err)
		http.Error(httpWriter, "Failed to delete salary", http.StatusInternalServerError)
		return
	}

	session, _ := config.Store.Get(request, config.SESSION_ID)
	session.AddFlash("Salary deleted successfully", "success")
	session.Save(request, httpWriter)

	http.Redirect(httpWriter, request, "/salary-list", http.StatusSeeOther)
}

func DeleteEmployeeSlip(httpWriter http.ResponseWriter, request *http.Request) {
	slip_id := request.URL.Query().Get("slip_id")
	salary_id := request.URL.Query().Get("id")
	int64Id, _ := strconv.ParseInt(slip_id, 10, 64)
	if slip_id == "" {
		http.Error(httpWriter, "ID tidak ditemukan", http.StatusBadRequest)
		return
	}

	err := models.NewSalaryModel().DeleteSalarySlip(int64Id)
	if err != nil {
		log.Println("Error deleting employee slip:", err)
		http.Error(httpWriter, "Failed to delete slip", http.StatusInternalServerError)
		return
	}

	session, _ := config.Store.Get(request, config.SESSION_ID)
	session.AddFlash("Slip deleted successfully", "success")
	session.Save(request, httpWriter)

	http.Redirect(httpWriter, request, "/salary/detail-salary?id="+salary_id, http.StatusSeeOther)
}