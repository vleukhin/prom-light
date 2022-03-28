package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/vleukhin/prom-light/cmd/server/storage"
	"github.com/vleukhin/prom-light/internal"
)

type UpdateMetricHandler struct {
	store storage.MetricsSetter
}

func NewUpdateMetricHandler(storage storage.MetricsSetter) UpdateMetricHandler {
	return UpdateMetricHandler{
		store: storage,
	}
}

func (h UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	switch internal.MetricTypeName(params["type"]) {
	case internal.GaugeTypeName:
		value, err := strconv.ParseFloat(params["value"], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Printf("Received gauge %s with value %.3f \n", params["name"], value)
		h.store.SetGauge(params["name"], internal.Gauge(value))
	case internal.CounterTypeName:
		value, err := strconv.ParseInt(params["value"], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Printf("Received counter %s with value %d \n", params["name"], value)
		h.store.SetCounter(params["name"], internal.Counter(value))
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	_, err := w.Write([]byte("Updated"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
