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

func (s *memoryStorage) SetGauge(_ context.Context, metricName string, value metrics.Gauge) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.gaugeMetrics[metricName] = value
	return nil
}
func (s *memoryStorage) IncCounter(_ context.Context, metricName string, value metrics.Counter) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	oldValue, ok := s.counterMetrics[metricName]
	if !ok {
		oldValue = 0
	}
	s.counterMetrics[metricName] = oldValue + value
	return nil
}

func (s *memoryStorage) SetMetric(_ context.Context, m metrics.Metric) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	switch m.Type {
	case metrics.GaugeTypeName:
		if m.Value == nil {
			return errors.New("nil gauge value")
		}
		s.gaugeMetrics[m.Name] = *m.Value
	case metrics.CounterTypeName:
		if m.Delta == nil {
			return errors.New("nil counter value")
		}
		oldValue, ok := s.counterMetrics[m.Name]
		if !ok {
			oldValue = 0
		}
		s.counterMetrics[m.Name] = oldValue + *m.Delta
	}

	return nil
}

func (s *memoryStorage) SetMetrics(ctx context.Context, mtrcs metrics.Metrics) error {
	for _, m := range mtrcs {
		if err := s.SetMetric(ctx, m); err != nil {
			return err
		}
	}

	return nil
}

func (s *memoryStorage) GetGauge(_ context.Context, metricName string) (metrics.Gauge, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	value, exists := s.gaugeMetrics[metricName]
	if !exists {
		return 0, errors.New("unknown gauge")
	}

	return value, nil
}

func (s *memoryStorage) GetCounter(_ context.Context, name string) (metrics.Counter, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	value, exists := s.counterMetrics[name]
	if !exists {
		return 0, errors.New("unknown counter")
	}

	return value, nil
}

func (s *memoryStorage) GetAllMetrics(_ context.Context, resetCounters bool) (metrics.Metrics, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	result := make(metrics.Metrics, len(s.gaugeMetrics)+len(s.counterMetrics))
	var i int
	for k, v := range s.gaugeMetrics {
		value := v
		result[i] = metrics.Metric{
			Name:  k,
			Type:  metrics.GaugeTypeName,
			Value: &value,
		}
		i++
	}
	for k, v := range s.counterMetrics {
		value := v
		result[i] = metrics.Metric{
			Name:  k,
			Type:  metrics.CounterTypeName,
			Delta: &value,
		}
		if resetCounters {
			s.counterMetrics[k] = 0
		}
		i++
	}

	return result, nil
}

func (s *memoryStorage) ShutDown(_ context.Context) error {
	// nothing to do here
	return nil
}

func (s *memoryStorage) Ping(_ context.Context) error {
	// nothing to do here
	return nil
}
