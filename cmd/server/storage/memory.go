package storage

import (
	"errors"
	"sync"

	"github.com/vleukhin/prom-light/internal"

	"github.com/vleukhin/prom-light/cmd/server/handlers"
)

type MemoryStorage struct {
	mutex          *sync.Mutex
	gaugeMetrics   map[string]internal.Gauge
	counterMetrics map[string]internal.Counter
}

func NewMemoryStorage() MemoryStorage {
	var mutex sync.Mutex
	return MemoryStorage{
		mutex:          &mutex,
		gaugeMetrics:   make(map[string]internal.Gauge),
		counterMetrics: make(map[string]internal.Counter),
	}
}

func (m MemoryStorage) SetGauge(metricName string, value internal.Gauge) {
	m.mutex.Lock()
	m.gaugeMetrics[metricName] = value
	m.mutex.Unlock()
}
func (m MemoryStorage) SetCounter(metricName string, value internal.Counter) {
	m.mutex.Lock()
	oldValue, ok := m.counterMetrics[metricName]
	if !ok {
		oldValue = 0
	}
	m.counterMetrics[metricName] = oldValue + value
	m.mutex.Unlock()
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
