package controllers

import (
	"database/sql"
	"hris/config"
	"hris/entities"
	"hris/helpers"
	"hris/models"
	"hris/services/sessiondata"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type OfficeController struct {
	db *sql.DB
}

func NewOfficeController(db *sql.DB) *OfficeController {
	return &OfficeController{db: db}
}

func (controller *OfficeController) Office(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/office/office.html",
	))

	if request.Method == http.MethodGet {

		var data = make(map[string]interface{})

		officeModel := models.NewOfficeModel(controller.db)
		offices, err := officeModel.FindAllOffice()

		if err != nil {
			data["error"] = "Terdapat kesahalan saat menampilkan data kantor " + err.Error()
			log.Println("error :", err.Error())
		} else {
			data["office"] = offices
		}

		errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
		if errSession != nil {
			log.Println("SetUserSessionData error:", errSession.Error())
		}
		
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}
}

func (controller *OfficeController) AddOffice(httpWriter http.ResponseWriter, request *http.Request) {

	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/office/add-office.html",
	))

	var data = make(map[string]interface{})
	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}
	
	if request.Method == http.MethodGet {
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	request.ParseForm()


	// Parsing radius
	radiusStr := request.Form.Get("radius")
	radius, _ := strconv.ParseInt(radiusStr, 10, 64)

	// Parsing latitude
	latitudeStr := request.Form.Get("latitude")
	latitude, _ := strconv.ParseFloat(latitudeStr, 64)

	// Parsing longitude
	longitudeStr := request.Form.Get("longitude")
	longitude, _ := strconv.ParseFloat(longitudeStr, 64)

	office := entities.Office{
		Name:      request.Form.Get("name"),
		Address:   request.Form.Get("address"),
		Latitude:  latitude,
		Longitude: longitude,
		Radius:    radius,
	}

	errorMessages := helpers.NewValidation().Struct(office)

	if errorMessages != nil {
		data["validation"] = errorMessages
		data["office"] = office
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	officeModel := models.NewOfficeModel(controller.db)
	err := officeModel.AddOffice(office)

	if err != nil {
		data["error"] = "Gagal menambahkan kantor: " + err.Error()
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
	} else {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Berhasil menambahkan kantor.", "success")
		session.Save(request, httpWriter)
		http.Redirect(httpWriter, request, "/office", http.StatusSeeOther)
	}

}

func (controller *OfficeController) EditOffice(httpWriter http.ResponseWriter, request *http.Request) {

	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/office/edit-office.html",
	))
	id := request.URL.Query().Get("id")
	int64Id, _ := strconv.ParseInt(id, 10, 64)

	var data = make(map[string]interface{})
	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	if request.Method == http.MethodGet {
		// Ambil data office berdasarkan ID
		officeModel := models.NewOfficeModel(controller.db)
		office, err := officeModel.FindOfficeByID(int64Id)
		if err != nil || id == "" {
			session, _ := config.Store.Get(request, config.SESSION_ID)
			session.AddFlash("Gagal mendapatkan data office! " + err.Error(), "error")
			session.Save(request, httpWriter)
		
			http.Redirect(httpWriter, request, "/office", http.StatusSeeOther)
		}

		// Kirim data ke template
		data["office"] = office
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	request.ParseForm()

	// Parsing radius
	radiusStr := request.Form.Get("radius")
	radius, _ := strconv.ParseInt(radiusStr, 10, 64)

	// Parsing latitude
	latitudeStr := request.Form.Get("latitude")
	latitude, _ := strconv.ParseFloat(latitudeStr, 64)

	// Parsing longitude
	longitudeStr := request.Form.Get("longitude")
	longitude, _ := strconv.ParseFloat(longitudeStr, 64)

	office := entities.Office{
		Id:        int64Id,
		Name:      request.Form.Get("name"),
		Address:   request.Form.Get("address"),
		Latitude:  latitude,
		Longitude: longitude,
		Radius:    radius,
	}

	errorMessages := helpers.NewValidation().Struct(office)

	if errorMessages != nil {
		data["validation"] = errorMessages
		data["office"] = office
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	officeModel := models.NewOfficeModel(controller.db)
	err := officeModel.EditOffice(office)

	if err != nil {
		data["error"] = "Edit data gagal: " + err.Error()
	} else {
		updatedOffice, errFind := officeModel.FindOfficeByID(int64Id)
		if errFind != nil {
			data["error"] = "Data berhasil diubah, tapi gagal menampilkan data terbaru: " + errFind.Error()
		} else {
			data["office"] = updatedOffice
			data["success"] = "Berhasil mengubah data kantor."
		}
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)

}

func (controller *OfficeController) DeleteOffice(httpWriter http.ResponseWriter, request *http.Request) {

	id := request.URL.Query().Get("id")
	int64Id, _ := strconv.ParseInt(id, 10, 64)

	if id == "" {
		http.Error(httpWriter, "ID tidak ditemukan", http.StatusBadRequest)
		return
	}

	officeModel := models.NewOfficeModel(controller.db)
	err := officeModel.SoftDeleteOffice(int64Id)
	if err != nil {
		http.Error(httpWriter, "Gagal menghapus data: "+err.Error(), http.StatusInternalServerError)
		return
	} else {
		session, _ := config.Store.Get(request, config.SESSION_ID)
		session.AddFlash("Berhasil menghapus kantor!", "success")
		session.Save(request, httpWriter)
	}

	http.Redirect(httpWriter, request, "/office", http.StatusSeeOther)
}