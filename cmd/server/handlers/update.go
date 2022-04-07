package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
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
		log.Printf("Received gauge %s with value %.3f \n", params["name"], value)
		h.store.SetGauge(params["name"], metrics.Gauge(value))
	case metrics.CounterTypeName:
		value, err := strconv.ParseInt(params["value"], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Received counter %s with value %d \n", params["name"], value)
		h.store.IncCounter(params["name"], metrics.Counter(value))
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
	var m metrics.Metric

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("UPDATE JSON metrics: " + string(body))
	err = json.Unmarshal(body, &m)
	if err != nil {
		log.Println("Failed to parse JSON: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch m.Type {
	case metrics.GaugeTypeName:
		if m.Value != nil {
			h.store.SetGauge(m.Name, *m.Value)
		}

	case metrics.CounterTypeName:
		if m.Delta != nil {
			h.store.IncCounter(m.Name, *m.Delta)
		}
	}
}
