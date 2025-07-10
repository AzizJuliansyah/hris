package entities

import "database/sql"

type EmployeeSalary struct {
	Id                  int64
	NIK                 string         `validate:"required,isunique=salary-nik" label:"NIK Karyawan"`
	EmployeeName        string
	Monthly_Wages       sql.NullString `validate:"required" label:"Gaji Bulanan"`
	Daily_Wages         sql.NullString `validate:"required" label:"Gaji Harian"`
	Meal_Allowance      sql.NullString `validate:"required" label:"Tunjangan Makan"`
	Transport_Allowance sql.NullString `validate:"required" label:"Tunjangan Transport"`
}

type EditEmployeeSalary struct {
	Id                  int64
	EmployeeName        string
	Monthly_Wages       sql.NullString `validate:"required" label:"Gaji Bulanan"`
	Daily_Wages         sql.NullString `validate:"required" label:"Gaji Harian"`
	Meal_Allowance      sql.NullString `validate:"required" label:"Tunjangan Makan"`
	Transport_Allowance sql.NullString `validate:"required" label:"Tunjangan Transport"`
}

type SalarySlip struct {
	Id                          int64
	NIK                         string
	EmployeeName        		string
	Period                      string         
	FormattedPeriod			 	string         
	Monthly_Wages               sql.NullString
	Daily_Wages                 sql.NullString
	Meal_Allowance              sql.NullString 
	Transport_Allowance         sql.NullString 
	Total_Working_Days          int            
	Total_Leave_Days            int           
	Monthly_Wages_Received      sql.NullInt64 
	Daily_Wages_Received        sql.NullInt64 
	Meal_Allowance_Received     sql.NullInt64 
	Transport_Allowance_Received sql.NullInt64
	Salary_Total                sql.NullInt64
	CreatedAt                   sql.NullTime 
}
