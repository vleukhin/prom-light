package storage

import (
	"errors"

	"github.com/vleukhin/prom-light/cmd/server/handlers"
	"github.com/vleukhin/prom-light/internal"
)

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

func (m MemoryStorage) GetGauge(name string) (internal.Gauge, error) {
	value, exists := m.gaugeMetrics[name]
	if !exists {
		return 0, errors.New("unknown gauge")
	}

	return value, nil
}

func (m MemoryStorage) GetCounter(name string) (internal.Counter, error) {
	value, exists := m.counterMetrics[name]
	if !exists {
		return 0, errors.New("unknown counter")
	}

	return value, nil
}

func (m MemoryStorage) GetAllMetrics() handlers.AllMetrics {
	return handlers.AllMetrics{
		GaugeMetrics:   m.gaugeMetrics,
		CounterMetrics: m.counterMetrics,
	}
}
