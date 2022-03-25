package handlers

import "github.com/vleukhin/prom-light/internal"

type MetricsStorage interface {
	StoreGauge(metricName string, value internal.Gauge)
	StoreCounter(metricName string, value internal.Counter)
	GetGauge(metricName string) (internal.Gauge, error)
	GetCounter(metricName string) (internal.Counter, error)
}
