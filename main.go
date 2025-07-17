package main

import (
	"fmt"
	"hris/config"
	"hris/controllers"
	"log"
	"net/http"
)

func main() {
	db, err := config.DBConnection()
	if err != nil {
		log.Fatal("Gagal membuat koneksi ke database:", err)
	} else {
		fmt.Println("Berhasil membuat koneksi ke database")
	}

	
	// baca file untuk resource yg ada di static (css, js, img)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("public/assets"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("public/images"))))

	
	// auth routes
	authController := controllers.NewAuthController(db)
	http.HandleFunc("/login", config.GuestOnly(authController.Login))
	http.HandleFunc("/logout", controllers.Logout)

	// user routes
	userController := controllers.NewUserController(db)
	http.HandleFunc("/pages-profile", config.AuthOnly(userController.Profile))

	// home routes
	homeController := controllers.NewHomeController(db)
	http.HandleFunc("/home", config.EmployeeOnly(homeController.Home))
	http.HandleFunc("/home-admin", config.AdminOnly(homeController.HomeAdmin))

	// news routes
	newsController := controllers.NewNewsController(db)
	http.HandleFunc("/news", config.AdminOnly(newsController.ListNews))
	http.HandleFunc("/news/add-news", config.AdminOnly(newsController.AddNews))
	http.HandleFunc("/news/edit-news", config.AdminOnly(newsController.EditNews))
	http.HandleFunc("/news/delete-news", config.AdminOnly(newsController.DeleteNews))
	
	// employee routes
	employeeController := controllers.NewEmployeeController(db)
	http.HandleFunc("/employee", config.AdminOnly(employeeController.FindAllEmployee))
	http.HandleFunc("/employee/add-employee", config.AdminOnly(employeeController.AddEmployee))
	http.HandleFunc("/employee/detail-employee", config.AdminOnly(employeeController.DetailEmployee))
	http.HandleFunc("/employee/edit-employee", config.AdminOnly(employeeController.EditEmployee))
	http.HandleFunc("/employee/soft-delete-employee", config.AdminOnly(employeeController.SoftDeleteEmployee))
	http.HandleFunc("/employee/deleted-employee", config.AdminOnly(employeeController.DeletedEmployee))
	http.HandleFunc("/employee/restore-employee", config.AdminOnly(employeeController.RestoreEmployee))
	http.HandleFunc("/employee/delete-employee", config.AdminOnly(employeeController.DeleteEmployee))

	// office routes
	officeController := controllers.NewOfficeController(db)
	http.HandleFunc("/office", config.AdminOnly(officeController.Office))                    
	http.HandleFunc("/office/add-office", config.AdminOnly(officeController.AddOffice))      
	http.HandleFunc("/office/edit-office", config.AdminOnly(officeController.EditOffice))    
	http.HandleFunc("/office/delete-office", config.AdminOnly(officeController.DeleteOffice))
	
	// shift routes
	shiftController := controllers.NewShiftController(db)
	http.HandleFunc("/shift", config.AdminOnly(shiftController.FindAllShift))                    
	http.HandleFunc("/shift/add-shift", config.AdminOnly(shiftController.AddShift))       
	http.HandleFunc("/shift/edit-shift", config.AdminOnly(shiftController.EditShift))     
	http.HandleFunc("/shift/delete-shift", config.AdminOnly(shiftController.DeleteShift))

	// attendance routes
	attendanceController := controllers.NewAttendanceController(db)
	http.HandleFunc("/attendance-submit", config.EmployeeOnly(attendanceController.SubmitAttendance))
	http.HandleFunc("/attendance-list", config.AdminOnly(attendanceController.ListAttendance))
	
	// leave routes
	leaveController := controllers.NewLeaveController(db)
	http.HandleFunc("/leave/leave-type", config.AdminOnly(leaveController.LeaveType))
	http.HandleFunc("/leave/delete-leave-type", config.AdminOnly(leaveController.DeleteLeaveType))
	http.HandleFunc("/leave-list", config.AdminOnly(leaveController.ListLeave))
	http.HandleFunc("/leave-submit", config.EmployeeOnly(leaveController.SubmitLeave))
	http.HandleFunc("/leave/approval", config.AdminOnly(leaveController.ApprovalLeave))

	// salary routes
	salaryController := controllers.NewSalaryController(db)
	http.HandleFunc("/salary-list", config.AdminOnly(salaryController.ListSalary))
	http.HandleFunc("/salary/detail-salary", config.AdminOnly(salaryController.DetailEmployeeSalary))
	http.HandleFunc("/slip-list", config.EmployeeOnly(salaryController.SlipListEmployeeSalary))
	http.HandleFunc("/salary/input-salary", config.AdminOnly(salaryController.InputEmployeeSalary))
	http.HandleFunc("/salary/edit-salary", config.AdminOnly(salaryController.EditEmployeeSalary))
	http.HandleFunc("/salary/delete-salary", config.AdminOnly(salaryController.DeleteEmployeeSalary))
	http.HandleFunc("/salary/delete-slip", config.AdminOnly(salaryController.DeleteEmployeeSlip))
	http.HandleFunc("/salary/download-slip", config.AuthOnly(salaryController.DownloadEmployeeSlip))

	// run service
	fmt.Println("Service is running on port 8000")
	http.ListenAndServe(":8000", nil)
}