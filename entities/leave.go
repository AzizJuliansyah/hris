package entities

import (
	"database/sql"
	"time"
)

type LeaveType struct {
	Id     int64
	Name   string `validate:"required" label:"Nama Cuti"`
	MaxDay string `validate:"required,numeric" label:"Maximal Hari Cuti"`
}

type SubmitLeave struct {
	Id            int64
	NIK           string
	LeaveDate     []string `validate:"required,gte=1" label:"Tanggal Cuti"`
	LeaveDateJoin string
	Attachment    *string
	Reason        string `validate:"required" label:"Alasan"`
	LeaveTypeId   string `validate:"required" label:"Tipe Cuti"`
	Status        int64
	ReasonStatus  sql.NullString
}

type Leave struct {
	Id            int64
	LeaveTypeId   int64
	LeaveTypeName string
	NIK           string
	LeaveDateJoin string
	Attachment    sql.NullString
	Reason        string
	Status        int64
	ReasonStatus  sql.NullString
	CreatedAt     time.Time
	UpdatedAt     sql.NullTime
	LeaveDate     []time.Time
	Name          string
}

type ApprovalLeave struct {
	Id           int64
	Status       int64  `validate:"required"`
	ReasonStatus string `validate:"required" label:"Catatan"`
	UpdatedAt    time.Time
}

type LeaveInAMonth struct {
	Id        int64
	NIK       string
	LeaveDate string
	Status    int
}