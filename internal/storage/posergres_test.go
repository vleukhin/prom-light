package storage

import (
	"context"
	"github.com/vleukhin/prom-light/internal/metrics"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
)

var (
	db *PostgresStorage
)

type testConfig struct {
	DSN string `env:"DATABASE_DSN_TEST" envDefault:"postgres://postgres:postgres@localhost:5454/tests?sslmode=disable"`
}

func TestMain(m *testing.M) {
	var err error
	ctx := context.Background()
	cfg := testConfig{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatal().Err(err)
	}

	db, err = NewPostgresStorage(cfg.DSN, time.Second*5)
	if err != nil {
		log.Fatal().Err(err)
	}

	if err := db.Migrate(ctx); err != nil {
		log.Fatal().Err(err)
	}
	exCode := m.Run()
	if err := db.CleanUp(ctx); err != nil {
		panic(err)
	}
	os.Exit(exCode)
}

func TestPostgresStorage(t *testing.T) {
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
			err := db.SetMetric(ctx, m)
			if err != nil {
				t.Errorf("SetMetric() error = %v", err)
				return
			}

			if m.IsCounter() {
				stored, err := db.GetCounter(ctx, m.Name)
				if err != nil {
					t.Errorf("GetCounter() error = %v", err)
					return
				}
				if *m.Delta != stored {
					t.Errorf("GetCounter() wrong value = %v; want %v", stored, *m.Delta)
					return
				}
			} else {
				stored, err := db.GetGauge(ctx, m.Name)
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

	_ = db.CleanUp(ctx)

	t.Run("Set metrics batch", func(t *testing.T) {
		err := db.SetMetrics(context.Background(), metricsData)
		if err != nil {
			t.Errorf("SetMetrics() error = %v", err)
			return
		}
		stored, err := db.GetAllMetrics(ctx)
		if err != nil {
			t.Errorf("GetAllMetrics() error = %v", err)
			return
		}

		if !reflect.DeepEqual(stored, metricsData) {
			t.Errorf("matching error\ngot: %v\nwant:%v\n", stored, metricsData)
		}
	})

	_ = db.CleanUp(ctx)

	t.Run("Inc counter", func(t *testing.T) {
		name := "counter"
		if err := db.IncCounter(ctx, name, 10); err != nil {
			t.Errorf("IncCounter() error = %v", err)
		}
		if err := db.IncCounter(ctx, name, 15); err != nil {
			t.Errorf("IncCounter() error = %v", err)
		}

		stored, err := db.GetCounter(ctx, name)
		if err != nil {
			t.Errorf("GetCounter() error = %v", err)
		}

		if stored != 25 {
			t.Errorf("GetCounter() wrong value = %v; want %v", stored, 25)
			return
		}
	})
}
