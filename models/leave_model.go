package models

import (
	"database/sql"
	"hris/config"
	"hris/entities"
	"log"
	"strings"
	"time"
)

type LeaveModel struct {
	db *sql.DB
}

func NewLeaveModel() *LeaveModel {
	conn, err := config.DBConnection()
	if err != nil {
		log.Println("Failed connect to database:", err)
	}
	return &LeaveModel{
		db: conn,
	}
}

// awal tambahan
func (model LeaveModel) FindAllLeaveType() ([]entities.LeaveType, error) {
	rows, err := model.db.Query("SELECT id, name, max_day FROM leave_type WHERE deleted_at IS NULL")
	if err != nil {
		return []entities.LeaveType{}, err
	}
	defer rows.Close()

	var leaveType []entities.LeaveType

	for rows.Next() {
		var leave entities.LeaveType
		err := rows.Scan(
			&leave.Id,
			&leave.Name,
			&leave.MaxDay,
		)
		if err != nil {
			return []entities.LeaveType{}, err
		}
		leaveType = append(leaveType, leave)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return leaveType, nil
}

func (model LeaveModel) AddLeaveType(leave entities.LeaveType) error {
	query := `INSERT INTO leave_type (name, max_day) VALUES (?, ?)`

	_, err := model.db.Exec(
		query,
		leave.Name,
		leave.MaxDay,
	)

	return err
}

func (model LeaveModel) EditLeaveType(leave entities.LeaveType) error {
	query := `UPDATE leave_type set Name = ?, max_day = ?, updated_at = ? WHERE id = ?`

	_, err := model.db.Exec(
		query,
		leave.Name,
		leave.MaxDay,
		time.Now(),
		leave.Id,
	)

	return err
}

func (model LeaveModel) DeleteLeaveType(id int64) error {
	query := `UPDATE leave_type set deleted_at = ? WHERE id = ?`

	_, err := model.db.Exec(
		query,
		time.Now(),
		id,
	)

	return err
}

func (model LeaveModel) InsertLeave(data entities.SubmitLeave) error {

	query := `
		INSERT INTO leave_employee 
		(nik, leave_type_id, leave_date, attachment, reason, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := model.db.Exec(
		query,
		data.NIK,
		data.LeaveTypeId,
		data.LeaveDateJoin,
		data.Attachment,
		data.Reason,
		data.Status,
	)

	return err
}

func (model LeaveModel) GetLeaveList(nik string, monthYear string, todayOnly bool) ([]entities.Leave, error) {
	var query string
	var adminName sql.NullString
	var args []interface{}

	parsedDate, err := time.Parse("January 2006", monthYear)
	if err != nil {
		return nil, err
	}

	baseQuery := `
		SELECT 
			le.id,
			le.nik,
			le.leave_type_id,
			lt.name AS leave_type_name,
			le.leave_date,
			le.attachment,
			le.reason,
			le.status,
			le.reason_status,
			le.created_at,
			le.updated_at,
			em.name AS employee_name,
			admin.name AS admin_name,
			em.uuid AS employee_uuid
		FROM leave_employee le
		LEFT JOIN leave_type lt ON le.leave_type_id = lt.id
		LEFT JOIN employee em ON le.nik = em.nik
		LEFT JOIN employee admin ON le.admin_nik = admin.nik
		WHERE le.deleted_at IS NULL
		AND MONTH(le.created_at) = ?
		AND YEAR(le.created_at) = ?
	`

	args = []interface{}{parsedDate.Month(), parsedDate.Year()}

	if todayOnly {
		baseQuery += " AND DATE(le.created_at) = CURDATE()"
	}
	if nik != "" {
		baseQuery += " AND le.nik = ?"
		args = append(args, nik)
	}
	query = baseQuery + " ORDER BY le.created_at DESC"

	rows, err := model.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaves []entities.Leave
	for rows.Next() {
		var leave entities.Leave
		err := rows.Scan(
			&leave.Id,
			&leave.NIK,
			&leave.LeaveTypeId,
			&leave.LeaveTypeName,
			&leave.LeaveDateJoin,
			&leave.Attachment,
			&leave.Reason,
			&leave.Status,
			&leave.ReasonStatus,
			&leave.CreatedAt,
			&leave.UpdatedAt,
			&leave.EmployeeName,
			&adminName,
			&leave.UUID,
		)
		if err != nil {
			return nil, err
		}

		if adminName.Valid {
			leave.AdminName = adminName
		} else {
			leave.AdminName.String = "-"
		}

		// Parse LeaveDateJoin ke []time.Time
		if leave.LeaveDateJoin != "" {
			dateStrings := strings.Split(leave.LeaveDateJoin, ",")
			for _, ds := range dateStrings {
				parsedDate, err := time.Parse("2006-01-02", strings.TrimSpace(ds))
				if err == nil {
					leave.LeaveDate = append(leave.LeaveDate, parsedDate)
				}
			}
		}

		leaves = append(leaves, leave)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return leaves, nil
}

func (model LeaveModel) CountAllLeave() (int, error) {
	row := model.db.QueryRow("SELECT Count(*) FROM leave_employee WHERE deleted_at IS NULL")

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (model LeaveModel) GetLeaveById(id int64) (*entities.Leave, error) {
	query := `
		SELECT 
			le.id,
			le.nik,
			le.leave_type_id,
			lt.name AS leave_type_name,
			le.leave_date,
			le.attachment,
			le.reason,
			le.status,
			le.reason_status,
			le.created_at,
			le.updated_at,
			em.name
		FROM leave_employee le
		LEFT JOIN leave_type lt ON le.leave_type_id = lt.id
		LEFT JOIN employee em ON le.nik = em.nik
		WHERE le.deleted_at IS NULL AND le.id = ?
		LIMIT 1
	`

	var leave entities.Leave

	err := model.db.QueryRow(query, id).Scan(
		&leave.Id,
		&leave.NIK,
		&leave.LeaveTypeId,
		&leave.LeaveTypeName,
		&leave.LeaveDateJoin,
		&leave.Attachment,
		&leave.Reason,
		&leave.Status,
		&leave.ReasonStatus,
		&leave.CreatedAt,
		&leave.UpdatedAt,
		&leave.EmployeeName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Tidak ditemukan
		}
		return nil, err
	}

	// Parse LeaveDateJoin ke []time.Time
	if leave.LeaveDateJoin != "" {
		dateStrings := strings.Split(leave.LeaveDateJoin, ",")
		for _, ds := range dateStrings {
			parsedDate, err := time.Parse("2006-01-02", strings.TrimSpace(ds))
			if err == nil {
				leave.LeaveDate = append(leave.LeaveDate, parsedDate)
			}
		}
	}

	return &leave, nil
}

func (model LeaveModel) FindLeavesByNIK(nik string) ([]entities.LeaveInAMonth, error) {
	query := `
		SELECT id, nik, leave_date, status 
		FROM leave_employee
		WHERE nik = ? AND status = 1
	`
	rows, err := model.db.Query(query, nik)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaves []entities.LeaveInAMonth
	for rows.Next() {
		var leave entities.LeaveInAMonth
		err := rows.Scan(&leave.Id, &leave.NIK, &leave.LeaveDate, &leave.Status)
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, leave)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return leaves, nil
}

func (model LeaveModel) CountLeaveDaysThisMonth(nik string, month time.Month, year int) (int, error) {
	query := `
		SELECT leave_date 
		FROM leave 
		WHERE nik = ? AND status = 1
	`

	rows, err := model.db.Query(query, nik)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var leaveDates string
		if err := rows.Scan(&leaveDates); err != nil {
			continue
		}

		dates := strings.Split(leaveDates, ",")
		for _, d := range dates {
			t, err := time.Parse("2006-01-02", strings.TrimSpace(d))
			if err == nil && t.Month() == month && t.Year() == year {
				count++
			}
		}
	}

	if err := rows.Err(); err != nil {
		return 0, err
	}

	return count, nil
}


func (model LeaveModel) UpdateLeaveStatus(approvalLeave entities.ApprovalLeave) error {
	query := `
		UPDATE leave_employee
		SET admin_nik = ?, status = ?, reason_status = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	_, err := model.db.Exec(
		query,
		approvalLeave.AdminNIK,
		approvalLeave.Status,
		approvalLeave.ReasonStatus,
		time.Now(),
		approvalLeave.Id,
	)
	return err
}