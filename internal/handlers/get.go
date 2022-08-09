package handlers

import (
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/vleukhin/prom-light/internal/metrics"
	"github.com/vleukhin/prom-light/internal/storage"
)

// GetMetricHandler хэнддер для получения метрик
type GetMetricHandler struct {
	store storage.MetricsGetter
}

// GetMetricJSONHandler хэнддер для получения метрик в формате JSON
type GetMetricJSONHandler struct {
	store  storage.MetricsGetter
	hasher hash.Hash
}

// NewGetMetricHandler создает новый хэнддер для получения метрик
func NewGetMetricHandler(storage storage.MetricsGetter) GetMetricHandler {
	return GetMetricHandler{
		store: storage,
	}
}

// NewGetMetricJSONHandler создаем новый хэнддер для получения метрик в формате JSON
func NewGetMetricJSONHandler(storage storage.MetricsGetter, hasher hash.Hash) GetMetricJSONHandler {
	return GetMetricJSONHandler{
		store:  storage,
		hasher: hasher,
	}
}

func (h GetMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	w.Header().Add("Content-type", "text/html")
	switch params["type"] {
	case metrics.GaugeTypeName:
		value, err := h.store.GetGauge(r.Context(), params["name"])
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
		value, err := h.store.GetCounter(r.Context(), params["name"])
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
	var m metrics.Metric

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Debug().Msg("GET JSON metrics: " + string(body))
	err = json.Unmarshal(body, &m)
	if err != nil {
		log.Error().Msg("Failed to parse JSON: " + err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch m.Type {
	case metrics.GaugeTypeName:
		value, err := h.store.GetGauge(r.Context(), m.Name)
		if err != nil {
			log.Error().Msg(err.Error())
			w.WriteHeader(http.StatusNotFound)
			return
		}
		m.Value = &value

	case metrics.CounterTypeName:
		value, err := h.store.GetCounter(r.Context(), m.Name)
		if err != nil {
			log.Error().Msg(err.Error())
			w.WriteHeader(http.StatusNotFound)
			return
		}
		m.Delta = &value
	}

	m.Sign(h.hasher)
	respBody, err := json.Marshal(m)
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
