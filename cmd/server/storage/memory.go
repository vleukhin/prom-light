package storage

import "github.com/vleukhin/prom-light/internal"

type MemoryStorage struct {
	gaugeMetrics   map[string]internal.Gauge
	counterMetrics map[string]internal.Counter
}

func NewMemoryStorage() MemoryStorage {
	return MemoryStorage{
		gaugeMetrics:   make(map[string]internal.Gauge),
		counterMetrics: make(map[string]internal.Counter),
	}
}

func (m MemoryStorage) StoreGauge(metricName string, value internal.Gauge) {
	m.gaugeMetrics[metricName] = value
}
func (m MemoryStorage) StoreCounter(metricName string, value internal.Counter) {
	oldValue, ok := m.counterMetrics[metricName]
	if !ok {
		oldValue = 0
	}

	m.counterMetrics[metricName] = oldValue + value
}
