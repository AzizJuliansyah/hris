package models

import (
	"database/sql"
	"fmt"
	"hris/config"
	"hris/entities"
	"log"
	"time"
)

type SalaryModel struct {
	db *sql.DB
}

func NewSalaryModel() *SalaryModel {
	conn, err := config.DBConnection()
	if err != nil {
		log.Println("Failed connect to database: ", err)
	}
	return &SalaryModel{
		db: conn,
	}
}

func (model SalaryModel) FindAllSalaries() ([]entities.EmployeeSalary, error) {
	var employeeName sql.NullString
	query := `
		SELECT 
		s.id,
		s.nik,
		s.monthly_wages,
		s.daily_wages,
		s.meal_allowance,
		s.transport_allowance,
		e.name AS employee_name
		FROM salary s
		JOIN employee e ON s.nik = e.nik
		WHERE s.deleted_at IS NULL
	`

	rows, err := model.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var salaries []entities.EmployeeSalary
	for rows.Next() {
		var salary entities.EmployeeSalary
		err := rows.Scan(
			&salary.Id,
			&salary.NIK,
			&salary.Monthly_Wages,
			&salary.Daily_Wages,
			&salary.Meal_Allowance,
			&salary.Transport_Allowance,
			&employeeName,
		)
		if err != nil {
			return nil, err
		}
		if employeeName.Valid {
			salary.EmployeeName = employeeName.String
		} else {
			salary.EmployeeName = "-"
		}
		salaries = append(salaries, salary)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return salaries, nil
}

func (model SalaryModel) GetEmployeeNameandNIK() ([]entities.Employee, error) {
	query := `
		SELECT nik, name FROM employee WHERE deleted_at IS NULL
	`

	rows, err := model.db.Query(query)
	if err != nil {
		return []entities.Employee{}, err
	}
	defer rows.Close()

	var employees []entities.Employee
	for rows.Next() {
		var employee entities.Employee

		err := rows.Scan(
			&employee.NIK,
			&employee.Name,
		)
		if err != nil {
			return []entities.Employee{}, err
		}

		employees = append(employees, employee)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return employees, nil
}

func (model SalaryModel) IsEmployeeExistByNIK(nik string) (bool, error) {
	var count int
	err := model.db.QueryRow("SELECT COUNT(*) FROM employee WHERE nik = ?", nik).Scan(&count)
	return count > 0, err
}

func (model SalaryModel) IsSlipExist(nik string, year int, month int) (bool, error) {
	period := fmt.Sprintf("%04d-%02d", year, month)
	query := `SELECT COUNT(*) FROM salary_slip WHERE nik = ? AND period = ?`
	var count int
	err := model.db.QueryRow(query, nik, period).Scan(&count)
	return count > 0, err
}

func (model SalaryModel) CreateSalarySlip(slip entities.SalarySlip) error {
	query := `
		INSERT INTO salary_slip 
		(nik,
		period,
		monthly_wages,
		daily_wages,
		meal_allowance,
		transport_allowance,
		total_working_days,
		total_leave_days,
		monthly_wages_received,
		daily_wages_received,
		meal_allowance_received,
		transport_allowance_received,
		salary_total,
		created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := model.db.Exec(
		query,
		slip.NIK,
		slip.Period,
		slip.Monthly_Wages,
		slip.Daily_Wages,
		slip.Meal_Allowance,
		slip.Transport_Allowance,
		slip.Total_Working_Days,
		slip.Total_Leave_Days,
		slip.Monthly_Wages_Received,
		slip.Daily_Wages_Received,
		slip.Meal_Allowance_Received,
		slip.Transport_Allowance_Received,
		slip.Salary_Total,
		time.Now())
	return err
}

func (model SalaryModel) GetSalarySlipsByNIK(nik string) ([]entities.SalarySlip, error) {
	query := `
		SELECT id, nik, period, created_at
		FROM salary_slip
		WHERE nik = ? AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := model.db.Query(query, nik)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slips []entities.SalarySlip
	for rows.Next() {
		var slip entities.SalarySlip
		err := rows.Scan(
			&slip.Id,
			&slip.NIK,
			&slip.Period,
			&slip.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		t, err := time.Parse("2006-01", slip.Period)
		if err == nil {
			slip.FormattedPeriod = t.Format("January 2006")
		} else {
			slip.FormattedPeriod = slip.Period
		}
		slips = append(slips, slip)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return slips, nil
}

func (model SalaryModel) GetSalarySlipByID(id int64) (entities.SalarySlip, error) {
	var slip entities.SalarySlip
	query := `
		SELECT 
			s.id,
			s.nik, 
			s.period,
			s.monthly_wages,
			s.daily_wages,
			s.meal_allowance,
			s.transport_allowance,
			s.total_working_days, 
			s.total_leave_days,
			s.monthly_wages_received,
			s.daily_wages_received,
			s.meal_allowance_received,
			s.transport_allowance_received,
			s.salary_total,
			s.created_at,
			e.name AS employee_name
		FROM salary_slip s
		JOIN employee e ON s.nik = e.nik
		WHERE s.id = ?
	`

	err := model.db.QueryRow(query, id).Scan(
		&slip.Id,
		&slip.NIK,
		&slip.Period,
		&slip.Monthly_Wages,
		&slip.Daily_Wages,
		&slip.Meal_Allowance,
		&slip.Transport_Allowance,
		&slip.Total_Working_Days,
		&slip.Total_Leave_Days,
		&slip.Monthly_Wages_Received,
		&slip.Daily_Wages_Received,
		&slip.Meal_Allowance_Received,
		&slip.Transport_Allowance_Received,
		&slip.Salary_Total,
		&slip.CreatedAt,
		&slip.EmployeeName,
	)

	if err != nil {
		return slip, err
	}

	t, err := time.Parse("2006-01", slip.Period)
	if err == nil {
		slip.FormattedPeriod = t.Format("January 2006")
	} else {
		slip.FormattedPeriod = slip.Period
	}

	return slip, nil
}

func (model SalaryModel) InputEmployeeSalary(salary entities.EmployeeSalary) error {
	query := `
	INSERT INTO salary (nik, monthly_wages, daily_wages, meal_allowance, transport_allowance)
	VALUES (?, ?, ?, ?, ?) 
	`

	_, err := model.db.Exec(
		query,
		salary.NIK,
		salary.Monthly_Wages,
		salary.Daily_Wages,
		salary.Meal_Allowance,
		salary.Transport_Allowance,
	)

	return err
}

func (model SalaryModel) FindSalaryByID(id int64) (entities.EmployeeSalary, error) {
	var salary entities.EmployeeSalary
	query := `
	SELECT 
		s.id,
		s.nik,
		s.monthly_wages,
		s.daily_wages,
		s.meal_allowance,
		s.transport_allowance,
		e.name AS employee_name
	FROM salary s
	JOIN employee e ON s.nik = e.nik
	WHERE s.id = ? AND s.deleted_at IS NULL
	`

	err := model.db.QueryRow(query, id).Scan(
		&salary.Id,
		&salary.NIK,
		&salary.Monthly_Wages,
		&salary.Daily_Wages,
		&salary.Meal_Allowance,
		&salary.Transport_Allowance,
		&salary.EmployeeName,
	)

	if err != nil {
		return salary, err
	}

	return salary, nil
}

func (model SalaryModel) EditEmployeeSalary(salary entities.EditEmployeeSalary) error {
	query := `
	UPDATE salary
	SET monthly_wages = ?, daily_wages = ?, meal_allowance = ?, transport_allowance = ?, updated_at = ?
	WHERE id = ?
	`

	_, err := model.db.Exec(
		query,
		salary.Monthly_Wages,
		salary.Daily_Wages,
		salary.Meal_Allowance,
		salary.Transport_Allowance,
		time.Now(),
		salary.Id,
	)

	return err
}

func (model SalaryModel) DeleteSalary(id int64) error {
	query := `
		DELETE FROM salary
		WHERE id = ?
	`

	_, err := model.db.Exec(query, id)
	return err
}

func (model SalaryModel) DeleteSalarySlip(id int64) error {
	query := `
		DELETE FROM salary_slip
		WHERE id = ?
	`

	_, err := model.db.Exec(query, id)
	return err
}