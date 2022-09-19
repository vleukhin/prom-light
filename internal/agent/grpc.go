package agent

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/vleukhin/prom-light/internal/metrics"
	"github.com/vleukhin/prom-light/internal/proto"
)

type grpcClient struct {
	conn   *grpc.ClientConn
	client proto.MetricsClient
}

func NewGRPCClient(addr string) (*grpcClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &grpcClient{
		conn:   conn,
		client: proto.NewMetricsClient(conn),
	}, nil
}

func (g *grpcClient) SendMetricToServer(ctx context.Context, m metrics.Metric) error {
	req := &proto.UpdateMetricRequest{Metric: metrics.ToProto(m)}
	_, err := g.client.UpdateMetric(ctx, req)
	return err
}

func (g *grpcClient) SendBatchMetricsToServer(ctx context.Context, m metrics.Metrics) error {
	req := &proto.UpdateMetricsBatchRequest{Metrics: metrics.BatchToProto(m)}
	_, err := g.client.UpdateMetricsBatch(ctx, req)
	return err
}

func (g *grpcClient) ShutDown() error {
	return g.conn.Close()
}
