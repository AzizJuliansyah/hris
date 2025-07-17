package controllers

import (
	"database/sql"
	"hris/config"
	"hris/entities"
	"hris/helpers"
	"hris/models"
	"hris/views"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	db *sql.DB
}

func NewAuthController(db *sql.DB) *AuthController {
	return &AuthController{db: db}
}

func (controller *AuthController) Login(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := "views/static/login/login.html"
	data := make(map[string]interface{})
	if request.Method == http.MethodGet {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		if flashes := session.Flashes("success"); len(flashes) > 0 {
			data["success"] = flashes[0]
			session.Save(request, httpWriter)
		}
		data["authInput"] = entities.Auth{}
		views.RenderTemplate(httpWriter, templateLayout, data)
		return
	}

	request.ParseForm()

	authInput := entities.Auth{
		NIK:	  request.Form.Get("nik"),
		Password: request.Form.Get("password"),
	}

	// validasi jika input kosong
	if validationError := helpers.NewValidation().Struct(authInput); validationError != nil {
		data["validation"] = validationError
		views.RenderTemplate(httpWriter, templateLayout, data)
		return
	}

	// cek data employee
	authModel := models.NewAuthModel(controller.db)
	employee, err := authModel.FindEmployeeByNIK(authInput.NIK)
	if err != nil {
		data["error"] = "NIK tidak ditemukan" + err.Error()
		data["authInput"] = authInput
		views.RenderTemplate(httpWriter, templateLayout, data)
		return
	}

	// cek bagian password
	if err := bcrypt.CompareHashAndPassword([]byte(employee.Password), []byte(authInput.Password)); err!= nil {
		data["error"] = "NIK atau Password salah"
		data["authInput"] = authInput
		views.RenderTemplate(httpWriter, templateLayout, data)
		return
	}

	session, _ := config.Store.Get(request, config.SESSION_ID)
	session.Values["loggedIn"] = true
	session.Values["nik"] = employee.NIK
	session.Values["name"] = employee.Name
	session.Values["isAdmin"] = employee.IsAdmin
	session.AddFlash("Berhasil login!", "success")
	session.Save(request, httpWriter)

	if employee.IsAdmin {
		http.Redirect(httpWriter, request, "/home-admin", http.StatusSeeOther)
	} else {
		http.Redirect(httpWriter, request, "/home", http.StatusSeeOther)
	}
}


func Logout(httpWriter http.ResponseWriter, request *http.Request) {
	session, _ := config.Store.Get(request, config.SESSION_ID)

	// kosongkan session
	session.Values = make(map[interface{}]interface{}) 
	// session.Options.MaxAge = -1

	// Tambahkan flash ke session baru
	session.AddFlash("Berhasil logout!", "success")
	session.Save(request, httpWriter)

	http.Redirect(httpWriter, request, "/login", http.StatusSeeOther)
}