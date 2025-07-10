package config

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func DBConnection() (db *sql.DB, err error) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := ""
	dbName := "hris"

	db, err = sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp(localhost:3306)/"+dbName+"?parseTime=true&loc=Asia%2FJakarta")
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 5)
	return db, err
}