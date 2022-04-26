package handlers

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/vleukhin/prom-light/internal/metrics"
	"github.com/vleukhin/prom-light/internal/storage"
)

//go:embed templates
var templates embed.FS

type HomeHandler struct {
	store storage.MetricsGetter
}

func NewHomeHandler(storage storage.MetricsGetter) HomeHandler {
	return HomeHandler{
		store: storage,
	}
}

func (h HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFS(templates, "templates/home.gohtml")
	if err != nil {
		fmt.Println("Template not found: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", "text/html")
	data, err := h.store.GetAllMetrics(r.Context(), false)
	if err != nil {
		log.Println("Failed to get metrics: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	viewData := struct {
		Metrics []metrics.Metric
	}{Metrics: data}

	if err := tpl.Execute(w, viewData); err != nil {
		log.Println("Failed to execute template: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
