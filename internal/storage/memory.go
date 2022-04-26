package storage

import (
	"context"
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

func (m *memoryStorage) SetGauge(_ context.Context, metricName string, value metrics.Gauge) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.gaugeMetrics[metricName] = value
	return nil
}
func (m *memoryStorage) IncCounter(_ context.Context, metricName string, value metrics.Counter) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	oldValue, ok := m.counterMetrics[metricName]
	if !ok {
		oldValue = 0
	}
	m.counterMetrics[metricName] = oldValue + value
	return nil
}

func (m *memoryStorage) GetGauge(_ context.Context, metricName string) (metrics.Gauge, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	value, exists := m.gaugeMetrics[metricName]
	if !exists {
		return 0, errors.New("unknown gauge")
	}

	return value, nil
}

func (m *memoryStorage) GetCounter(_ context.Context, name string) (metrics.Counter, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	value, exists := m.counterMetrics[name]
	if !exists {
		return 0, errors.New("unknown counter")
	}

	return value, nil
}

func (m *memoryStorage) GetAllMetrics(_ context.Context, resetCounters bool) ([]metrics.Metric, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	result := make([]metrics.Metric, len(m.gaugeMetrics)+len(m.counterMetrics))
	var i int
	for k, v := range m.gaugeMetrics {
		value := v
		result[i] = metrics.Metric{
			Name:  k,
			Type:  metrics.GaugeTypeName,
			Value: &value,
		}
		i++
	}
	for k, v := range m.counterMetrics {
		value := v
		result[i] = metrics.Metric{
			Name:  k,
			Type:  metrics.CounterTypeName,
			Delta: &value,
		}
		if resetCounters {
			m.counterMetrics[k] = 0
		}
		i++
	}

	return result, nil
}

func (m *memoryStorage) ShutDown(_ context.Context) error {
	// nothing to do here
	return nil
}

func (m *memoryStorage) Ping(_ context.Context) error {
	// nothing to do here
	return nil
}
