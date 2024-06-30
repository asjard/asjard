package handlers

import (
	"context"

	"github.com/asjard/asjard/pkg/protobuf/healthpb"
	"github.com/asjard/asjard/pkg/server/rest"
	"google.golang.org/grpc"
)

type Health struct {
	healthpb.UnimplementedHealthServer
}

func init() {
	AddServerDefaultHandler("health", &Health{})
}

// Check 健康检查
func (Health) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{
		Status: healthpb.HealthCheckResponse_SERVING,
	}, nil
}

func (Health) RestServiceDesc() *rest.ServiceDesc {
	return &healthpb.HealthRestServiceDesc
}

func (Health) GrpcServiceDesc() *grpc.ServiceDesc {
	return &healthpb.Health_ServiceDesc
}
