package handlers

import (
	"encoding/json"
	"fmt"
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
		err = h.store.SetGauge(r.Context(), params["name"], metrics.Gauge(value))
		if err != nil {
			log.Println(fmt.Sprintf("Failed to set gauge %s: %s", params["name"], err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case metrics.CounterTypeName:
		value, err := strconv.ParseInt(params["value"], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = h.store.IncCounter(r.Context(), params["name"], metrics.Counter(value))
		if err != nil {
			log.Println(fmt.Sprintf("Failed to inc counter %s: %s", params["name"], err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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
		if m.Value == nil {
			log.Println("Invalid gauge value")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err := h.store.SetGauge(r.Context(), m.Name, *m.Value)
		if err != nil {
			log.Println(fmt.Sprintf("Failed to set gauge %s: %s", m.Name, err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	case metrics.CounterTypeName:
		if m.Delta == nil {
			log.Println("Invalid counter value")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err := h.store.IncCounter(r.Context(), m.Name, *m.Delta)
		if err != nil {
			log.Println(fmt.Sprintf("Failed to inc counter %s: %s", m.Name, err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
