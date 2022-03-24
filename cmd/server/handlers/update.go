package handlers

import (
	"github.com/vleukhin/prom-light/internal"
	"net/http"
)

type UpdateMetricHandler struct {
	storage MetricsStorage
}

type MetricsStorage interface {
	StoreGauge(metricType internal.MetricTypeName, metricName string, value internal.Gauge)
	StoreCounter(metricType internal.MetricTypeName, metricName string, value internal.Counter)
}

func NewUpdateMetricHandler(storage MetricsStorage) UpdateMetricHandler {
	return UpdateMetricHandler{
		storage: storage,
	}
}

func (h UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("UpdateMetricHandler"))
}
