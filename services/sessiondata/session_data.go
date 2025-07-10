package sessiondata

import (
	"hris/config"
	"hris/models"
	"net/http"
)

func SetUserSessionData(request *http.Request, data map[string]interface{}) error {
	session, _ := config.Store.Get(request, config.SESSION_ID)

	nik, ok := session.Values["nik"].(string)
	if !ok {
		data["error"] = "NIK tidak ditemukan di sesi"
		return nil
	}

	data["name"] = session.Values["name"]
	data["isAdmin"] = session.Values["isAdmin"]

	user, err := models.NewUserModel().FindUserByNIK(nik)
	if err != nil {
		data["error"] = "Gagal mengambil data profile: " + err.Error()
		return err
	}

	if user.Photo.Valid && user.Photo.String != "" {
		data["photoPath"] = "/images/user_photo/" + user.Photo.String
	} else {
		data["photoPath"] = "/images/user_default.png"
	}

	data["user"] = user // kalau mau sekalian kasih struct user-nya
	return nil
}
