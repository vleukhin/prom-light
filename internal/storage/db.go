package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/vleukhin/prom-light/internal/metrics"
)

type DatabaseStorage struct {
	conn *pgx.Conn
}

func NewDatabaseStorage(dsn string, connTimeout time.Duration) (DatabaseStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
	defer cancel()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return DatabaseStorage{}, err
	}

	return DatabaseStorage{
		conn,
	}, nil
}

func (s DatabaseStorage) GetConnection() *pgx.Conn {
	return s.conn
}

// language=PostgreSQL
const getMetricSQL = `SELECT value FROM metrics WHERE name = $1`

func (s DatabaseStorage) GetGauge(ctx context.Context, metricName string) (metrics.Gauge, error) {
	var value float64

	row := s.conn.QueryRow(ctx, getMetricSQL, metricName)
	err := row.Scan(&value)
	if err != nil {
		return 0, err
	}

	return metrics.Gauge(value), nil
}

func (s DatabaseStorage) GetCounter(ctx context.Context, metricName string) (metrics.Counter, error) {
	var value int

	row := s.conn.QueryRow(ctx, getMetricSQL, metricName)
	err := row.Scan(&value)
	if err != nil {
		return 0, err
	}

	return metrics.Counter(value), nil
}

// language=PostgreSQL
const getAllMetricsSQL = `SELECT name, type, value  FROM metrics`

func (s DatabaseStorage) GetAllMetrics(ctx context.Context, resetCounters bool) []metrics.Metric {
	rows, err := s.conn.Query(ctx, getAllMetricsSQL)
	if err != nil {
		return nil
	}

	var result metrics.Metrics

	for rows.Next() {
		var name, metricType string
		var rawValue float64
		err := rows.Scan(&name, &metricType, &rawValue)
		if err != nil {
			return nil
		}

		metric := metrics.Metric{}
		switch metricType {
		case metrics.GaugeTypeName:
			value := metrics.Gauge(rawValue)
			metric.Value = &value
		case metrics.CounterTypeName:
			delta := metrics.Counter(rawValue)
			metric.Delta = &delta
		default:
			return nil
		}

		result = append(result, metric)
	}

	return []metrics.Metric{}
}

// language=PostgreSQL
const setMetricSQL = `
	INSERT INTO metrics (name, type, value)
	VALUES ('test', 'gauge', 5.123)
	ON CONFLICT ON CONSTRAINT metrics_name_key DO UPDATE
	SET value = excluded.value
`

func (s DatabaseStorage) SetGauge(ctx context.Context, metricName string, value metrics.Gauge) {
	s.conn.Exec(ctx, setMetricSQL, metricName, metrics.GaugeTypeName, value)
}

func (s DatabaseStorage) IncCounter(ctx context.Context, metricName string, value metrics.Counter) {
	s.conn.Exec(ctx, setMetricSQL, metricName, metrics.CounterTypeName, value)
}

func (s DatabaseStorage) ShutDown(ctx context.Context) error {
	return s.conn.Close(ctx)
}

func (s DatabaseStorage) Ping(ctx context.Context) error {
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

func (s DatabaseStorage) Migrate(ctx context.Context) error {
	_, err := s.conn.Exec(ctx, createMetricsTable)
	return err
}
