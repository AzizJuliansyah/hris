package controllers

import (
	"database/sql"
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
	"strconv"
	"strings"
	"time"
)

type NewsController struct {
	db *sql.DB
}

func NewNewsController(db *sql.DB) *NewsController {
	return &NewsController{db: db}
}

func (controller *NewsController) ListNews(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/news/news.html",
	))
	data := make(map[string]interface{})

	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}

	newsModel := models.NewNewsModel(controller.db)
	newss, err := newsModel.FindAllNews()
	if err != nil {
		data["error"] = "Terdapat kesahalan saat menampilkan data news " + err.Error()
		log.Println("error :", err.Error())
	} else {
		data["news"] = newss
	}

	if request.Method == http.MethodGet {
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}
}

func (controller *NewsController) AddNews(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/news/add-news.html",
	))

	data := make(map[string]interface{})
	session, _ := config.Store.Get(request, config.SESSION_ID)
	sessionNIK := session.Values["nik"].(string)
	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}
	data["news"] = entities.News{} // atasi no value

	salaryModel := models.NewSalaryModel(controller.db)
	employees, err := salaryModel.GetEmployeeNameandNIK()
	if err != nil {
		log.Println("Error Getting Employee NIK and Name", err)
		return
	}
	data["employees"] = employees

	if request.Method == http.MethodPost {
		request.ParseMultipartForm(5 << 20)
		assigneNIKInput := request.Form.Get("assigne_nik")

		startDateInput := request.Form.Get("start_date")
		endDateInput := request.Form.Get("end_date")
		startDate, _ := time.Parse("2006-01-02", startDateInput)
		endDate, _ := time.Parse("2006-01-02", endDateInput)
		news := entities.News{
			Creator_NIK: sessionNIK,
			Assigne_NIK: sql.NullString{
				String: assigneNIKInput,
				Valid:  assigneNIKInput != "",
			},
			Title:       request.Form.Get("title"),
			Content:     request.Form.Get("content"),
			Footer:      request.Form.Get("footer"),
			Start_Date: sql.NullTime{
				Time:  startDate,
				Valid: startDateInput != "",
			},
			End_Date: sql.NullTime{
				Time:  endDate,
				Valid: endDateInput != "",
			},
		}

		// Validasi
		errorMessages := helpers.NewValidation().Struct(news)
		if errorMessages != nil {
			data["validation"] = errorMessages
			data["news"] = news
			data["currentPath"] = request.URL.Path
			templateLayout.ExecuteTemplate(httpWriter, "base", data)
			return
		}

		// Upload thumbnail jika ada
		file, handler, err := request.FormFile("thumbnail")
		if err == nil {
			defer file.Close()

			// Validasi ukuran dan tipe
			if handler.Size > 2*1024*1024 {
				data["error"] = "Ukuran file maksimal 2MB"
				data["news"] = news
				templateLayout.ExecuteTemplate(httpWriter, "base", data)
				return
			}

			ext := strings.ToLower(filepath.Ext(handler.Filename))
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
				data["error"] = "Tipe file harus jpg, jpeg, png, atau webp"
				data["news"] = news
				templateLayout.ExecuteTemplate(httpWriter, "base", data)
				return
			}

			// Simpan file ke public/images/user_photo
			filename := fmt.Sprintf("news_%d%s", time.Now().UnixNano(), ext)
			path := filepath.Join("public/images/news_thumbnail", filename)

			out, err := os.Create(path)
			if err != nil {
				data["error"] = "Gagal menyimpan foto: " + err.Error()
				data["news"] = news
				templateLayout.ExecuteTemplate(httpWriter, "base", data)
				return
			}
			defer out.Close()
			_, err = io.Copy(out, file)
			if err != nil {
				data["error"] = "Gagal menyimpan file"
				data["news"] = news
				templateLayout.ExecuteTemplate(httpWriter, "base", data)
				return
			}

			news.Thumbnail = sql.NullString{String: filename, Valid: true}
		}

		// Simpan ke DB
		newsModel := models.NewNewsModel(controller.db)
		err = newsModel.AddNews(news)
		if err != nil {
			data["error"] = "Gagal menambahkan berita: " + err.Error()
		} else {
			session.AddFlash("Berhasil menambahkan berita baru.", "success")
			session.Save(request, httpWriter)
			http.Redirect(httpWriter, request, "/news", http.StatusSeeOther)
		}
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func (controller *NewsController) EditNews(httpWriter http.ResponseWriter, request *http.Request) {
	templateLayout := template.Must(template.ParseFiles(
		"views/static/layouts/base.html",
		"views/static/layouts/header.html",
		"views/static/layouts/navbar.html",
		"views/static/layouts/sidebar.html",
		"views/static/layouts/footer.html",
		"views/static/layouts/footer_js.html",
		"views/static/news/edit-news.html",
	))

	data := make(map[string]interface{})
	session, _ := config.Store.Get(request, config.SESSION_ID)
	sessionNIK := session.Values["nik"].(string)
	
	errSession := sessiondata.SetUserSessionData(httpWriter, request, data, controller.db)
	if errSession != nil {
		log.Println("SetUserSessionData error:", errSession.Error())
	}
	id := request.URL.Query().Get("id")
	int64Id, _ := strconv.ParseInt(id, 10, 64)

	data["news"] = entities.News{} // atasi no value

	salaryModel := models.NewSalaryModel(controller.db)
	employees, err := salaryModel.GetEmployeeNameandNIK()
	if err != nil {
		log.Println("Error Getting Employee NIK and Name", err)
		return
	}
	data["employees"] = employees

	if request.Method == http.MethodGet {

		newsModel := models.NewNewsModel(controller.db)
		news, err := newsModel.FindNewsByID(int64Id)
		if err != nil || id == "" {
			session, _ := config.Store.Get(request, config.SESSION_ID)
			session.AddFlash("Gagal mendapatkan berita! " + err.Error(), "error")
			session.Save(request, httpWriter)

			http.Redirect(httpWriter, request, "/news", http.StatusSeeOther)
		}
		
		data["news"] = news
		data["currentPath"] = request.URL.Path
		templateLayout.ExecuteTemplate(httpWriter, "base", data)
		return
	}

	if request.Method == http.MethodPost {
		request.ParseMultipartForm(5 << 20)
		assigneNIKInput := request.Form.Get("assigne_nik")

		startDateInput := request.Form.Get("start_date")
		endDateInput := request.Form.Get("end_date")
		startDate, _ := time.Parse("2006-01-02", startDateInput)
		endDate, _ := time.Parse("2006-01-02", endDateInput)
		news := entities.News{
			Creator_NIK: sessionNIK,
			Assigne_NIK: sql.NullString{
				String: assigneNIKInput,
				Valid:  assigneNIKInput != "",
			},
			Title:       request.Form.Get("title"),
			Content:     request.Form.Get("content"),
			Footer:      request.Form.Get("footer"),
			Start_Date: sql.NullTime{
				Time:  startDate,
				Valid: startDateInput != "",
			},
			End_Date: sql.NullTime{
				Time:  endDate,
				Valid: endDateInput != "",
			},
			Id: int64Id,
		}

		// Validasi
		errorMessages := helpers.NewValidation().Struct(news)
		if errorMessages != nil {
			data["validation"] = errorMessages
			data["news"] = news
			data["currentPath"] = request.URL.Path
			templateLayout.ExecuteTemplate(httpWriter, "base", data)
			return
		}

		// Upload thumbnail jika ada
		file, handler, err := request.FormFile("thumbnail")
		if err == nil {
			defer file.Close()

			// Validasi ukuran dan tipe
			if handler.Size > 2*1024*1024 {
				data["error"] = "Ukuran file maksimal 2MB"
				data["news"] = news
				templateLayout.ExecuteTemplate(httpWriter, "base", data)
				return
			}

			ext := strings.ToLower(filepath.Ext(handler.Filename))
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
				data["error"] = "Tipe file harus jpg, jpeg, png, atau webp"
				data["news"] = news
				templateLayout.ExecuteTemplate(httpWriter, "base", data)
				return
			}

			newsModel := models.NewNewsModel(controller.db)
			oldPhoto, err := newsModel.GetThumbnailByID(int64Id)
			if err != nil {
				data["error"] = "Gagal mendapatkan data Photo."
				data["news"] = news
				templateLayout.ExecuteTemplate(httpWriter, "base", data)
				return
			}

			// Simpan file ke public/images/user_photo
			filename := fmt.Sprintf("news_%d%s", time.Now().UnixNano(), ext)
			path := filepath.Join("public/images/news_thumbnail", filename)

			if oldPhoto.Valid && oldPhoto.String != "" {
				oldPath := filepath.Join("public/images/news_thumbnail", oldPhoto.String)
				if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
					data["error"] = "Gagal menghapus foto lama: " + err.Error()
					data["news"] = news
					templateLayout.ExecuteTemplate(httpWriter, "base", data)
					return
				}
			}

			out, err := os.Create(path)
			if err != nil {
				data["error"] = "Gagal menyimpan foto: " + err.Error()
				data["news"] = news
				templateLayout.ExecuteTemplate(httpWriter, "base", data)
				return
			}
			defer out.Close()
			_, err = io.Copy(out, file)
			if err != nil {
				data["error"] = "Gagal menyimpan file"
				data["news"] = news
				templateLayout.ExecuteTemplate(httpWriter, "base", data)
				return
			}

			news.Thumbnail = sql.NullString{String: filename, Valid: true}
		} else {
			// Tidak ada file baru: pakai thumbnail lama
			newsModel := models.NewNewsModel(controller.db)
			oldPhoto, err := newsModel.GetThumbnailByID(int64Id)
			if err == nil {
				news.Thumbnail = oldPhoto
			}
		}

		// Simpan ke DB
		newsModel := models.NewNewsModel(controller.db)
		err = newsModel.EditNews(news)
		if err != nil {
			data["error"] = "Gagal mengubah berita: " + err.Error()
		} else {
			session.AddFlash("Berhasil mengubah berita.", "success")
			session.Save(request, httpWriter)
			http.Redirect(httpWriter, request, "/news/edit-news?id="+id, http.StatusSeeOther)
		}
	}

	data["currentPath"] = request.URL.Path
	templateLayout.ExecuteTemplate(httpWriter, "base", data)
}

func (controller *NewsController) DeleteNews(httpWriter http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("id")
	if id == "" {
		http.Error(httpWriter, "ID tidak ditemukan", http.StatusBadRequest)
		return
	}

	int64Id, _ := strconv.ParseInt(id, 10, 64)

	// Ambil nama thumbnail
	// thumbnail, err := newsModel.GetThumbnailByID(int64Id)
	// if err == nil && thumbnail.Valid && thumbnail.String != "" {
	// 	// Hapus file dari folder
	// 	path := filepath.Join("public/images/news_thumbnail", thumbnail.String)
	// 	if _, err := os.Stat(path); err == nil {
	// 		os.Remove(path)
	// 	}
	// }

	// Soft delete
	newsModel := models.NewNewsModel(controller.db)
	err := newsModel.SoftDeleteNews(int64Id)
	if err != nil {
		http.Error(httpWriter, "Gagal menghapus data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	session, _ := config.Store.Get(request, config.SESSION_ID)
	session.AddFlash("Berhasil menghapus berita!", "success")
	session.Save(request, httpWriter)

	http.Redirect(httpWriter, request, "/news", http.StatusSeeOther)
}
