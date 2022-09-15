package server

import (
	"context"
	"github.com/vleukhin/prom-light/internal/metrics"
	"github.com/vleukhin/prom-light/internal/proto"
	"github.com/vleukhin/prom-light/internal/storage"
	"google.golang.org/grpc/codes"

	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type GRPSServer struct {
	server *grpc.Server
	addr   string
}

type MetricsServer struct {
	proto.UnimplementedMetricsServer
	store storage.MetricsStorage
}

func newMetricsServer(store storage.MetricsStorage) proto.MetricsServer {
	return &MetricsServer{
		store: store,
	}
}

func (s MetricsServer) UpdateMetric(ctx context.Context, request *proto.UpdateMetricRequest) (*proto.UpdateMetricResponse, error) {
	m, err := metricFromGRPC(request.Metric)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unknown metric type '%s'", request.Metric.Type)
	}

	err = s.store.SetMetric(ctx, m)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update metric")
	}

	return &proto.UpdateMetricResponse{}, nil
}

func (s MetricsServer) UpdateMetricsBatch(ctx context.Context, request *proto.UpdateMetricsBatchRequest) (*proto.UpdateMetricsBatchResponse, error) {
	mtrcs := make(metrics.Metrics, len(request.Metrics))
	for _, i := range request.Metrics {
		m, err := metricFromGRPC(i)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "unknown metric type '%s'", i.Type)
		}
		mtrcs = append(mtrcs, m)
	}
	err := s.store.SetMetrics(ctx, mtrcs)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update metrics")
	}
	return &proto.UpdateMetricsBatchResponse{}, nil
}

func (s MetricsServer) GetMetric(ctx context.Context, request *proto.GetMetricRequest) (*proto.GetMetricResponse, error) {
	var m metrics.Metric
	switch request.Type {
	case proto.MetricType_GAUGE:
		v, err := s.store.GetGauge(ctx, request.Name)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to get gauge")
		}
		m = metrics.MakeGaugeMetric(request.Name, v)
	case proto.MetricType_COUNTER:
		v, err := s.store.GetCounter(ctx, request.Name)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to get counter")
		}
		m = metrics.MakeCounterMetric(request.Name, v)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown metric type '%s'", request.Type)
	}

	return &proto.GetMetricResponse{
		Metric: metricToGRPC(m),
	}, nil
}

func (s GRPSServer) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	return s.server.Serve(listener)
}

func (s GRPSServer) Shutdown(_ context.Context) error {
	s.server.Stop()
	return nil
}

func NewGRPCServer(store storage.MetricsStorage) Server {
	server := grpc.NewServer()
	proto.RegisterMetricsServer(server, newMetricsServer(store))

	return GRPSServer{
		server: grpc.NewServer(),
	}
}

func metricFromGRPC(metric *proto.Metric) (metrics.Metric, error) {
	switch metric.Type {
	case proto.MetricType_GAUGE:
		return metrics.MakeGaugeMetric(metric.Name, metrics.Gauge(metric.Value)), nil
	case proto.MetricType_COUNTER:
		return metrics.MakeCounterMetric(metric.Name, metrics.Counter(metric.Delta)), nil
	default:
		return metrics.Metric{}, status.Errorf(codes.InvalidArgument, "unknown metric type '%s'", metric.Type)
	}
}

func metricToGRPC(metric metrics.Metric) *proto.Metric {
	m := &proto.Metric{
		Name:  metric.Name,
		Type:  metricTypeToGRPC(metric.Type),
		Value: float64(*metric.Value),
		Delta: int64(*metric.Delta),
	}

	switch metric.Type {
	case metrics.GaugeTypeName:
		if metric.Value != nil {
			m.Value = float64(*metric.Value)
		}
	default:
		if metric.Delta != nil {
			m.Delta = int64(*metric.Delta)
		}
	}

	return m
}

func metricTypeToGRPC(t string) proto.MetricType {
	switch t {
	case metrics.GaugeTypeName:
		return proto.MetricType_GAUGE
	default:
		return proto.MetricType_COUNTER
	}
}
