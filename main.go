package main

import (
	"fmt"
	"hris/config"
	"hris/controllers"
	"net/http"
)

func main() {
	
	// baca file untuk resource yg ada di static (css, js, img)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("views/static"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("public/images"))))

	
	// auth routes
	http.HandleFunc("/login", config.GuestOnly(controllers.Login))
	http.HandleFunc("/logout", controllers.Logout)

	// user routes
	http.HandleFunc("/pages-profile", config.AuthOnly(controllers.Profile))

	// home routes
	http.HandleFunc("/home", config.EmployeeOnly(controllers.Home))
	http.HandleFunc("/home-admin", config.AdminOnly(controllers.HomeAdmin))

	// news routes
	http.HandleFunc("/news", config.AdminOnly(controllers.ListNews))
	http.HandleFunc("/news/add-news", config.AdminOnly(controllers.AddNews))
	http.HandleFunc("/news/edit-news", config.AdminOnly(controllers.EditNews))
	http.HandleFunc("/news/delete-news", config.AdminOnly(controllers.DeleteNews))
	
	// employee routes
	http.HandleFunc("/employee", config.AdminOnly(controllers.FindAllEmployee))
	http.HandleFunc("/employee/add-employee", config.AdminOnly(controllers.AddEmployee))
	http.HandleFunc("/employee/detail-employee", config.AdminOnly(controllers.DetailEmployee))
	http.HandleFunc("/employee/edit-employee", config.AdminOnly(controllers.EditEmployee))
	http.HandleFunc("/employee/delete-employee", config.AdminOnly(controllers.DeleteEmployee))

	// office routes
	http.HandleFunc("/office", config.AdminOnly(controllers.Office))                    
	http.HandleFunc("/office/add-office", config.AdminOnly(controllers.AddOffice))      
	http.HandleFunc("/office/edit-office", config.AdminOnly(controllers.EditOffice))    
	http.HandleFunc("/office/delete-office", config.AdminOnly(controllers.DeleteOffice))
	
	// shift routes
	http.HandleFunc("/shift", config.AdminOnly(controllers.FindAllShift))                    
	http.HandleFunc("/shift/add-shift", config.AdminOnly(controllers.AddShift))       
	http.HandleFunc("/shift/edit-shift", config.AdminOnly(controllers.EditShift))     
	http.HandleFunc("/shift/delete-shift", config.AdminOnly(controllers.DeleteShift))

	// attendance routes
	http.HandleFunc("/attendance-submit", config.EmployeeOnly(controllers.SubmitAttendance))
	http.HandleFunc("/attendance-list", config.AdminOnly(controllers.ListAttendance))
	
	// leave routes
	http.HandleFunc("/leave/leave-type", config.AdminOnly(controllers.LeaveType))
	http.HandleFunc("/leave/delete-leave-type", config.AdminOnly(controllers.DeleteLeaveType))
	
	http.HandleFunc("/leave-list", config.AdminOnly(controllers.ListLeave))
	http.HandleFunc("/leave-submit", config.EmployeeOnly(controllers.SubmitLeave))
	http.HandleFunc("/leave/approval", config.AdminOnly(controllers.ApprovalLeave))

	// salary routes
	http.HandleFunc("/salary-list", config.AdminOnly(controllers.ListSalary))
	http.HandleFunc("/salary/detail-salary", config.AdminOnly(controllers.DetailEmployeeSalary))
	http.HandleFunc("/slip-list", config.EmployeeOnly(controllers.SlipListEmployeeSalary))
	http.HandleFunc("/salary/input-salary", config.AdminOnly(controllers.InputEmployeeSalary))
	http.HandleFunc("/salary/edit-salary", config.AdminOnly(controllers.EditEmployeeSalary))
	http.HandleFunc("/salary/delete-salary", config.AdminOnly(controllers.DeleteEmployeeSalary))
	http.HandleFunc("/salary/delete-slip", config.AdminOnly(controllers.DeleteEmployeeSlip))
	http.HandleFunc("/salary/download-slip", config.AuthOnly(controllers.DownloadEmployeeSlip))

	// run service
	fmt.Println("Service is running on port 8000")
	http.ListenAndServe(":8000", nil)
}