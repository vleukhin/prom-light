package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/vleukhin/prom-light/cmd/server/storage"
	"github.com/vleukhin/prom-light/internal/metrics"
)

type UpdateMetricHandler struct {
	store storage.MetricsSetter
}

type UpdateMetricJSONHandler struct {
	store storage.MetricsSetter
}

func NewUpdateMetricHandler(storage storage.MetricsSetter) UpdateMetricHandler {
	return UpdateMetricHandler{
		store: storage,
	}
}

func NewUpdateMetricJSONHandler(storage storage.MetricsSetter) UpdateMetricJSONHandler {
	return UpdateMetricJSONHandler{
		store: storage,
	}
}

func (h UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	switch params["type"] {
	case metrics.GaugeTypeName:
		value, err := strconv.ParseFloat(params["value"], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Printf("Received gauge %s with value %.3f \n", params["name"], value)
		h.store.SetGauge(params["name"], metrics.Gauge(value))
	case metrics.CounterTypeName:
		value, err := strconv.ParseInt(params["value"], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Printf("Received counter %s with value %d \n", params["name"], value)
		h.store.SetCounter(params["name"], metrics.Counter(value))
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	_, err := w.Write([]byte("Updated"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h UpdateMetricJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var mtrcs metrics.Metrics
	err := json.NewDecoder(r.Body).Decode(&mtrcs)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, m := range mtrcs {
		switch m.Type {
		case metrics.GaugeTypeName:
			if m.Value != nil {
				h.store.SetGauge(m.Name, *m.Value)
			}

		case metrics.CounterTypeName:
			if m.Delta != nil {
				h.store.SetCounter(m.Name, *m.Delta)
			}
		}
	}
}
