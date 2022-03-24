package storage

import "github.com/vleukhin/prom-light/internal"

type MemoryStorage struct {
}

func NewMemoryStorage() MemoryStorage {
	return MemoryStorage{}
}

func (m MemoryStorage) StoreGauge(metricName string, value internal.Gauge) {

}
func (m MemoryStorage) StoreCounter(metricName string, value internal.Counter) {

}
