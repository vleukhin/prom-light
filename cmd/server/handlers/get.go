package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/vleukhin/prom-light/cmd/server/storage"
	"github.com/vleukhin/prom-light/internal/metrics"
)

type GetMetricHandler struct {
	store storage.MetricsGetter
}

func NewGetMetricHandler(storage storage.MetricsGetter) GetMetricHandler {
	return GetMetricHandler{
		store: storage,
	}
}

func (h GetMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	w.Header().Add("Content-type", "text/html")
	switch params["type"] {
	case metrics.GaugeTypeName:
		value, err := h.store.GetGauge(params["name"])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		_, err = w.Write([]byte(fmt.Sprintf("%.3f", value)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	case metrics.CounterTypeName:
		value, err := h.store.GetCounter(params["name"])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, err = w.Write([]byte(fmt.Sprintf("%d", value)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
}
