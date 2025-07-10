package models

import (
	"database/sql"
	"hris/config"
	"hris/entities"
	"log"
	"time"

	"github.com/goodsign/monday"
)

type EmployeeModel struct {
	db *sql.DB
}

func NewEmployeeModel() *EmployeeModel {
	conn, err := config.DBConnection()
	if err != nil {
		log.Println("Failed connect to database: ", err)
	}
	return &EmployeeModel{
		db: conn,
	}
}

func (model EmployeeModel) FindAllEmployee(adminOnly bool, employeeOnly bool) ([]entities.Employee, error) {
	var photo sql.NullString

	query := `
		SELECT uuid, nik, name, email, phone, gender, birth_date, is_admin, photo
		FROM employee WHERE deleted_at IS NULL
	`
	
	if adminOnly {
		query += " AND is_admin = 1"
	}
	if employeeOnly {
		query += " AND is_admin = 0"
	}
	
	rows, err := model.db.Query(query)
	if err != nil {
		return []entities.Employee{}, err
	}
	defer rows.Close()

	var employees []entities.Employee
	for rows.Next() {
		var employee entities.Employee
		var birthDateTime time.Time
		err := rows.Scan(
			&employee.UUID,
			&employee.NIK,
			&employee.Name,
			&employee.Email,
			&employee.Phone,
			&employee.Gender,
			&birthDateTime,
			&employee.IsAdmin,
			&photo,
		)
		if err != nil {
			return []entities.Employee{}, err
		}

		employee.BirthDate = monday.Format(birthDateTime, "01 Januari 2006", monday.LocaleIdID)
		if photo.Valid {
			employee.Photo = photo
		} else {
			employee.Photo = sql.NullString{String: "", Valid: false}
		}

		employees = append(employees, employee)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return employees, nil
}


func (model EmployeeModel) CountAllActiveEmployee() (int, error) {
	row := model.db.QueryRow(`SELECT COUNT(*) FROM employee WHERE deleted_at IS NULL`)
	
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (model EmployeeModel) AddEmployee(employee entities.Employee) error {
	_, err := model.db.Exec(
		"INSERT INTO employee (uuid, nik, name, email, phone, address, gender, birth_date, is_admin, password) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		employee.UUID, employee.NIK, employee.Name, employee.Email, employee.Phone, employee.Address, employee.Gender, employee.BirthDate, employee.IsAdmin, employee.Password,
	)
	return err
}



func (model EmployeeModel) FindEmployeeByUUID(uuid string) (entities.Employee, error) {
	var employee entities.Employee
	var birthDateTime time.Time
	var photo sql.NullString

	query := "SELECT uuid, nik, name, email, phone, address, gender, birth_date, is_admin, photo FROM employee WHERE uuid = ? AND deleted_at IS NULL"

	err := model.db.QueryRow(query, uuid).Scan(
		&employee.UUID,
		&employee.NIK,
		&employee.Name,
		&employee.Email,
		&employee.Phone,
		&employee.Address,
		&employee.Gender,
		&birthDateTime,
		&employee.IsAdmin,
		&photo,
	)
	
	
	if err != nil {
		return employee, err
	}
	employee.BirthDate = birthDateTime.Format("2006-01-02")
	employee.BirthDateFormat = monday.Format(birthDateTime, "01 Januari 2006", monday.LocaleIdID)
	if photo.Valid {
		employee.Photo = photo
	} else {
		employee.Photo = sql.NullString{String: "", Valid: false}
	}

	return employee, nil
}

func (model EmployeeModel) EditEmployee(employee entities.EditEmployee) error {
	query := `
		UPDATE employee
		SET nik = ?, name = ?, email = ?, phone = ?, address = ?, gender = ? ,birth_date = ?, is_admin = ?, updated_at = ? WHERE uuid = ? 
	`

	_, err := model.db.Exec(
		query,
		employee.NIK,
		employee.Name,
		employee.Email,
		employee.Phone,
		employee.Address,
		employee.Gender,
		employee.BirthDate,
		employee.IsAdmin,
		time.Now(),
		employee.UUID,
	)

	return err
}


func (model EmployeeModel) DeleteEmployee(uuid string) error {
	query := `
		UPDATE employee
		SET deleted_at = ?
		WHERE uuid = ?
	`

	_, err := model.db.Exec(query, time.Now(), uuid)

	return err
}
