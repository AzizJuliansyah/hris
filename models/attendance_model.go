package models

import (
	"database/sql"
	"hris/config"
	"hris/entities"
	"hris/helpers"
	"log"
	"time"

	"github.com/goodsign/monday"
)

type AttendanceModel struct {
	db *sql.DB
}

func NewAttendanceModel() *AttendanceModel {
	conn, err := config.DBConnection()
	if err != nil {
		log.Println("Failed connect to database: ", err)
	}
	return &AttendanceModel{
		db: conn,
	}
}

func (model AttendanceModel) GetLastAttendance(nik string) string {
	row := model.db.QueryRow(`
        SELECT id, nik, checkin_time, checkout_time
        FROM attendance
        WHERE nik = ?
          AND deleted_at IS NULL
        ORDER BY checkin_time DESC LIMIT 1`, nik)

	var att entities.Attendance
	err := row.Scan(&att.ID, &att.NIK, &att.CheckInTime, &att.CheckOutTime)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Belum ada data absen sama sekali.")
			return helpers.NOT_CHECKED_IN
		}
		log.Println("Error GetLastAttendance:", err.Error())
		return "error"
	}

	now := time.Now()
	today := now.Format("2006-01-02")

	checkInDate := att.CheckInTime.Format("2006-01-02")

	// Check-in bukan hari ini
	if checkInDate != today {
		return helpers.NOT_CHECKED_IN
	}

	// Status kehadiran berdasarkan hari ini
	if att.CheckOutTime.Valid {
		return helpers.CHECKED_OUT
	} else {
		return helpers.CHECKED_IN
	}
}

func (model AttendanceModel) CountDaysPresent(nik string, month time.Month, year int) (int, error) {
	query := `
		SELECT COUNT(*) FROM attendance 
		WHERE nik = ? 
		  AND MONTH(checkin_time) = ? 
		  AND YEAR(checkin_time) = ? 
		  AND checkout_time IS NOT NULL
	`
	var count int
	err := model.db.QueryRow(query, nik, int(month), year).Scan(&count)
	return count, err
}


func (model AttendanceModel) GetAttendanceList(nik string, monthYear string, todayOnly bool) ([]entities.Attendance, error) {
	var query string
	var args []interface{}

	parsedDate, err := time.Parse("January 2006", monthYear)
	if err != nil {
		return nil, err
	}

	baseQuery := `
        SELECT 
			a.id, a.nik, a.office_id, a.shift_id,
			a.checkin_time, a.checkin_latitude, a.checkin_longitude,
			a.checkin_photo, a.checkin_notes, a.checkout_time,
			a.checkout_latitude, a.checkout_longitude, a.checkout_photo,
			a.checkout_notes, a.is_late, a.is_early,
			o.name AS office_name, 
			e.name AS employee_name,
			e.uuid AS employee_uuid,
			s.name AS shift_name, s.start_time, s.end_time
		FROM attendance a
		LEFT JOIN office o ON a.office_id = o.id
		LEFT JOIN employee e ON a.nik = e.nik
		LEFT JOIN shift s ON a.shift_id = s.id
		WHERE a.deleted_at IS NULL
		AND MONTH(a.checkin_time) = ?
		AND YEAR(a.checkin_time) = ?

    `

	args = []interface{}{parsedDate.Month(), parsedDate.Year()}

	if todayOnly {
		baseQuery += " AND DATE(a.checkin_time) = CURDATE()"
	}
	if nik != "" {
		baseQuery += " AND a.nik = ?"
		args = append(args, nik)
	}
	query = baseQuery + " ORDER BY a.checkin_time DESC"

	rows, err := model.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attendances []entities.Attendance
	for rows.Next() {
		var att entities.Attendance
		err := rows.Scan(
			&att.ID, &att.NIK, &att.OfficeID, &att.ShiftID,
			&att.CheckInTime, &att.CheckInLatitude, &att.CheckInLongitude,
			&att.CheckInPhoto, &att.CheckInNotes, &att.CheckOutTime,
			&att.CheckOutLatitude, &att.CheckOutLongitude, &att.CheckOutPhoto,
			&att.CheckOutNotes, &att.IsLate, &att.IsEarly,
			&att.OfficeName, &att.EmployeeName, &att.EmployeeUUID,
			&att.ShiftName, &att.ShiftStartTime, &att.ShiftEndTime,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		att.FormattedDate = formatTanggalIndonesia(att.CheckInTime)
		attendances = append(attendances, att)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return attendances, nil
	
}

func (model AttendanceModel) GetAttendanceCounts(nik string, monthYear string) (int, int, error) {
	parsedDate, err := time.Parse("January 2006", monthYear)
	if err != nil {
		return 0, 0, err
	}

	// 1. Total semua kehadiran (checkin & checkout tidak null)
	queryTotal := `
		SELECT COUNT(*) 
		FROM attendance 
		WHERE deleted_at IS NULL 
			AND checkin_time IS NOT NULL 
			AND checkout_time IS NOT NULL
	`
	// 2. Total bulan ini
	queryThisMonth := `
		SELECT COUNT(*) 
		FROM attendance 
		WHERE deleted_at IS NULL 
			AND checkin_time IS NOT NULL 
			AND checkout_time IS NOT NULL
			AND MONTH(checkin_time) = ? 
			AND YEAR(checkin_time) = ?
	`

	var args []interface{}
	argsMonth := []interface{}{parsedDate.Month(), parsedDate.Year()}

	if nik != "" {
		queryTotal += " AND nik = ?"
		queryThisMonth += " AND nik = ?"
		args = append(args, nik)
		argsMonth = append(argsMonth, nik)
	}

	var totalAll int
	err = model.db.QueryRow(queryTotal, args...).Scan(&totalAll)
	if err != nil {
		return 0, 0, err
	}

	var totalThisMonth int
	err = model.db.QueryRow(queryThisMonth, argsMonth...).Scan(&totalThisMonth)
	if err != nil {
		return 0, 0, err
	}

	return totalAll, totalThisMonth, nil
}


func (model AttendanceModel) CountAllAttendance() (int, error) {
	row := model.db.QueryRow("SELECT COUNT(*) FROM attendance WHERE deleted_at IS NULL")

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func formatTanggalIndonesia(t time.Time) string {
	// Format: "2 January 2006" untuk menampilkan "18 Mei 2025"
	return monday.Format(t, "2 January 2006", monday.LocaleIdID)
}

func (model AttendanceModel) GetLatestOfficeAndShift(nik string) (int64, int64, error) {
	var officeID int64
	var shiftID int64

	row := model.db.QueryRow(`
		SELECT office_id, shift_id 
		FROM attendance 
		WHERE nik = ? AND deleted_at IS NULL 
		ORDER BY checkin_time DESC 
		LIMIT 1`, nik)

	err := row.Scan(&officeID, &shiftID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, nil // data tidak ditemukan
		}
		return 0, 0, err // error lainnya
	}

	return officeID, shiftID, nil
}

func (model AttendanceModel) CheckIn(att entities.CheckIn) error {
	_, err := model.db.Exec(`
		INSERT INTO attendance
		(nik, office_id, shift_id, checkin_time, checkin_latitude, checkin_longitude, checkin_photo, is_late, checkin_notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		att.NIK,
		att.OfficeID,
		att.ShiftID,
		att.Time,
		att.Latitude,
		att.Longitude,
		att.Photo,
		att.IsLate,
		att.Notes,
	)
	return err
}

func (model AttendanceModel) CheckOut(nik string, att entities.CheckOut) error {
	_, err := model.db.Exec(`
		UPDATE attendance 
		SET checkout_time = ?, checkout_latitude = ?, checkout_longitude = ?, 
			checkout_photo = ?, checkout_notes = ?, is_early = ?, updated_at = ?
		WHERE nik = ? AND DATE(checkin_time) = CURDATE() AND deleted_at IS NULL`,
		att.Time,
		att.Latitude,
		att.Longitude,
		att.Photo,
		att.Notes,
		att.IsEarly,
		time.Now(),
		nik,
	)
	return err
}