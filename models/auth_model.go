package models

import (
	"database/sql"
	"hris/entities"
)

type AuthModel struct {
	db *sql.DB
}

func NewAuthModel(db *sql.DB) *AuthModel {
	return &AuthModel{
		db: db,
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