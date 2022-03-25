package storage

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vleukhin/prom-light/cmd/server/handlers"
	"github.com/vleukhin/prom-light/internal"
	"testing"
)

type MockStorage struct {
	gaugeMetrics   map[string]internal.Gauge
	counterMetrics map[string]internal.Counter
}

func NewMockStorage() MockStorage {
	return MockStorage{
		gaugeMetrics:   make(map[string]internal.Gauge),
		counterMetrics: make(map[string]internal.Counter),
	}
}

func (s MockStorage) StoreGauge(name string, value internal.Gauge) {
	s.gaugeMetrics[name] = value
}
func (s MockStorage) StoreCounter(name string, value internal.Counter) {
	s.counterMetrics[name] = value
}

func (s MockStorage) AssertGaugeStoredWithValue(t *testing.T, name string, expected internal.Gauge) {
	actual, ok := s.gaugeMetrics[name]
	assert.True(t, ok, fmt.Sprintf("Gauge '%s' was not stored", name))
	assert.Equal(t, expected, actual, fmt.Sprintf("Gauge '%s' was stored with wrong value. Expected: %f Actual: %f", name, expected, actual))
}

func (s MockStorage) AssertCounterStoredWithValue(t *testing.T, name string, expected internal.Counter) {
	actual, ok := s.counterMetrics[name]
	assert.True(t, ok, fmt.Sprintf("Counter '%s' was not stored", name))
	assert.Equal(t, expected, actual, fmt.Sprintf("Counter '%s' was stored with wrong value. Expected: %d Actual: %d", name, expected, actual))
}

func (s MockStorage) GetGauge(name string) (internal.Gauge, error) {
	value, exists := s.gaugeMetrics[name]
	if !exists {
		return 0, errors.New("unknown gauge")
	}

	return value, nil
}

func (s MockStorage) GetCounter(name string) (internal.Counter, error) {
	value, exists := s.counterMetrics[name]
	if !exists {
		return 0, errors.New("unknown counter")
	}
	return value, nil
}

func (s MockStorage) GetAllMetrics() handlers.AllMetrics {
	return handlers.AllMetrics{
		GaugeMetrics:   s.gaugeMetrics,
		CounterMetrics: s.counterMetrics,
	}
}
