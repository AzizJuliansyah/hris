package models

import (
	"database/sql"
	"hris/config"
	"hris/entities"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	db *sql.DB
}

func NewUserModel() *UserModel {
	conn, err := config.DBConnection()
	if err != nil {
		log.Println("Failed connect to database: ", err)
	}
	return &UserModel{
		db: conn,
	}
}

func (model UserModel) FindUserByNIK(nik string) (entities.User, error) {
	var user entities.User
	var birthDateTime time.Time
	var photo sql.NullString

	query := "SELECT nik, name, email, phone, address, gender, birth_date, is_admin, photo FROM employee WHERE nik = ? AND deleted_at IS NULL"

	err := model.db.QueryRow(query, nik).Scan(
		&user.NIK,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.Address,
		&user.Gender,
		&birthDateTime,
		&user.IsAdmin,
		&photo,
	)
	
	if err != nil {
		return user, err
	}
	
	user.BirthDate = birthDateTime.Format("2006-01-02")
	if photo.Valid {
		user.Photo = photo
	} else {
		user.Photo = sql.NullString{String: "", Valid: false}
	}

	return user, nil
}


func (model UserModel) EditProfile(nik string, user entities.EditProfile) error {
	query := `
		UPDATE employee
		SET name = ?, email = ?, phone = ?, address = ?, gender = ?, birth_date = ?, updated_at = ?
		` + func() string {
			if user.Photo != "" {
				return `, photo = ? `
			}
			return ""
		}() + `
		WHERE nik = ? AND deleted_at IS NULL
	`

	args := []interface{}{
		user.Name,
		user.Email,
		user.Phone,
		user.Address,
		user.Gender,
		user.BirthDate,
		time.Now(),
	}
	if user.Photo != "" {
		args = append(args, user.Photo)
	}
	args = append(args, nik)

	_, err := model.db.Exec(query, args...)
	return err
}



func (model UserModel) ChangePassword(nik, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = model.db.Exec(`
		UPDATE employee
		SET password = ?, updated_at = ?
		WHERE nik = ? AND deleted_at IS NULL
	`, string(hashedPassword), time.Now(), nik)

	return err
}

func (model UserModel) GetPhotoByNIK(nik string) (sql.NullString, error) {
	var photo sql.NullString
	err := model.db.QueryRow(`
		SELECT photo FROM employee WHERE nik = ? AND deleted_at IS NULL
	`, nik).Scan(&photo)

	return photo, err
}

func (model UserModel) GetPasswordByNIK(nik string) (string, error) {
	var password string
	err := model.db.QueryRow(`
		SELECT password FROM employee WHERE nik = ? AND deleted_at IS NULL
	`, nik).Scan(&password)

	return password, err
}