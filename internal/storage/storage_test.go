package storage

import (
	"context"
	"github.com/vleukhin/prom-light/internal/metrics"
	"reflect"
	"testing"
)

func testStorage(storage MetricsStorage, t *testing.T) {
	var metricsData = metrics.Metrics{
		metrics.MakeGaugeMetric("Gauge1", 5.5),
		metrics.MakeGaugeMetric("Gauge2", 0),
		metrics.MakeGaugeMetric("Gauge3", -8),
		metrics.MakeCounterMetric("Counter1", 0),
		metrics.MakeCounterMetric("Counter2", 12312),
		metrics.MakeCounterMetric("Counter3", 4444),
	}

	ctx := context.Background()
	t.Run("Set metrics one by one", func(t *testing.T) {
		for _, m := range metricsData {
			err := storage.SetMetric(ctx, m)
			if err != nil {
				t.Errorf("SetMetric() error = %v", err)
				return
			}

			if m.IsCounter() {
				stored, err := storage.GetCounter(ctx, m.Name)
				if err != nil {
					t.Errorf("GetCounter() error = %v", err)
					return
				}
				if *m.Delta != stored {
					t.Errorf("GetCounter() wrong value = %v; want %v", stored, *m.Delta)
					return
				}
			} else {
				stored, err := storage.GetGauge(ctx, m.Name)
				if err != nil {
					t.Errorf("GetGauge() error = %v", err)
					return
				}
				if *m.Value != stored {
					t.Errorf("GetGauge() wrong value = %v; want %v", stored, *m.Value)
					return
				}
			}
		}
	})

	_ = storage.CleanUp(ctx)

	t.Run("Set metrics batch", func(t *testing.T) {
		err := storage.SetMetrics(context.Background(), metricsData)
		if err != nil {
			t.Errorf("SetMetrics() error = %v", err)
			return
		}
		stored, err := storage.GetAllMetrics(ctx)
		if err != nil {
			t.Errorf("GetAllMetrics() error = %v", err)
			return
		}

		if !reflect.DeepEqual(stored, metricsData) {
			t.Errorf("matching error\ngot: %v\nwant:%v\n", stored, metricsData)
		}
	})

	_ = storage.CleanUp(ctx)

	t.Run("Inc counter", func(t *testing.T) {
		name := "counter"
		if err := storage.IncCounter(ctx, name, 10); err != nil {
			t.Errorf("IncCounter() error = %v", err)
		}
		if err := storage.IncCounter(ctx, name, 15); err != nil {
			t.Errorf("IncCounter() error = %v", err)
		}

		stored, err := storage.GetCounter(ctx, name)
		if err != nil {
			t.Errorf("GetCounter() error = %v", err)
		}

		if stored != 25 {
			t.Errorf("GetCounter() wrong value = %v; want %v", stored, 25)
			return
		}
	})
}
