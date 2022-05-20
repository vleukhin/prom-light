package storage

import (
	"context"

	"github.com/vleukhin/prom-light/internal/metrics"
)

type AllMetrics struct {
	GaugeMetrics   map[string]metrics.Gauge   `json:"gauge_metrics"`
	CounterMetrics map[string]metrics.Counter `json:"counter_metrics"`
}

type MetricsStorage interface {
	MetricsGetter
	MetricsSetter
	Ping(ctx context.Context) error
	ShutDown(ctx context.Context) error
}

type MetricsGetter interface {
	GetGauge(ctx context.Context, metricName string) (metrics.Gauge, error)
	GetCounter(ctx context.Context, metricName string) (metrics.Counter, error)
	GetAllMetrics(ctx context.Context) (metrics.Metrics, error)
}

type MetricsSetter interface {
	SetMetrics(ctx context.Context, mtrcs metrics.Metrics) error
	SetMetric(ctx context.Context, m metrics.Metric) error
}
