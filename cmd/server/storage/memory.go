package storage

import "github.com/vleukhin/prom-light/internal"

type MemoryStorage struct {
}

func NewMemoryStorage() MemoryStorage {
	return MemoryStorage{}
}

func (m MemoryStorage) StoreGauge(metricType internal.MetricTypeName, metricName string, value internal.Gauge) {

}
func (m MemoryStorage) StoreCounter(metricType internal.MetricTypeName, metricName string, value internal.Counter) {

}
