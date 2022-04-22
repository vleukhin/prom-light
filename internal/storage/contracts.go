package storage

import (
	"github.com/vleukhin/prom-light/internal/metrics"
)

type AllMetrics struct {
	GaugeMetrics   map[string]metrics.Gauge   `json:"gauge_metrics"`
	CounterMetrics map[string]metrics.Counter `json:"counter_metrics"`
}

type MetricsStorage interface {
	MetricsGetter
	MetricsSetter
	ShutDown() error
}

type MetricsGetter interface {
	GetGauge(metricName string) (metrics.Gauge, error)
	GetCounter(metricName string) (metrics.Counter, error)
	GetAllMetrics(resetCounters bool) []metrics.Metric
}

type MetricsSetter interface {
	SetGauge(metricName string, value metrics.Gauge)
	IncCounter(metricName string, value metrics.Counter)
}
