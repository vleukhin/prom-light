package storage

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vleukhin/prom-light/internal"
	"testing"
)

type MockStorage struct {
	t              *testing.T
	gaugeMetrics   map[string]internal.Gauge
	counterMetrics map[string]internal.Counter
}

func NewMockStorage(t *testing.T) MockStorage {
	return MockStorage{
		gaugeMetrics:   make(map[string]internal.Gauge),
		counterMetrics: make(map[string]internal.Counter),
		t:              t,
	}
}

func (s MockStorage) StoreGauge(name string, value internal.Gauge) {
	s.gaugeMetrics[name] = value
}
func (s MockStorage) StoreCounter(name string, value internal.Counter) {
	s.counterMetrics[name] = value
}

func (s MockStorage) AssertGaugeStoredWithValue(name string, expected internal.Gauge) {
	actual, ok := s.gaugeMetrics[name]
	assert.True(s.t, ok, fmt.Sprintf("Gauge '%s' was not stored", name))
	assert.Equal(s.t, expected, actual, fmt.Sprintf("Gauge '%s' was stored with wrong value. Expected: %f Actual: %f", name, expected, actual))
}

func (s MockStorage) AssertCounterStoredWithValue(name string, expected internal.Counter) {
	actual, ok := s.counterMetrics[name]
	assert.True(s.t, ok, fmt.Sprintf("Counter '%s' was not stored", name))
	assert.Equal(s.t, expected, actual, fmt.Sprintf("Counter '%s' was stored with wrong value. Expected: %d Actual: %d", name, expected, actual))
}
