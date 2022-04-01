package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/vleukhin/prom-light/cmd/server/storage"
	"github.com/vleukhin/prom-light/internal/metrics"
)

type GetMetricHandler struct {
	store storage.MetricsGetter
}

type GetMetricJSONHandler struct {
	store storage.MetricsGetter
}

func NewGetMetricHandler(storage storage.MetricsGetter) GetMetricHandler {
	return GetMetricHandler{
		store: storage,
	}
}

func NewGetMetricJSONHandler(storage storage.MetricsGetter) GetMetricJSONHandler {
	return GetMetricJSONHandler{
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

func (h GetMetricJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var reqMetrics, respMetrics metrics.Metrics
	err := json.NewDecoder(r.Body).Decode(&reqMetrics)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	for _, m := range reqMetrics {
		switch m.Type {
		case metrics.GaugeTypeName:
			value, err := h.store.GetGauge(m.Name)
			if err != nil {
				continue
			}
			respMetrics = append(respMetrics, metrics.Metric{
				Name:  m.Name,
				Type:  m.Type,
				Value: &value,
			})

		case metrics.CounterTypeName:
			value, err := h.store.GetCounter(m.Name)
			if err != nil {
				continue
			}

			respMetrics = append(respMetrics, metrics.Metric{
				Name:  m.Name,
				Type:  m.Type,
				Delta: &value,
			})
		}
	}

	respBody, err := json.Marshal(respMetrics)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(respBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
