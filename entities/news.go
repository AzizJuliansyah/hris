package entities

import (
	"database/sql"
	"time"
)

type News struct {
	Id					int64
	Creator_NIK			string
	Creator_Name		string
	Updated_Creator_NIK	string
	Assigne_NIK			sql.NullString
	Assigne_Name		string
	Thumbnail			sql.NullString
	Title				string    `validate:"required" label:"Judul Berita"`
	Content				string    `validate:"required" label:"Content Berita"`
	Footer				string    `validate:"required" label:"Penutup Berita"`
	Start_Date			sql.NullTime
	End_Date			sql.NullTime
	Created_at			time.Time
	Created_atFormat	string
}