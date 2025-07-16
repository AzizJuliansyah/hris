package sessiondata

import (
	"database/sql"
	"hris/config"
	"hris/models"
	"net/http"
)

func SetUserSessionData(httpWriter http.ResponseWriter, request *http.Request, data map[string]interface{}, db *sql.DB) error {
	session, _ := config.Store.Get(request, config.SESSION_ID)

	nik, ok := session.Values["nik"].(string)
	if !ok {
		data["error"] = "NIK tidak ditemukan di sesi"
		return nil
	}

	data["sessionNIK"] = session.Values["nik"].(string)
	data["name"] = session.Values["name"]
	data["isAdmin"] = session.Values["isAdmin"]

	userModel := models.NewUserModel(db)
	user, err := userModel.FindUserByNIK(nik)
	if err != nil {
		data["error"] = "Gagal mengambil data profile: " + err.Error()
		return err
	}

	if user.Photo.Valid && user.Photo.String != "" {
		data["photoPath"] = "/images/user_photo/" + user.Photo.String
	} else {
		data["photoPath"] = "/images/user_default.png"
	}
	

	if flashes := session.Flashes("success"); len(flashes) > 0 {
		data["success"] = flashes[0]
		session.Save(request, httpWriter)
	}
	if flashes := session.Flashes("error"); len(flashes) > 0 {
		data["error"] = flashes[0]
		session.Save(request, httpWriter)
	}

	return nil
}
