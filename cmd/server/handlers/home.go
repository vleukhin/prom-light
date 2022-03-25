package handlers

import (
	"html/template"
	"net/http"
	"path"
)

type HomeHandler struct {
	storage MetricsStorage
}

func NewHomeHandler(storage MetricsStorage) HomeHandler {
	return HomeHandler{
		storage: storage,
	}
}

func (h HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fp := path.Join("templates", "home.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, h.storage.GetAllMetrics()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
