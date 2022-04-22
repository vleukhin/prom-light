package storage

import (
	"context"

	"github.com/jackc/pgx/v4"

	"github.com/vleukhin/prom-light/internal/metrics"
)

type DatabaseStorage struct {
	conn *pgx.Conn
}

func NewDatabaseStorage(ctx context.Context, dsn string) (DatabaseStorage, error) {
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

func (s DatabaseStorage) GetGauge(metricName string) (metrics.Gauge, error) {
	return 0, nil
}

func (s DatabaseStorage) GetCounter(metricName string) (metrics.Counter, error) {
	return 0, nil
}

func (s DatabaseStorage) GetAllMetrics(resetCounters bool) []metrics.Metric {
	return []metrics.Metric{}
}

func (s DatabaseStorage) SetGauge(metricName string, value metrics.Gauge) {

}

func (s DatabaseStorage) IncCounter(metricName string, value metrics.Counter) {

}

func (s DatabaseStorage) ShutDown() error {
	return s.conn.Close(context.TODO())
}
