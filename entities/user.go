package entities

import "database/sql"

type User struct {
	UUID      string
	NIK       string `validate:"required"`
	Name      string `validate:"required" label:"Nama"`
	Email     string `validate:"required,email"`
	Phone     string `validate:"required,number,gte=10" label:"No. Handphone"`
	Address   string `validate:"required" label:"Alamat"`
	Gender    string `validate:"required,oneof=F M" label:"Jenis Kelamin"`
	BirthDate string `validate:"required" label:"Tanggal Lahir"`
	IsAdmin   bool
	Photo     sql.NullString
}

type EditProfile struct {
	Name      string `validate:"required" label:"Nama"`
	Email     string `validate:"required,email"`
	Phone     string `validate:"required,number,gte=10" label:"No. Handphone"`
	Address   string `validate:"required" label:"Alamat"`
	Gender    string `validate:"required,oneof=F M" label:"Jenis Kelamin"`
	BirthDate string `validate:"required" label:"Tanggal Lahir"`
	Photo 	  string	
}

type ChangePassword struct {
	OldPassword    string `validate:"required" label:"Password Lama"`
	NewPassword    string `validate:"required,min=5" label:"Password Baru"`
	RepeatPassword string `validate:"required,eqfield=NewPassword" label:"Ulangi Password"`
}