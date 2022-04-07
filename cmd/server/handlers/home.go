package handlers

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/vleukhin/prom-light/internal/metrics"

	"github.com/vleukhin/prom-light/cmd/server/storage"
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

func (h HomeHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	tpl, err := template.ParseFS(templates, "templates/home.gohtml")
	if err != nil {
		fmt.Println("Template not found: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", "text/html")
	viewData := struct {
		Metrics []metrics.Metric
	}{Metrics: h.store.GetAllMetrics(false)}

	if err := tpl.Execute(w, viewData); err != nil {
		fmt.Println("Failed to execute template: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
