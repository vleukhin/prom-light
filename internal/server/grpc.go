package server

import (
	"context"

	"google.golang.org/grpc/codes"

	"github.com/vleukhin/prom-light/internal/metrics"
	"github.com/vleukhin/prom-light/internal/proto"
	"github.com/vleukhin/prom-light/internal/storage"

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
	m, err := metrics.FromProto(request.Metric)
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
		m, err := metrics.FromProto(i)
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
		Metric: metrics.ToProto(m),
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

func NewGRPCServer(addr string, store storage.MetricsStorage) GRPSServer {
	server := grpc.NewServer()
	proto.RegisterMetricsServer(server, newMetricsServer(store))

	return GRPSServer{
		addr:   addr,
		server: server,
	}
}
