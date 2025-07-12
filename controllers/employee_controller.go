package controllers

import (
	"fmt"
	"hris/config"
	"hris/entities"
	"hris/helpers"
	"hris/models"
	"hris/services/sessiondata"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func FindAllEmployee(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/employee/employee.html",
	))

	if request.Method == http.MethodGet {
		var data = make(map[string]interface{})

		adminOnly := request.URL.Query().Get("admin_only") == "true"
		data["adminOnly"] = adminOnly
		employeeOnly := request.URL.Query().Get("employee_only") == "true"
		data["employeeOnly"] = employeeOnly

		employees, err := models.NewEmployeeModel().FindAllEmployee(adminOnly, employeeOnly)

		if err != nil {
			data["error"] = "Gagal mengambil data karyawan: " + err.Error()
			log.Println("error find all employee: ", err.Error())
		} else {
			data["employees"] = employees
		}

		session, _ := config.Store.Get(request, config.SESSION_ID)
		if flashes := session.Flashes("success"); len(flashes) > 0 {
			data["success"] = flashes[0]
			session.Save(request, httpWriter)
		}
		if flashes := session.Flashes("error"); len(flashes) > 0 {
			data["error"] = flashes[0]
			session.Save(request, httpWriter)
		}
		sessionNIK := session.Values["nik"].(string)
		errSession := sessiondata.SetUserSessionData(request, data)
		if errSession != nil {
			log.Println("SetUserSessionData error:", errSession.Error())
		}
		data["sessionNIK"] = sessionNIK

		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}
}


func AddEmployee(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/employee/add-employee.html",
	))

	var data = make(map[string]interface{})
	session, _ := config.Store.Get(request, config.SESSION_ID)
	errSession := sessiondata.SetUserSessionData(request, data)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	if request.Method == http.MethodGet {
		data["employee"] = entities.Employee{}
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	request.ParseForm()
	employee := entities.Employee{
		UUID		: uuid.New().String(),
		Name		: request.Form.Get("name"),
		Email		: request.Form.Get("email"),
		Phone		: request.Form.Get("phone"),
		Address		: request.Form.Get("address"),
		NIK			: request.Form.Get("nik"),
		Gender		: request.Form.Get("gender"),
		BirthDate	: request.Form.Get("birth_date"),
		IsAdmin		: request.Form.Get("is_admin") != "",
	}

	hashPassword, _ := bcrypt.GenerateFromPassword([]byte("12345"), bcrypt.DefaultCost)
	employee.Password = string(hashPassword)

	errorMessages := helpers.NewValidation().Struct(employee)
	if errorMessages != nil {
		data["validation"] = errorMessages
		data["employee"] = employee
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	err := models.NewEmployeeModel().AddEmployee(employee)
	if err != nil {
		data["error"] = "Registrasi gagal: " + err.Error()
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
	} else {
		session.AddFlash("Registrasi berhasil, silahkan arahkan karyawan untuk login menggunakan email dan password default \"12345\".", "success")
		session.Save(request, httpWriter)
		http.Redirect(httpWriter, request, "/employee", http.StatusSeeOther)
	}

}

func DetailEmployee(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/employee/detail-employee.html",
	))

	uuid := request.URL.Query().Get("uuid")
	
	if request.URL.Query().Get("uuid") == "" {
		http.Error(httpWriter, "UUID tidak ditemukan", http.StatusBadRequest)
		return
	}
	var data = make(map[string]interface{})

	errSession := sessiondata.SetUserSessionData(request, data)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	employee, err := models.NewEmployeeModel().FindEmployeeByUUID(uuid)
	if err != nil {
		data["error"] = "Gagal menampilkan profile employee" + err.Error()
	}
	if employee.Photo.Valid && employee.Photo.String != "" {
		data["employeePhoto"] = "/images/user_photo/" + employee.Photo.String
	} else {
		data["employeePhoto"] = "/images/user_default.png"
	}
	data["employee"] = employee

	currentDate := time.Now()
	var months []string
	for i := 0; i < 5; i++ {
		previousMonth := currentDate.AddDate(0, -i, 0)
		months = append(months, previousMonth.Format("January 2006"))
	}
	data["months"] = months

	selectedAttendanceMonth := request.URL.Query().Get("month_attendance")
	if selectedAttendanceMonth == "" {
		selectedAttendanceMonth = currentDate.Format("January 2006")
	}
	data["selectedAttendanceMonth"] = selectedAttendanceMonth

	todayAttendance := request.URL.Query().Get("today_attendance") == "true"
	data["todayAttendance"] = todayAttendance
	attendedList, err := models.NewAttendanceModel().GetAttendanceList(employee.NIK, selectedAttendanceMonth, todayAttendance)
	if err != nil {
		data["errorList"] = "Error saat menampilkan list kehadiran: " + err.Error()
	}
	data["attendances"] = attendedList

	totalAttendanceAll, totalAttendanceThisMonth, err := models.NewAttendanceModel().GetAttendanceCounts(employee.NIK, selectedAttendanceMonth)
	if err != nil {
		fmt.Println(err)
	}
	data["totalAttendanceAll"] = totalAttendanceAll
	data["totalAttendanceThisMonth"] = totalAttendanceThisMonth


	selectedLeaveMonth := request.URL.Query().Get("month_leave")
	if selectedLeaveMonth == "" {
		selectedLeaveMonth = currentDate.Format("January 2006")
	}
	data["selectedLeaveMonth"] = selectedLeaveMonth

	todayLeave := request.URL.Query().Get("today_leave") == "true"
	data["todayLeave"] = todayLeave
	leaveList, err := models.NewLeaveModel().GetLeaveList(employee.NIK, selectedLeaveMonth, todayLeave)
	if err != nil {
		data["errorList"] = "Error saat menampilkan list kehadiran: " + err.Error()
	}
	data["leaves"] = leaveList

	totalLeaveAll, totalLeaveThisMonth, err := models.NewLeaveModel().GetLeaveCounts(employee.NIK, selectedLeaveMonth)
	if err != nil {
		fmt.Println(err)
	}
	data["totalLeaveAll"] = totalLeaveAll
	data["totalLeaveThisMonth"] = totalLeaveThisMonth

	slip, errSlip := models.NewSalaryModel().GetSalarySlipsByNIK(employee.NIK)
	if errSlip != nil {
		data["error"] = "Gagal mendapatkan slip gaji: " + errSlip.Error()
	} else {
		data["salarySlips"] = slip
	}


	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func EditEmployee(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/employee/edit-employee.html",
	))
	uuid := request.URL.Query().Get("uuid")
	var data = make(map[string]interface{})
	errSession := sessiondata.SetUserSessionData(request, data)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	if request.Method == http.MethodGet {
		if request.URL.Query().Get("uuid") == "" {
			http.Error(httpWriter, "UUID tidak ditemukan", http.StatusBadRequest)
			return
		}

		employee, err := models.NewEmployeeModel().FindEmployeeByUUID(uuid)
		if err != nil {
			data["error"] = "Gagal mengambil data karyawan: " + err.Error()
			return
		}
		data["employee"] = employee
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	request.ParseForm()
	employee := entities.EditEmployee{
		UUID		: uuid,
		Name		: request.Form.Get("name"),
		Email		: request.Form.Get("email"),
		Phone		: request.Form.Get("phone"),
		Address		: request.Form.Get("address"),
		NIK			: request.Form.Get("nik"),
		Gender		: request.Form.Get("gender"),
		BirthDate	: request.Form.Get("birth_date"),
		IsAdmin		: request.Form.Get("is_admin") != "",
	}

	errorMessages := helpers.NewValidation().Struct(employee)
	if errorMessages != nil {
		data["validation"] = errorMessages
		data["employee"] = employee
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}
	

	err := models.NewEmployeeModel().EditEmployee(employee)
	if err != nil {
		data["error"] = "Gagal mengubah data karyawan: " + err.Error()
	} else {
		updatedEmployee, errFind := models.NewEmployeeModel().FindEmployeeByUUID(uuid)
		if errFind != nil {
			data["error"] = "Data berhasil diubah, tapi gagal menampilkan data terbaru: " + errFind.Error()
		} else {
			data["employee"] = updatedEmployee
			data["success"] = "Berhasil mengubah data karyawan."
		}
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func DeletedEmployee(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/employee/deleted-employee.html",
	))

	if request.Method == http.MethodGet {
		var data = make(map[string]interface{})

		adminOnly := request.URL.Query().Get("admin_only") == "true"
		data["adminOnly"] = adminOnly
		employeeOnly := request.URL.Query().Get("employee_only") == "true"
		data["employeeOnly"] = employeeOnly

		deletedemployees, err := models.NewEmployeeModel().FindAllDeletedEmployee(adminOnly, employeeOnly)

		if err != nil {
			data["error"] = "Gagal mengambil data karyawan: " + err.Error()
			log.Println("error find all employee: ", err.Error())
		} else {
			data["deletedemployees"] = deletedemployees
		}

		session, _ := config.Store.Get(request, config.SESSION_ID)
		if flashes := session.Flashes("success"); len(flashes) > 0 {
			data["success"] = flashes[0]
			session.Save(request, httpWriter)
		}
		sessionNIK := session.Values["nik"].(string)
		errSession := sessiondata.SetUserSessionData(request, data)
		if errSession != nil {
			log.Println("SetUserSessionData error:", errSession.Error())
		}
		data["sessionNIK"] = sessionNIK

		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}
}

func SoftDeleteEmployee(httpWriter http.ResponseWriter, request *http.Request) {
	uuid := request.URL.Query().Get("uuid")
	if uuid == "" {
		http.Error(httpWriter, "UUID tidak ditemukan", http.StatusBadRequest)
		return
	}

	err := models.NewEmployeeModel().SoftDeleteEmployee(uuid)

	if err != nil {
		http.Error(httpWriter, "Gagal Menghapus data", http.StatusBadRequest)
		return
	} else {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Berhasil menghapus karyawan!", "success")
		session.Save(request, httpWriter)
	}
	http.Redirect(httpWriter, request, "/employee", http.StatusSeeOther)

}

func RestoreEmployee(httpWriter http.ResponseWriter, request *http.Request) {
	uuid := request.URL.Query().Get("uuid")
	if uuid == "" {
		http.Error(httpWriter, "UUID tidak ditemukan", http.StatusBadRequest)
		return
	}

	err := models.NewEmployeeModel().RestoreEmployee(uuid)

	if err != nil {
		fmt.Println(err)
		http.Error(httpWriter, "Gagal Mengembalikan data", http.StatusBadRequest)
		return
	} else {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Berhasil mengembalikan karyawan!", "success")
		session.Save(request, httpWriter)
	}
	http.Redirect(httpWriter, request, "/employee", http.StatusSeeOther)

}

func DeleteEmployee(httpWriter http.ResponseWriter, request *http.Request) {
	uuid := request.URL.Query().Get("uuid")
	if uuid == "" {
		http.Error(httpWriter, "UUID tidak ditemukan", http.StatusBadRequest)
		return
	}

	oldPhoto, errGetPhoto := models.NewEmployeeModel().GetPhotoByUUID(uuid)
	fmt.Println(oldPhoto)
	if errGetPhoto != nil {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Gagal mendapatkan foto karyawan!" + errGetPhoto.Error(), "error")
		session.Save(request, httpWriter)
	}

	// Hapus foto lama jika ada
	if oldPhoto.Valid && oldPhoto.String != "" {
		oldPath := filepath.Join("public/images/user_photo", oldPhoto.String)
		fmt.Println(oldPath)
		if errDelPhoto := os.Remove(oldPath); errDelPhoto != nil && !os.IsNotExist(errDelPhoto) {
			session, _ := config.Store.Get(request, config.SESSION_ID)
			session.AddFlash("Gagal menghapus foto karyawan!" + errDelPhoto.Error(), "error")
			session.Save(request, httpWriter)
		}
	}

	errDelEmployee := models.NewEmployeeModel().DeleteEmployee(uuid)
	if errDelEmployee != nil {
		http.Error(httpWriter, "Gagal Menghapus data", http.StatusBadRequest)
		return
	} else {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Berhasil menghapus karyawan!", "success")
		session.Save(request, httpWriter)
	}
	http.Redirect(httpWriter, request, "/employee", http.StatusSeeOther)
}