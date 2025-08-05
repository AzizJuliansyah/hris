package entities

import (
	"database/sql"
	"time"
)

type KPI struct {
	ID             int
	Employee_NIK   string
	Admin_NIK      string
	Title          string
	Description    string
	Target_Value   int
	Unit           string
	Status         string
	Deadline       string
	DeadlineTimeFormat string
	Started_At     string
	Completed_At   string
	Actual_Value   string
	Evidence_File  sql.NullString
	Final_Status   string
	Score          string
	Admin_Notes    string
	Checked_At     string
	Created_At     string
	IsCreatedOver1Day bool
	Updated_At     string
	Admin_Name     string
	Employee_Name  string
}


type AddKPI struct {
	ID           int64
	Employee_NIK string `validate:"required" label:"NIK Karyawan"`
	Admin_NIK 	 string 
	Title        string `validate:"required,min=2,max=255" label:"Judul"`
	Description  string `validate:"required,min=2" label:"Deskripsi"`
	Target_Value int64 `validate:"required,gte=1" label:"Target Dicapai"`
	Unit 		 string `validate:"required" label:"satuan"`
	Deadline 	 time.Time `validate:"required" label:"Tenggat Waktu"`
}

type EditKPI struct {
	ID           	  int64
	Employee_NIK 	  string `validate:"required" label:"NIK Karyawan"`
	Updated_Admin_NIK string 
	Title        	  string `validate:"required,min=2,max=255" label:"Judul"`
	Description  	  string `validate:"required,min=2" label:"Deskripsi"`
	Target_Value 	  int64 `validate:"required,gte=1" label:"Target Dicapai"`
	Unit 		 	  string `validate:"required" label:"satuan"`
	Deadline 	 	  time.Time `validate:"required" label:"Tenggat Waktu"`
	Status			  string
}

type EvaluasiKPI struct {
	ID 				  int64
	Checked_Admin_NIK string
	Final_Status 	  string `validate:"required" label:"status"`
	Score 			  float64 `validate:"required,min=0" label:"status"`
	Admin_Notes 	  string
}

type SubmitProgressKPI struct {
	ID				int64
	Actual_Value	int		`validate:"required" label:"Progress"`
	Evidence_File	string
}