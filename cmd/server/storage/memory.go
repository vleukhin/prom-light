package storage

import (
	"errors"
	"sync"

	"github.com/vleukhin/prom-light/internal/metrics"
)

type MemoryStorage struct {
	mutex          *sync.Mutex
	gaugeMetrics   map[string]metrics.Gauge
	counterMetrics map[string]metrics.Counter
}

func NewMemoryStorage() MemoryStorage {
	var mutex sync.Mutex
	return MemoryStorage{
		mutex:          &mutex,
		gaugeMetrics:   make(map[string]metrics.Gauge),
		counterMetrics: make(map[string]metrics.Counter),
	}
}

func (m MemoryStorage) SetGauge(metricName string, value metrics.Gauge) {
	m.mutex.Lock()
	defer m.mutex.Lock()
	m.gaugeMetrics[metricName] = value
}
func (m MemoryStorage) SetCounter(metricName string, value metrics.Counter) {
	m.mutex.Lock()
	defer m.mutex.Lock()
	oldValue, ok := m.counterMetrics[metricName]
	if !ok {
		oldValue = 0
	}
	m.counterMetrics[metricName] = oldValue + value
}

func (m MemoryStorage) GetGauge(name string) (metrics.Gauge, error) {
	m.mutex.Lock()
	defer m.mutex.Lock()
	value, exists := m.gaugeMetrics[name]
	if !exists {
		return 0, errors.New("unknown gauge")
	}

	return value, nil
}

func (m MemoryStorage) GetCounter(name string) (metrics.Counter, error) {
	m.mutex.Lock()
	defer m.mutex.Lock()
	value, exists := m.counterMetrics[name]
	if !exists {
		return 0, errors.New("unknown counter")
	}

	return value, nil
}

func (m MemoryStorage) GetAllMetrics() AllMetrics {
	m.mutex.Lock()
	defer m.mutex.Lock()
	return AllMetrics{
		GaugeMetrics:   m.gaugeMetrics,
		CounterMetrics: m.counterMetrics,
	}
}
