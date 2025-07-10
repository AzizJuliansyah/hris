package views

import (
	"html/template"
	"net/http"
)

func RenderTemplate(httpWriter http.ResponseWriter, path string, data interface{}) {
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		http.Error(httpWriter, "Template Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(httpWriter, data)

}