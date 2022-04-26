package handlers

import (
	"encoding/json"
	"hash"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/vleukhin/prom-light/internal/metrics"
	"github.com/vleukhin/prom-light/internal/storage"
)

type UpdateMetricHandler struct {
	store storage.MetricsSetter
}

type UpdateMetricJSONHandler struct {
	store  storage.MetricsSetter
	hasher hash.Hash
}

func NewUpdateMetricHandler(storage storage.MetricsSetter) UpdateMetricHandler {
	return UpdateMetricHandler{
		store: storage,
	}
}

func NewUpdateMetricJSONHandler(storage storage.MetricsSetter, hasher hash.Hash) UpdateMetricJSONHandler {
	return UpdateMetricJSONHandler{
		store:  storage,
		hasher: hasher,
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
		h.store.SetGauge(r.Context(), params["name"], metrics.Gauge(value))
	case metrics.CounterTypeName:
		value, err := strconv.ParseInt(params["value"], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Received counter %s with value %d \n", params["name"], value)
		h.store.IncCounter(r.Context(), params["name"], metrics.Counter(value))
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

	if !m.IsValid(h.hasher) {
		log.Println("Invalid hash in metric " + m.Name)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch m.Type {
	case metrics.GaugeTypeName:
		if m.Value != nil {
			h.store.SetGauge(r.Context(), m.Name, *m.Value)
		}

	case metrics.CounterTypeName:
		if m.Delta != nil {
			h.store.IncCounter(r.Context(), m.Name, *m.Delta)
		}
	}
}
