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

type UpdateMetricsBatchHandler struct {
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

func NewUpdateMetricsBatchHandler(storage storage.MetricsSetter, hasher hash.Hash) UpdateMetricsBatchHandler {
	return UpdateMetricsBatchHandler{
		store:  storage,
		hasher: hasher,
	}
}

func (h UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	m := metrics.Metric{
		Name: params["name"],
		Type: params["type"],
	}

	switch params["type"] {
	case metrics.GaugeTypeName:
		rawValue, err := strconv.ParseFloat(params["value"], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		value := metrics.Gauge(rawValue)
		m.Value = &value
	case metrics.CounterTypeName:
		rawValue, err := strconv.ParseInt(params["value"], 10, 64)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		value := metrics.Counter(rawValue)
		m.Delta = &value
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	err := h.store.SetMetric(r.Context(), m)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte("Updated"))
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

	err = h.store.SetMetric(r.Context(), m)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h UpdateMetricsBatchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var mtrcs metrics.Metrics

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &mtrcs)
	if err != nil {
		log.Println("Failed to parse JSON: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !mtrcs.IsValid(h.hasher) {
		log.Println("Invalid hash")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.store.SetMetrics(r.Context(), mtrcs)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
