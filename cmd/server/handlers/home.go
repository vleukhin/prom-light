package handlers

import (
	"html/template"
	"net/http"
	"path"

	"github.com/vleukhin/prom-light/cmd/server/storage"
)

type HomeHandler struct {
	store storage.MetricsGetter
}

func NewHomeHandler(storage storage.MetricsGetter) HomeHandler {
	return HomeHandler{
		store: storage,
	}
}

func (h HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fp := path.Join("templates", "home.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, h.store.GetAllMetrics()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
