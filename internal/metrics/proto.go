package metrics

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vleukhin/prom-light/internal/proto"
)

func FromProto(metric *proto.Metric) (Metric, error) {
	switch metric.Type {
	case proto.MetricType_GAUGE:
		return MakeGaugeMetric(metric.Name, Gauge(metric.Value)), nil
	case proto.MetricType_COUNTER:
		return MakeCounterMetric(metric.Name, Counter(metric.Delta)), nil
	default:
		return Metric{}, status.Errorf(codes.InvalidArgument, "unknown metric type '%s'", metric.Type)
	}
}

func ToProto(metric Metric) *proto.Metric {
	m := &proto.Metric{
		Name: metric.Name,
		Type: TypeToProto(metric.Type),
	}

	switch metric.Type {
	case GaugeTypeName:
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

func BatchToProto(metrics Metrics) []*proto.Metric {
	res := make([]*proto.Metric, 0, len(metrics))
	for _, m := range metrics {
		res = append(res, ToProto(m))

	}

	return res
}

func TypeToProto(t string) proto.MetricType {
	switch t {
	case GaugeTypeName:
		return proto.MetricType_GAUGE
	default:
		return proto.MetricType_COUNTER
	}
}
