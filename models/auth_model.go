package models

import (
	"database/sql"
	"hris/config"
	"hris/entities"
	"log"
)

type AuthModel struct {
	db *sql.DB
}

func NewAuthModel() *AuthModel {
	conn, err := config.DBConnection()
	if err != nil {
		log.Println("Failed connect to database: ", err)
	}
	return &AuthModel{
		db: conn,
	}
}

func (model AuthModel) FindEmployeeByNIK(nik string) (entities.Employee, error) {
	var employee entities.Employee
	query := `
		SELECT nik, name, is_admin, password FROM employee WHERE nik = ? AND deleted_at IS NULL
	`

	err := model.db.QueryRow(query, nik). Scan(
		&employee.NIK,
		&employee.Name,
		&employee.IsAdmin,
		&employee.Password,
	)

	if err != nil {
		return employee, err
	}

	return employee, nil
}