package entities

import "database/sql"

type Employee struct {
	UUID      string
	Name      string `validate:"required" label:"Nama"`
	Email     string `validate:"required,email,isunique=employee-email"`
	Phone     string `validate:"required,number,gte=10" label:"No. Handphone"`
	Address   string `validate:"required" label:"Alamat"`
	NIK       string `validate:"required,isunique=employee-nik"`
	Gender    string `validate:"required,oneof=F M" label:"Jenis Kelamin"`
	BirthDate string `validate:"required" label:"Tanggal Lahir"`
	BirthDateFormat string 
	Photo     sql.NullString
	Password  string
	IsAdmin   bool
}

type EditEmployee struct {
	UUID      string
	Name      string `validate:"required" label:"Nama"`
	Email     string `validate:"required,email"`
	Phone     string `validate:"required,number,gte=10" label:"No. Handphone"`
	Address   string `validate:"required" label:"Alamat"`
	NIK       string `validate:"required"`
	Gender    string `validate:"required,oneof=F M" label:"Jenis Kelamin"`
	BirthDate string `validate:"required" label:"Tanggal Lahir"`
	IsAdmin   bool
}