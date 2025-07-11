package controllers

import (
	"fmt"
	"hris/config"
	"hris/entities"
	"hris/helpers"
	"hris/models"
	"hris/services/sessiondata"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goodsign/monday"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func Profile(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/user/pages-profile.html",
	))

	data := make(map[string]interface{})
	session, _ := config.Store.Get(request, config.SESSION_ID)
	errSession := sessiondata.SetUserSessionData(request, data)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}
	data["currentPath"] = request.URL.Path

	if request.Method == http.MethodPost && request.FormValue("change_password") == "1" {
		if result, err := ChangePassword(request, session); err != nil {
			for key, value := range result {
				data[key] = value
			}
		} else {
			data["tab"] = "password"
			data["success"] = ("Berhasil mengubah password.")
			templateLayout.ExecuteTemplate(httpWriter, "base", data)
		}
	}

	if request.Method == http.MethodPost && request.FormValue("edit-profile") == "1" {
		if result, err := EditProfile(request, session); err != nil {
			for key, value := range result {
				data[key] = value
			}
		} else {
			data["tab"] = "profile"
			data["success"] = "Berhasil mengubah profile."
	
			sessionNIK := session.Values["nik"].(string)
			user, err := models.NewUserModel().FindUserByNIK(sessionNIK)
			if err != nil {
				data["error"] = "Gagal mengambil data profile setelah update: " + err.Error()
			} else {
				data["user"] = user

				errSession := sessiondata.SetUserSessionData(request, data)
				if errSession != nil {
					log.Println("SetUserSessionData error:", errSession.Error())
				}

				
				timeLayout := "2006-01-02"
				birthDateTime, err := time.Parse(timeLayout, user.BirthDate)
				if err != nil {
					data["birthDateFormat"] = "-"
				} else {
					data["birthDateFormat"] = monday.Format(birthDateTime, "02 Januari 2006", monday.LocaleIdID)
				}
			}
	
			templateLayout.ExecuteTemplate(httpWriter, "base", data)
			return
		}
	}
	

	sessionNIK := session.Values["nik"].(string)
	user, err := models.NewUserModel().FindUserByNIK(sessionNIK)
	if err != nil {
		data["error"] = "Gagal mengambil data profile: " + err.Error()
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}
	data["user"] = user

	timeLayout := "2006-01-02"
	birthDateTime, err := time.Parse(timeLayout, user.BirthDate)
	if err != nil {
		data["birthDateFormat"] = "-"
	} else {
		data["birthDateFormat"] = monday.Format(birthDateTime, "02 Januari 2006", monday.LocaleIdID)
	}

	if data["tab"] == nil {
		data["tab"] = "profile"
	}

	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}


func EditProfile(request *http.Request, session *sessions.Session) (map[string]interface{}, error) {
	request.ParseMultipartForm(5 << 20) // max 5MB form size

	userInput := entities.EditProfile{
		Name:      request.FormValue("name"),
		Email:     request.FormValue("email"),
		Phone:     request.FormValue("phone"),
		Gender:    request.FormValue("gender"),
		BirthDate: request.FormValue("birth_date"),
		Address:   request.FormValue("address"),
	}

	errors := make(map[string]interface{})
	validationResult := helpers.NewValidation().Struct(userInput)
	if validationResult != nil {
		errors = validationResult.(map[string]interface{})
		return map[string]interface{}{
			"validation": errors,
			"userInput":  userInput,
			"tab":        "profile",
		}, fmt.Errorf("validation error")
	}

	// Handle file photo
	file, handler, err := request.FormFile("photo")
	if err == nil {
		defer file.Close()

		// Validasi ekstensi dan size
		if handler.Size > 2*1024*1024 {
			return map[string]interface{}{
				"error": "Ukuran file maksimal 2MB",
				"userInput":  userInput,
				"tab":        "profile",
			}, fmt.Errorf("size validation error")
		}

		ext := strings.ToLower(filepath.Ext(handler.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			return map[string]interface{}{
				"error": "Tipe file harus jpg, jpeg, png, atau webp",
				"userInput":  userInput,
				"tab":        "profile",
			}, fmt.Errorf("type validation error")
		}

		// Simpan file
		nik := session.Values["nik"].(string)
		oldPhoto, err := models.NewUserModel().GetPhotoByNIK(nik)
		if err != nil {
			return map[string]interface{}{
				"error": "Gagal mendapatkan data Photo.",
				"tab":   "profile",
			}, err
		}
		
		filename := fmt.Sprintf("user_%d%s", time.Now().UnixNano(), ext)
		path := filepath.Join("public/images/user_photo", filename)

		// Hapus foto lama jika ada
		if oldPhoto.Valid && oldPhoto.String != "" {
			oldPath := filepath.Join("public/images/user_photo", oldPhoto.String)
			if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
				return map[string]interface{}{
					"error": "Gagal menghapus foto lama: " + err.Error(),
					"tab":   "profile",
				}, fmt.Errorf("gagal menghapus foto lama")
			}
		}

		out, err := os.Create(path)
		if err != nil {
			return map[string]interface{}{
				"error": "Gagal menyimpan foto: " + err.Error(),
				"userInput":  userInput,
				"tab":        "profile",
			}, fmt.Errorf("gagal menyimpan foto")
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			return map[string]interface{}{
				"error": "Gagal menyimpan file",
				"userInput":  userInput,
				"tab":        "profile",
			}, fmt.Errorf("gagal menyimpan file")
		}

		// Set nama file ke model
		userInput.Photo = filename
	}

	nik := session.Values["nik"].(string)
	err = models.NewUserModel().EditProfile(nik, userInput)
	if err != nil {
		return map[string]interface{}{
			"error": "Gagal mengubah data profile: " + err.Error(),
			"userInput":  userInput,
			"tab":        "profile",
		}, err
	}

	return nil, nil
}



func ChangePassword(request *http.Request, session *sessions.Session) (map[string]interface{}, error) {
	request.ParseForm()
	form := entities.ChangePassword{
		OldPassword:    request.Form.Get("old_password"),
		NewPassword:    request.Form.Get("new_password"),
		RepeatPassword: request.Form.Get("repeat_password"),
	}

	errors := make(map[string]interface{})
	validationResult := helpers.NewValidation().Struct(form)
	if validationResult != nil {
		errors = validationResult.(map[string]interface{})
		return map[string]interface{}{
			"validation": errors,
			"pass":       form,
			"tab":        "password",
		}, fmt.Errorf("validation error")
	}

	nik := session.Values["nik"].(string)
	oldHashedPassword, err := models.NewUserModel().GetPasswordByNIK(nik)
	if err != nil {
		return map[string]interface{}{
			"error": "Gagal mendapatkan data password.",
			"tab":   "password",
		}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(oldHashedPassword), []byte(form.OldPassword)); err != nil {
		errors["OldPassword"] = "Password lama tidak sesuai"
		return map[string]interface{}{
			"validation": errors,
			"pass":       form,
			"tab":        "password",
		}, fmt.Errorf("invalid old password")
	}

	err = models.NewUserModel().ChangePassword(nik, form.NewPassword)
	if err != nil {
		return map[string]interface{}{
			"error": "Gagal mengubah password: " + err.Error(),
			"tab":   "password",
		}, err
	}

	return nil, nil
}
