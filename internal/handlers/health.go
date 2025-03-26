package handlers

import (
	"html/template"
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/auth.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	err2 := tmpl.Execute(w, nil)
	if err2 != nil {
		return
	}
}
