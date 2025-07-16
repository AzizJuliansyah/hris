package models

import (
	"database/sql"
	"hris/entities"
	"time"
)

type ShiftModel struct {
	db *sql.DB
}

func NewShiftModel(db *sql.DB) *ShiftModel {
	return &ShiftModel{
		db: db,
	}
}

func (model ShiftModel) AddShift(shift entities.Shift) error {
	_, err := model.db.Exec(
		"INSERT INTO shift (name, start_time, end_time) VALUES (?, ?, ?)",
		shift.Name, shift.StartTime, shift.EndTime,
	)
	return err
}

func (model ShiftModel) FindAllShift() ([]entities.Shift, error) {
	rows, err := model.db.Query(
		"SELECT id, name, start_time, end_time FROM shift WHERE deleted_at IS NULL",
	)
	if err != nil {
		return []entities.Shift{}, err
	}
	defer rows.Close()

	var shifts []entities.Shift
	for rows.Next() {
		var shift entities.Shift
		err := rows.Scan(&shift.Id, &shift.Name, &shift.StartTime, &shift.EndTime)
		if err != nil {
			return []entities.Shift{}, err
		}
		shifts = append(shifts, shift)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return shifts, nil
}

func (model ShiftModel) FindShiftByID(id int64) (entities.Shift, error) {
	var shift entities.Shift
	query := `
		SELECT id, name, start_time, end_time 
		FROM shift 
		WHERE id = ? AND deleted_at IS NULL
	`
	err := model.db.QueryRow(query, id).Scan(
		&shift.Id,
		&shift.Name,
		&shift.StartTime,
		&shift.EndTime,
	)
	return shift, err
}

func (model ShiftModel) EditShift(shift entities.Shift) error {
	query := `
		UPDATE shift 
		SET name = ?, start_time = ?, end_time = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	_, err := model.db.Exec(
		query,
		shift.Name,
		shift.StartTime,
		shift.EndTime,
		time.Now(),
		shift.Id,
	)
	return err
}

// SoftDeleteShift menandai shift sebagai dihapus (soft delete)
func (model ShiftModel) SoftDeleteShift(id int64) error {
	query := `
		UPDATE shift 
		SET deleted_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	_, err := model.db.Exec(query, time.Now(), id)
	return err
}