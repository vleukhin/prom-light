package handlers

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/vleukhin/prom-light/internal"
	"net/http"
)

type GetMetricHandler struct {
	storage MetricsStorage
}

func NewGetMetricHandler(storage MetricsStorage) GetMetricHandler {
	return GetMetricHandler{
		storage: storage,
	}
}

func (h GetMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	w.Header().Add("Content-type", "text/html")
	switch internal.MetricTypeName(params["type"]) {
	case internal.GaugeTypeName:
		value, err := h.storage.GetGauge(params["name"])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		_, err = w.Write([]byte(fmt.Sprintf("%.3f", value)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	case internal.CounterTypeName:
		value, err := h.storage.GetCounter(params["name"])
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
