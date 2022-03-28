package storage

import (
	"github.com/vleukhin/prom-light/internal/metrics"
)

type AllMetrics struct {
	GaugeMetrics   map[string]metrics.Gauge
	CounterMetrics map[string]metrics.Counter
}

type MetricsStorage interface {
	MetricsGetter
	MetricsSetter
}

type MetricsGetter interface {
	GetGauge(metricName string) (metrics.Gauge, error)
	GetCounter(metricName string) (metrics.Counter, error)
	GetAllMetrics() AllMetrics
}

type MetricsSetter interface {
	SetGauge(metricName string, value metrics.Gauge)
	SetCounter(metricName string, value metrics.Counter)
}
