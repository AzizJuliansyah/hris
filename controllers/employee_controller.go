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
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type EmployeeController struct {
	db *sql.DB
}

func NewEmployeeController(db *sql.DB) *EmployeeController {
	return &EmployeeController{db: db}
}



func (controller *EmployeeController) FindAllEmployee(httpWriter http.ResponseWriter, request *http.Request) {
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

		employeeModel := models.NewEmployeeModel(controller.db)
		employees, err := employeeModel.FindAllEmployee(adminOnly, employeeOnly)

		if err != nil {
			data["error"] = "Gagal mengambil data karyawan: " + err.Error()
			log.Println("error find all employee: ", err.Error())
		} else {
			data["employees"] = employees
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


func (controller *EmployeeController) AddEmployee(httpWriter http.ResponseWriter, request *http.Request) {
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
	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
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

	employeeModel := models.NewEmployeeModel(controller.db)
	err := employeeModel.AddEmployee(employee)
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

func (controller *EmployeeController) DetailEmployee(httpWriter http.ResponseWriter, request *http.Request) {
	funcMap := template.FuncMap{
		"formatIDR": humanizeIDR,
		"toInt64": toInt64,
	}

	templateLayout := template.Must(template.New("base").Funcs(funcMap).ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/employee/detail-employee.html",
	))

	uuid := request.URL.Query().Get("uuid")
	if uuid == "" {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Gagal, UUID kosong!", "error")
		session.Save(request, httpWriter)

		http.Redirect(httpWriter, request, "/employee", http.StatusSeeOther)
	}
	var data = make(map[string]interface{})

	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	employeeModel := models.NewEmployeeModel(controller.db)
	employee, err := employeeModel.FindEmployeeByUUID(uuid)
	if err != nil {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Gagal mendapatkan data karyawan!" + err.Error(), "error")
		session.Save(request, httpWriter)

		http.Redirect(httpWriter, request, "/employee", http.StatusSeeOther)
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

	attendanceModel := models.NewAttendanceModel(controller.db)
	attendedList, err := attendanceModel.GetAttendanceList(employee.NIK, selectedAttendanceMonth, todayAttendance)
	if err != nil {
		data["errorList"] = "Error saat menampilkan list kehadiran: " + err.Error()
	}
	data["attendances"] = attendedList

	totalAttendanceAll, totalAttendanceThisMonth, err := attendanceModel.GetAttendanceCounts(employee.NIK, selectedAttendanceMonth)
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

	leaveModel := models.NewLeaveModel(controller.db)
	leaveList, err := leaveModel.GetLeaveList(employee.NIK, selectedLeaveMonth, todayLeave)
	if err != nil {
		data["errorList"] = "Error saat menampilkan list pengajuan cuti: " + err.Error()
	}
	data["leaves"] = leaveList

	totalLeaveAll, totalLeaveThisMonth, err := leaveModel.GetLeaveCounts(employee.NIK, selectedLeaveMonth)
	if err != nil {
		fmt.Println(err)
	}
	data["totalLeaveAll"] = totalLeaveAll
	data["totalLeaveThisMonth"] = totalLeaveThisMonth

	salaryModel := models.NewSalaryModel(controller.db)
	slip, errSlip := salaryModel.GetSalarySlipsByNIK(employee.NIK)
	if errSlip != nil {
		data["error"] = "Gagal mendapatkan slip gaji: " + errSlip.Error()
	} else {
		data["salarySlips"] = slip
	}

	wages, errWages := salaryModel.GetEmployeeWagesByNIK(employee.NIK)
	if errWages != nil {
		data["error"] = "Gagal mengambil data gaji" + errWages.Error()
	} else {
		data["wages"] = wages
	}


	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func (controller *EmployeeController) EditEmployee(httpWriter http.ResponseWriter, request *http.Request) {
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
	if uuid == "" {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Gagal, UUID kosong!", "error")
		session.Save(request, httpWriter)

		http.Redirect(httpWriter, request, "/employee", http.StatusSeeOther)
	}
	var data = make(map[string]interface{})
	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	if request.Method == http.MethodGet {
		employeeModel := models.NewEmployeeModel(controller.db)
		employee, err := employeeModel.FindEmployeeByUUID(uuid)
		if err != nil {
			session, _ := config.Store.Get(request, config.SESSION_ID)
			session.AddFlash("Gagal mendapatkan data karyawan!" + err.Error(), "error")
			session.Save(request, httpWriter)

			http.Redirect(httpWriter, request, "/employee", http.StatusSeeOther)
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
	
	employeeModel := models.NewEmployeeModel(controller.db)
	err := employeeModel.EditEmployee(employee)
	if err != nil {
		data["error"] = "Gagal mengubah data karyawan: " + err.Error()
	} else {
		updatedEmployee, errFind := employeeModel.FindEmployeeByUUID(uuid)
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

func (controller *EmployeeController) DeletedEmployee(httpWriter http.ResponseWriter, request *http.Request) {
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

		employeeModel := models.NewEmployeeModel(controller.db)
		deletedemployees, err := employeeModel.FindAllDeletedEmployee(adminOnly, employeeOnly)

		if err != nil {
			data["error"] = "Gagal mengambil data karyawan: " + err.Error()
			log.Println("error find all employee: ", err.Error())
		} else {
			data["deletedemployees"] = deletedemployees
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

func (controller *EmployeeController) SoftDeleteEmployee(httpWriter http.ResponseWriter, request *http.Request) {
	uuid := request.URL.Query().Get("uuid")
	if uuid == "" {
		http.Error(httpWriter, "UUID tidak ditemukan", http.StatusBadRequest)
		return
	}

	employeeModel := models.NewEmployeeModel(controller.db)
	err := employeeModel.SoftDeleteEmployee(uuid)

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

func (controller *EmployeeController) RestoreEmployee(httpWriter http.ResponseWriter, request *http.Request) {
	uuid := request.URL.Query().Get("uuid")
	if uuid == "" {
		http.Error(httpWriter, "UUID tidak ditemukan", http.StatusBadRequest)
		return
	}

	employeeModel := models.NewEmployeeModel(controller.db)
	err := employeeModel.RestoreEmployee(uuid)

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

func (controller *EmployeeController) DeleteEmployee(httpWriter http.ResponseWriter, request *http.Request) {
	uuid := request.URL.Query().Get("uuid")
	if uuid == "" {
		http.Error(httpWriter, "UUID tidak ditemukan", http.StatusBadRequest)
		return
	}

	employeeModel := models.NewEmployeeModel(controller.db)
	oldPhoto, errGetPhoto := employeeModel.GetPhotoByUUID(uuid)
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

	errDelEmployee := employeeModel.DeleteEmployee(uuid)
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