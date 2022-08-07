package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/vleukhin/prom-light/internal/metrics"
)

type DatabaseStorage struct {
	conn *pgxpool.Pool
}

func NewDatabaseStorage(dsn string, connTimeout time.Duration) (*DatabaseStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
	defer cancel()

	conn, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return &DatabaseStorage{
		conn: conn,
	}, nil
}

// language=PostgreSQL
const getMetricSQL = `SELECT value FROM metrics WHERE name = $1`

func (s *DatabaseStorage) GetGauge(ctx context.Context, metricName string) (metrics.Gauge, error) {
	var value float64

	row := s.conn.QueryRow(ctx, getMetricSQL, metricName)
	err := row.Scan(&value)
	if err != nil {
		return 0, err
	}

	return metrics.Gauge(value), nil
}

func (s *DatabaseStorage) GetCounter(ctx context.Context, metricName string) (metrics.Counter, error) {
	var value int

	row := s.conn.QueryRow(ctx, getMetricSQL, metricName)
	err := row.Scan(&value)
	if err != nil {
		return 0, err
	}

	return metrics.Counter(value), nil
}

// language=PostgreSQL
const getAllMetricsSQL = `SELECT name, type, value  FROM metrics order by id`

func (s *DatabaseStorage) GetAllMetrics(ctx context.Context) (metrics.Metrics, error) {
	rows, err := s.conn.Query(ctx, getAllMetricsSQL)
	if err != nil {
		return nil, err
	}

	var result metrics.Metrics

	for rows.Next() {
		metric := metrics.Metric{}
		var rawValue float64
		err := rows.Scan(&metric.Name, &metric.Type, &rawValue)
		if err != nil {
			return nil, err
		}

		switch metric.Type {
		case metrics.GaugeTypeName:
			value := metrics.Gauge(rawValue)
			metric.Value = &value
		case metrics.CounterTypeName:
			delta := metrics.Counter(rawValue)
			metric.Delta = &delta
		default:
			return nil, errors.New("unknown metric type: " + metric.Type)
		}

		result = append(result, metric)
	}

	return result, nil
}

// language=PostgreSQL
const setGaugeSQL = `
	INSERT INTO metrics (name, type, value)
	VALUES ($1, $2, $3)
	ON CONFLICT ON CONSTRAINT metrics_name_key DO UPDATE
	SET value = excluded.value
`

func (s *DatabaseStorage) SetMetric(ctx context.Context, m metrics.Metric) error {
	var err error
	switch m.Type {
	case metrics.GaugeTypeName:
		if m.Value == nil {
			return errors.New("nil gauge value")
		}
		_, err = s.conn.Exec(ctx, setGaugeSQL, m.Name, metrics.GaugeTypeName, *m.Value)
	case metrics.CounterTypeName:
		if m.Delta == nil {
			return errors.New("nil counter value")
		}
		_, err = s.conn.Exec(ctx, incCounterSQL, m.Name, metrics.CounterTypeName, *m.Delta)
	}

	return err
}
func (s *DatabaseStorage) SetMetrics(ctx context.Context, mtrcs metrics.Metrics) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}

	for _, m := range mtrcs {
		if err := s.SetMetric(ctx, m); err != nil {
			txErr := tx.Rollback(ctx)
			if txErr != nil {
				return txErr
			}
			return err
		}
	}

	return tx.Commit(ctx)
}

// language=PostgreSQL
const incCounterSQL = `
	INSERT INTO metrics (name, type, value)
	VALUES ($1, $2, $3)
	ON CONFLICT ON CONSTRAINT metrics_name_key DO UPDATE
	SET value = metrics.value + excluded.value
`

func (s *DatabaseStorage) IncCounter(ctx context.Context, metricName string, value metrics.Counter) error {
	_, err := s.conn.Exec(ctx, incCounterSQL, metricName, metrics.CounterTypeName, value)
	return err
}

func (s *DatabaseStorage) ShutDown(_ context.Context) error {
	s.conn.Close()
	return nil
}

func (s *DatabaseStorage) Ping(ctx context.Context) error {
	return s.conn.Ping(ctx)
}

// language=PostgreSQL
const createMetricsTable = `
	CREATE TABLE IF NOT EXISTS metrics (
		id    serial constraint table_name_pk primary key,
		name  varchar(255) not null unique,
		type  varchar(255) not null,
		value float8       not null
	)
`

func (s *DatabaseStorage) Migrate(ctx context.Context) error {
	_, err := s.conn.Exec(ctx, createMetricsTable)
	return err
}
