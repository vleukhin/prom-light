package httphandlers

import (
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/vleukhin/prom-light/internal/metrics"

	"github.com/vleukhin/prom-light/internal/storage"
)

type MetricsController struct {
	store  storage.MetricsStorage
	hasher hash.Hash
}

func NewMetricsController(storage storage.MetricsStorage, hasher hash.Hash) MetricsController {
	return MetricsController{
		store:  storage,
		hasher: hasher,
	}
}

func (c MetricsController) UpdateMetric(w http.ResponseWriter, r *http.Request) {
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

	err := c.store.SetMetric(r.Context(), m)
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

func (c MetricsController) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	var m metrics.Metric

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
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

	if !m.IsValid(c.hasher) {
		log.Error().Msg("Invalid hash in metric " + m.Name)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = c.store.SetMetric(r.Context(), m)
	if err != nil {
		log.Error().Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c MetricsController) UpdateMetricsBatch(w http.ResponseWriter, r *http.Request) {
	var mtrcs metrics.Metrics

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
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

	if !mtrcs.IsValid(c.hasher) {
		log.Error().Msg("Invalid hash")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = c.store.SetMetrics(r.Context(), mtrcs)
	if err != nil {
		log.Error().Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c MetricsController) GetMetric(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	w.Header().Add("Content-type", "text/html")
	switch params["type"] {
	case metrics.GaugeTypeName:
		value, err := c.store.GetGauge(r.Context(), params["name"])
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
		value, err := c.store.GetCounter(r.Context(), params["name"])
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

func (c MetricsController) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
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
		value, err := c.store.GetGauge(r.Context(), m.Name)
		if err != nil {
			log.Error().Msg(err.Error())
			w.WriteHeader(http.StatusNotFound)
			return
		}
		m.Value = &value

	case metrics.CounterTypeName:
		value, err := c.store.GetCounter(r.Context(), m.Name)
		if err != nil {
			log.Error().Msg(err.Error())
			w.WriteHeader(http.StatusNotFound)
			return
		}
		m.Delta = &value
	}

	m.Sign(c.hasher)
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
