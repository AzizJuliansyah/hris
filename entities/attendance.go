package entities

import (
	"database/sql"
	"time"
)

type Attendance struct {
	ID                int64
	NIK               string
	OfficeID          int64
	ShiftID           int64
	CheckInTime       time.Time
	CheckInLatitude   float64
	CheckInLongitude  float64
	CheckInPhoto      string
	CheckInNotes      sql.NullString
	CheckOutTime      sql.NullTime
	CheckOutLatitude  sql.NullFloat64
	CheckOutLongitude sql.NullFloat64
	CheckOutPhoto     sql.NullString
	CheckOutNotes     sql.NullString
	IsLate            bool
	IsEarly           sql.NullBool
	OfficeName        string
	FormattedDate     string
	EmployeeName      string
	EmployeeUUID      string
	ShiftName         string
	ShiftStartTime    string
	ShiftEndTime      string
}

type CheckIn struct {
	NIK        string
	OfficeID   string `validate:"required" label:"Kantor"`
	ShiftID    string `validate:"required" label:"Kantor"`
	Time       time.Time
	Latitude   float64
	Longitude  float64
	Notes      string
	IsLate     bool
	Photo      string `validate:"required" label:"Foto"`
	LatLongStr string `validate:"required" label:"Lokasi"`
}

type CheckOut struct {
	NIK        string
	Time       time.Time
	Latitude   float64
	Longitude  float64
	Notes      string
	IsEarly    bool
	Photo      string `validate:"required" label:"Foto"`
	LatLongStr string `validate:"required" label:"Lokasi"`
}