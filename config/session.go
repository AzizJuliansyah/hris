package config

import "github.com/gorilla/sessions"

const SESSION_ID = "hris-app"

var Store *sessions.CookieStore

func init() {
	Store = sessions.NewCookieStore([]byte("hris-app"))
	Store.Options = &sessions.Options{
		Path: "/",
		MaxAge: 3600 * 8,
		HttpOnly: true,
		Secure: true,
	}
}