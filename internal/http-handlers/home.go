package httphandlers

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/vleukhin/prom-light/internal/metrics"
	"github.com/vleukhin/prom-light/internal/storage"
)

//go:embed templates
var templates embed.FS

// HomeHandlerController хэндлер для просмотра метрик
type HomeHandlerController struct {
	store storage.MetricsGetter
}

// NewHomeHandler создаёт новый хэндлер для просмотра метрик
func NewHomeHandler(storage storage.MetricsGetter) HomeHandlerController {
	return HomeHandlerController{
		store: storage,
	}
}

func (h HomeHandlerController) Home(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFS(templates, "templates/home.gohtml")
	if err != nil {
		log.Error().Msg("Template not found: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", "text/html")
	data, err := h.store.GetAllMetrics(r.Context())
	if err != nil {
		log.Error().Msg("Failed to get metrics: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	viewData := struct {
		Metrics []metrics.Metric
	}{Metrics: data}

	if err := tpl.Execute(w, viewData); err != nil {
		log.Error().Msg("Failed to execute template: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
