// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.6.1
// source: internal/proto/metrics.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// MetricsClient is the client API for Metrics service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MetricsClient interface {
	UpdateMetric(ctx context.Context, in *UpdateMetricRequest, opts ...grpc.CallOption) (*UpdateMetricResponse, error)
	UpdateMetricsBatch(ctx context.Context, in *UpdateMetricsBatchRequest, opts ...grpc.CallOption) (*UpdateMetricsBatchResponse, error)
	GetMetric(ctx context.Context, in *GetMetricRequest, opts ...grpc.CallOption) (*GetMetricResponse, error)
}

type metricsClient struct {
	cc grpc.ClientConnInterface
}

func NewMetricsClient(cc grpc.ClientConnInterface) MetricsClient {
	return &metricsClient{cc}
}

func (c *metricsClient) UpdateMetric(ctx context.Context, in *UpdateMetricRequest, opts ...grpc.CallOption) (*UpdateMetricResponse, error) {
	out := new(UpdateMetricResponse)
	err := c.cc.Invoke(ctx, "/metrics.Metrics/UpdateMetric", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricsClient) UpdateMetricsBatch(ctx context.Context, in *UpdateMetricsBatchRequest, opts ...grpc.CallOption) (*UpdateMetricsBatchResponse, error) {
	out := new(UpdateMetricsBatchResponse)
	err := c.cc.Invoke(ctx, "/metrics.Metrics/UpdateMetricsBatch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricsClient) GetMetric(ctx context.Context, in *GetMetricRequest, opts ...grpc.CallOption) (*GetMetricResponse, error) {
	out := new(GetMetricResponse)
	err := c.cc.Invoke(ctx, "/metrics.Metrics/GetMetric", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MetricsServer is the server API for Metrics service.
// All implementations must embed UnimplementedMetricsServer
// for forward compatibility
type MetricsServer interface {
	UpdateMetric(context.Context, *UpdateMetricRequest) (*UpdateMetricResponse, error)
	UpdateMetricsBatch(context.Context, *UpdateMetricsBatchRequest) (*UpdateMetricsBatchResponse, error)
	GetMetric(context.Context, *GetMetricRequest) (*GetMetricResponse, error)
	mustEmbedUnimplementedMetricsServer()
}

// UnimplementedMetricsServer must be embedded to have forward compatible implementations.
type UnimplementedMetricsServer struct {
}

func (UnimplementedMetricsServer) UpdateMetric(context.Context, *UpdateMetricRequest) (*UpdateMetricResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateMetric not implemented")
}
func (UnimplementedMetricsServer) UpdateMetricsBatch(context.Context, *UpdateMetricsBatchRequest) (*UpdateMetricsBatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateMetricsBatch not implemented")
}
func (UnimplementedMetricsServer) GetMetric(context.Context, *GetMetricRequest) (*GetMetricResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMetric not implemented")
}
func (UnimplementedMetricsServer) mustEmbedUnimplementedMetricsServer() {}

// UnsafeMetricsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MetricsServer will
// result in compilation errors.
type UnsafeMetricsServer interface {
	mustEmbedUnimplementedMetricsServer()
}

func RegisterMetricsServer(s grpc.ServiceRegistrar, srv MetricsServer) {
	s.RegisterService(&Metrics_ServiceDesc, srv)
}

func _Metrics_UpdateMetric_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateMetricRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServer).UpdateMetric(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/metrics.Metrics/UpdateMetric",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServer).UpdateMetric(ctx, req.(*UpdateMetricRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Metrics_UpdateMetricsBatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateMetricsBatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServer).UpdateMetricsBatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/metrics.Metrics/UpdateMetricsBatch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServer).UpdateMetricsBatch(ctx, req.(*UpdateMetricsBatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Metrics_GetMetric_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMetricRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServer).GetMetric(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/metrics.Metrics/GetMetric",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServer).GetMetric(ctx, req.(*GetMetricRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Metrics_ServiceDesc is the grpc.ServiceDesc for Metrics service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Metrics_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "metrics.Metrics",
	HandlerType: (*MetricsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "UpdateMetric",
			Handler:    _Metrics_UpdateMetric_Handler,
		},
		{
			MethodName: "UpdateMetricsBatch",
			Handler:    _Metrics_UpdateMetricsBatch_Handler,
		},
		{
			MethodName: "GetMetric",
			Handler:    _Metrics_GetMetric_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "internal/proto/metrics.proto",
}
