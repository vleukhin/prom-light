package storage

import (
	"errors"
	"sync"

	"github.com/vleukhin/prom-light/internal/metrics"
)

type memoryStorage struct {
	mutex          sync.Mutex
	gaugeMetrics   map[string]metrics.Gauge
	counterMetrics map[string]metrics.Counter
}

func NewMemoryStorage() *memoryStorage {
	return &memoryStorage{
		gaugeMetrics:   make(map[string]metrics.Gauge),
		counterMetrics: make(map[string]metrics.Counter),
	}
}

func (m *memoryStorage) SetGauge(metricName string, value metrics.Gauge) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.gaugeMetrics[metricName] = value
}
func (m *memoryStorage) SetCounter(metricName string, value metrics.Counter) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	oldValue, ok := m.counterMetrics[metricName]
	if !ok {
		oldValue = 0
	}
	m.counterMetrics[metricName] = oldValue + value
}

func (m *memoryStorage) GetGauge(name string) (metrics.Gauge, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	value, exists := m.gaugeMetrics[name]
	if !exists {
		return 0, errors.New("unknown gauge")
	}

	return value, nil
}

func (m *memoryStorage) GetCounter(name string) (metrics.Counter, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	value, exists := m.counterMetrics[name]
	if !exists {
		return 0, errors.New("unknown counter")
	}

	return value, nil
}

func (m *memoryStorage) GetAllMetrics() AllMetrics {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	result := AllMetrics{
		make(map[string]metrics.Gauge),
		make(map[string]metrics.Counter),
	}

	for k, v := range m.gaugeMetrics {
		result.GaugeMetrics[k] = v
	}
	for k, v := range m.counterMetrics {
		result.CounterMetrics[k] = v
	}

	return result
}

func (m *memoryStorage) ShutDown() {
	// nothing to do here
}
