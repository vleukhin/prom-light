package handlers

import "github.com/vleukhin/prom-light/internal"

type AllMetrics struct {
	GaugeMetrics   map[string]internal.Gauge
	CounterMetrics map[string]internal.Counter
}

type MetricsStorage interface {
	SetGauge(metricName string, value internal.Gauge)
	SetCounter(metricName string, value internal.Counter)
	GetGauge(metricName string) (internal.Gauge, error)
	GetCounter(metricName string) (internal.Counter, error)
	GetAllMetrics() AllMetrics
}
