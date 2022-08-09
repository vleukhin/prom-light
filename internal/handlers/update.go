package handlers

import (
	"encoding/json"
	"hash"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/vleukhin/prom-light/internal/metrics"
	"github.com/vleukhin/prom-light/internal/storage"
)

// UpdateMetricHandler хэндлер для обновления метрики
type UpdateMetricHandler struct {
	store storage.MetricsSetter
}

// UpdateMetricJSONHandler хэндлер для обновления метрики в формате JSON
type UpdateMetricJSONHandler struct {
	store  storage.MetricsSetter
	hasher hash.Hash
}

// UpdateMetricsBatchHandler хэндлер для массового обновления метрик в формате JSON
type UpdateMetricsBatchHandler struct {
	store  storage.MetricsSetter
	hasher hash.Hash
}

//NewUpdateMetricHandler создаёт хэндлер для обновления метрики
func NewUpdateMetricHandler(storage storage.MetricsSetter) UpdateMetricHandler {
	return UpdateMetricHandler{
		store: storage,
	}
}

//NewUpdateMetricJSONHandler создаёт хэндлер для обновления метрики в формате JSON
func NewUpdateMetricJSONHandler(storage storage.MetricsSetter, hasher hash.Hash) UpdateMetricJSONHandler {
	return UpdateMetricJSONHandler{
		store:  storage,
		hasher: hasher,
	}
}

//NewUpdateMetricsBatchHandler создаёт хэндлер для массового обновления метрик в формате JSON
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
		log.Debug().Msgf("Received gauge %s with value %.3f \n", params["name"], rawValue)
		value := metrics.Gauge(rawValue)
		m.Value = &value
	case metrics.CounterTypeName:
		rawValue, err := strconv.ParseInt(params["value"], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug().Msgf("Received counter %s with value %d \n", params["name"], rawValue)
		value := metrics.Counter(rawValue)
		m.Delta = &value
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	err := h.store.SetMetric(r.Context(), m)
	if err != nil {
		log.Error().Msg(err.Error())
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
		log.Error().Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Debug().Msg("UPDATE JSON metrics: " + string(body))

	err = json.Unmarshal(body, &m)
	if err != nil {
		log.Error().Msg("Failed to parse JSON: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !m.IsValid(h.hasher) {
		log.Error().Msg("Invalid hash in metric " + m.Name)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.store.SetMetric(r.Context(), m)
	if err != nil {
		log.Error().Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h UpdateMetricsBatchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var mtrcs metrics.Metrics

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Debug().Msg("UPDATE BATCH JSON metrics: " + string(body))

	err = json.Unmarshal(body, &mtrcs)
	if err != nil {
		log.Error().Msg("Failed to parse JSON: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !mtrcs.IsValid(h.hasher) {
		log.Error().Msg("Invalid hash")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.store.SetMetrics(r.Context(), mtrcs)
	if err != nil {
		log.Error().Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
