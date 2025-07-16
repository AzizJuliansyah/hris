package models

import (
	"database/sql"
	"hris/entities"
	"time"
)

type OfficeModel struct {
	db *sql.DB
}

func NewOfficeModel(db *sql.DB) *OfficeModel {
	return &OfficeModel{
		db: db,
	}
}

func (model OfficeModel) AddOffice(office entities.Office) error {

	_, err := model.db.Exec(
		"INSERT INTO office (name, address, latitude, longitude, radius) VALUES (?,?,?,?,?)",
		office.Name, office.Address, office.Latitude, office.Longitude, office.Radius,
	)

	return err
}

func (model OfficeModel) FindAllOffice() ([]entities.Office, error) {
	rows, err := model.db.Query("SELECT id, name, address, latitude, longitude, radius FROM office WHERE deleted_at IS NULL")
	if err != nil {
		return []entities.Office{}, err
	}
	defer rows.Close()

	var offices []entities.Office

	for rows.Next() {
		var office entities.Office
		err := rows.Scan(
			&office.Id,
			&office.Name,
			&office.Address,
			&office.Latitude,
			&office.Longitude,
			&office.Radius,
		)
		if err != nil {
			return []entities.Office{}, err
		}
		offices = append(offices, office)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return offices, nil
}

func (model OfficeModel) FindOfficeByID(id int64) (entities.Office, error) {
	var office entities.Office

	query := `
		SELECT id, name, address, latitude, longitude, radius 
		FROM office 
		WHERE id = ? AND deleted_at IS NULL
	`
	err := model.db.QueryRow(query, id).Scan(
		&office.Id,
		&office.Name,
		&office.Address,
		&office.Latitude,
		&office.Longitude,
		&office.Radius,
	)

	return office, err
}

func (model OfficeModel) EditOffice(office entities.Office) error {
	query := `
		UPDATE office 
		SET name = ?, address = ?, latitude = ?, longitude = ?, radius = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	_, err := model.db.Exec(
		query,
		office.Name,
		office.Address,
		office.Latitude,
		office.Longitude,
		office.Radius,
		time.Now(),
		office.Id,
	)

	return err
}

func (model OfficeModel) SoftDeleteOffice(id int64) error {
	query := `
		UPDATE office 
		SET deleted_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	_, err := model.db.Exec(query, time.Now(), id)
	return err
}