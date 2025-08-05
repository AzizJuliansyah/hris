package models

import (
	"database/sql"
	"fmt"
	"hris/entities"
	"time"
)

type KPIModel struct {
	db *sql.DB
}

func NewKPIModel(db *sql.DB) *KPIModel {
	return &KPIModel{
		db: db,
	}
}

func (model KPIModel) FindKPIList(nik string, monthYear string, todayOnly bool) ([]entities.KPI, error) {
	var started_atTime sql.NullTime
	var completed_atTime sql.NullTime
	var actual_value sql.NullInt64
	var score sql.NullInt64
	var admin_notes sql.NullString
	var checked_at sql.NullTime
	var created_at sql.NullTime
	var updated_at sql.NullTime
	var adminName sql.NullString
	var employeeName sql.NullString
	var query string
	var args []interface{}

	parsedDate, err := time.Parse("January 2006", monthYear)
	if err != nil {
		return nil, err
	}

	baseQuery := `
		SELECT
			kpi.id,
			kpi.employee_nik,
			kpi.admin_nik,
			kpi.title,
			kpi.description,
			kpi.target_value,
			kpi.unit,
			kpi.status,
			kpi.deadline,
			kpi.started_at,
			kpi.completed_at,
			kpi.actual_value,
			kpi.evidence_file,
			kpi.final_status,
			kpi.score,
			kpi.admin_notes,
			kpi.checked_at,
			kpi.created_at,
			kpi.updated_at,
			admin.name AS admin_name,
			employee.name AS employee_name
		FROM kpi
		JOIN employee AS admin ON kpi.admin_nik = admin.nik
		LEFT JOIN employee AS employee ON kpi.employee_nik = employee.nik
		WHERE kpi.deleted_at IS NULL
		AND MONTH(kpi.created_at) = ?
		AND YEAR(kpi.created_at) = ?
	`

	args = []interface{}{parsedDate.Month(), parsedDate.Year()}

	if todayOnly {
		baseQuery += " AND DATE(kpi.created_at) = CURDATE()"
	}
	if nik != "" {
		baseQuery += " AND kpi.employee_nik = ?"
		args = append(args, nik)
	}
	query = baseQuery + " ORDER BY kpi.created_at DESC"

	rows, err := model.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var kpis []entities.KPI
	for rows.Next() {
		var kpi entities.KPI
		var deadline time.Time

		err := rows.Scan(
			&kpi.ID,
			&kpi.Employee_NIK,
			&kpi.Admin_NIK,
			&kpi.Title,
			&kpi.Description,
			&kpi.Target_Value,
			&kpi.Unit,
			&kpi.Status,
			&deadline,
			&started_atTime,
			&completed_atTime,
			&actual_value,
			&kpi.Evidence_File,
			&kpi.Final_Status,
			&score,
			&admin_notes,
			&checked_at,
			&created_at,
			&updated_at,
			&adminName,
			&employeeName,
		)
		if err != nil {
			return nil, err
		}

		// Format waktu
		kpi.Deadline = deadline.Format("02 January 2006 15:04")

		if started_atTime.Valid {
			kpi.Started_At = started_atTime.Time.Format("02 January 2006 15:04")
		} else {
			kpi.Started_At = "-"
		}

		if completed_atTime.Valid {
			kpi.Completed_At = completed_atTime.Time.Format("02 January 2006 15:04")
		} else {
			kpi.Completed_At = "-"
		}

		if actual_value.Valid {
			kpi.Actual_Value = fmt.Sprintf("%d", actual_value.Int64)
		} else {
			kpi.Actual_Value = "-"
		}

		if score.Valid {
			kpi.Score = fmt.Sprintf("%d", score.Int64)
		} else {
			kpi.Score = "-"
		}

		if admin_notes.Valid {
			kpi.Admin_Notes = admin_notes.String
		} else {
			kpi.Admin_Notes = "-"
		}

		if checked_at.Valid {
			kpi.Checked_At = checked_at.Time.Format("02 January 2006 15:04")
		} else {
			kpi.Checked_At = "-"
		}

		if created_at.Valid {
			kpi.Created_At = created_at.Time.Format("02 January 2006 15:04")

			if time.Since(created_at.Time).Hours() > 24 {
				kpi.IsCreatedOver1Day = true
			}
		} else {
			kpi.Created_At = "-"
		}

		if updated_at.Valid {
			kpi.Updated_At = updated_at.Time.Format("02 January 2006 15:04")
		} else {
			kpi.Updated_At = "-"
		}

		if adminName.Valid {
			kpi.Admin_Name = adminName.String
		} else {
			kpi.Admin_Name = "-"
		}

		if employeeName.Valid {
			kpi.Employee_Name = employeeName.String
		} else {
			kpi.Employee_Name = "-"
		}

		kpis = append(kpis, kpi)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return kpis, nil
}

func (model KPIModel) FindKPIByID(id int64) (entities.KPI, error) {
	var started_atTime sql.NullTime
	var completed_atTime sql.NullTime
	var actual_value sql.NullInt64
	var score sql.NullInt64
	var admin_notes sql.NullString
	var checked_at sql.NullTime
	var created_at sql.NullTime
	var updated_at sql.NullTime
	var adminName sql.NullString
	var employeeName sql.NullString
	var deadline time.Time

	query := `
		SELECT
			kpi.id,
			kpi.employee_nik,
			kpi.admin_nik,
			kpi.title,
			kpi.description,
			kpi.target_value,
			kpi.unit,
			kpi.status,
			kpi.deadline,
			kpi.started_at,
			kpi.completed_at,
			kpi.actual_value,
			kpi.evidence_file,
			kpi.final_status,
			kpi.score,
			kpi.admin_notes,
			kpi.checked_at,
			kpi.created_at,
			kpi.updated_at,
			admin.name AS admin_name,
			employee.name AS employee_name
		FROM kpi
		JOIN employee AS admin ON kpi.admin_nik = admin.nik
		LEFT JOIN employee AS employee ON kpi.employee_nik = employee.nik
		WHERE kpi.id = ? AND kpi.deleted_at IS NULL
	`

	
		var kpi entities.KPI
	
		err := model.db.QueryRow(query, id).Scan(
			&kpi.ID,
			&kpi.Employee_NIK,
			&kpi.Admin_NIK,
			&kpi.Title,
			&kpi.Description,
			&kpi.Target_Value,
			&kpi.Unit,
			&kpi.Status,
			&deadline,
			&started_atTime,
			&completed_atTime,
			&actual_value,
			&kpi.Evidence_File,
			&kpi.Final_Status,
			&score,
			&admin_notes,
			&checked_at,
			&created_at,
			&updated_at,
			&adminName,
			&employeeName,
		)
		// Format waktu
		kpi.Deadline = deadline.Format("02 January 2006 15:04")
		kpi.DeadlineTimeFormat = deadline.Format("2006-01-02T15:04")

		if started_atTime.Valid {
			kpi.Started_At = started_atTime.Time.Format("02 January 2006 15:04")
		} else {
			kpi.Started_At = "-"
		}

		if completed_atTime.Valid {
			kpi.Completed_At = completed_atTime.Time.Format("02 January 2006 15:04")
		} else {
			kpi.Completed_At = "-"
		}

		if actual_value.Valid {
			kpi.Actual_Value = fmt.Sprintf("%d", actual_value.Int64)
		} else {
			kpi.Actual_Value = "-"
		}

		if score.Valid {
			kpi.Score = fmt.Sprintf("%d", score.Int64)
		} else {
			kpi.Score = "-"
		}

		if admin_notes.Valid {
			kpi.Admin_Notes = admin_notes.String
		} else {
			kpi.Admin_Notes = "-"
		}

		if checked_at.Valid {
			kpi.Checked_At = checked_at.Time.Format("02 January 2006 15:04")
		} else {
			kpi.Checked_At = "-"
		}

		if created_at.Valid {
			kpi.Created_At = created_at.Time.Format("02 January 2006 15:04")
		} else {
			kpi.Created_At = "-"
		}

		if updated_at.Valid {
			kpi.Updated_At = updated_at.Time.Format("02 January 2006 15:04")
		} else {
			kpi.Updated_At = "-"
		}

		if adminName.Valid {
			kpi.Admin_Name = adminName.String
		} else {
			kpi.Admin_Name = "-"
		}

		if employeeName.Valid {
			kpi.Employee_Name = employeeName.String
		} else {
			kpi.Employee_Name = "-"
		}

	return kpi, err
}


func (model KPIModel) AddKPI(kpi entities.AddKPI) error {
	query := `INSERT INTO kpi (employee_nik, admin_nik, title, description, target_value, unit, deadline) VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := model.db.Exec(
		query,
		kpi.Employee_NIK,
		kpi.Admin_NIK,
		kpi.Title,
		kpi.Description,
		kpi.Target_Value,
		kpi.Unit,
		kpi.Deadline,
	)

	return err
}

func (model KPIModel) EditKPI(kpi entities.EditKPI) error {
	query := `UPDATE kpi SET employee_nik = ?, updated_admin_nik = ?, title = ?, description = ?, target_value = ?, unit = ?, deadline = ?, status = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`

	_, err := model.db.Exec(
		query,
		kpi.Employee_NIK,
		kpi.Updated_Admin_NIK,
		kpi.Title,
		kpi.Description,
		kpi.Target_Value,
		kpi.Unit,
		kpi.Deadline,
		kpi.Status,
		time.Now(),
		kpi.ID,
	)

	return err
}

func (model KPIModel) InputEvaluasiKPI(kpi entities.EvaluasiKPI) error {
	query := `UPDATE kpi SET checked_admin_nik = ?, final_status = ?, score = ?, admin_notes = ?, checked_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`

	_, err := model.db.Exec(
		query,
		kpi.Checked_Admin_NIK,
		kpi.Final_Status,
		kpi.Score,
		kpi.Admin_Notes,
		time.Now(),
		time.Now(),
		kpi.ID,
	)

	return err
}

func (model KPIModel) ResetEvaluasiKPI(id int64, nik string) error {
	query := `
		UPDATE kpi
		SET 
			final_status = "pending", 
			score = NULL, 
			admin_notes = NULL, 
			checked_admin_nik = NULL, 
			checked_at = NULL, 
			reset_evaluasi_admin_nik = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	_, err := model.db.Exec(query, nik, id)
	return err
}


func (model KPIModel) StartKPI(id int64) error {
	query := `
		UPDATE kpi
		SET 
			started_at = ?,
			status = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	_, err := model.db.Exec(query, time.Now(), "in_progress", id)
	return err
}

func (model KPIModel) FinishKPI(id int64) error {
	query := `
		UPDATE kpi
		SET 
			completed_at = ?,
			status = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	_, err := model.db.Exec(query, time.Now(), "done", id)
	return err
}

func (model KPIModel) GetKPIEvidenceFileByID(id int64) (sql.NullString, error) {
	var evidence_file sql.NullString
	err := model.db.QueryRow("SELECT evidence_file FROM kpi WHERE id = ? AND deleted_at IS NULL", id).Scan(&evidence_file)
	return evidence_file, err
}

func (model KPIModel) GetKPITargetValueByID(id int64) (string, error) {
	var target_value string
	err := model.db.QueryRow("SELECT target_value FROM kpi WHERE id = ? AND deleted_at IS NULL", id).Scan(&target_value)
	return target_value, err
}

func (model KPIModel) SubmitProgressKPI(kpi entities.SubmitProgressKPI) error {
	query := `
		UPDATE kpi
		SET 
			actual_value = ?,
			evidence_file = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	_, err := model.db.Exec(
		query, 
		kpi.Actual_Value,
		kpi.Evidence_File,
		kpi.ID,
	) 
	return err
}

func (model KPIModel) SoftDeleteKPI(id int64) error {
	query := `
		UPDATE kpi
		SET deleted_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	_, err := model.db.Exec(query, time.Now(), id)
	return err
}