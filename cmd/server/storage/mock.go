package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vleukhin/prom-light/internal/metrics"
)

type mockStorage struct {
	*memoryStorage
}

func NewMockStorage() *mockStorage {
	return &mockStorage{NewMemoryStorage()}
}

func (s *mockStorage) AssertGaugeStoredWithValue(t *testing.T, name string, expected metrics.Gauge) {
	actual, ok := s.gaugeMetrics[name]
	assert.True(t, ok, fmt.Sprintf("Gauge '%s' was not stored", name))
	assert.Equal(t, expected, actual, fmt.Sprintf("Gauge '%s' was stored with wrong value. Expected: %f Actual: %f", name, expected, actual))
}

func (s *mockStorage) AssertCounterStoredWithValue(t *testing.T, name string, expected metrics.Counter) {
	actual, ok := s.counterMetrics[name]
	assert.True(t, ok, fmt.Sprintf("Counter '%s' was not stored", name))
	assert.Equal(t, expected, actual, fmt.Sprintf("Counter '%s' was stored with wrong value. Expected: %d Actual: %d", name, expected, actual))
}
