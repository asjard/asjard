package handlers

import (
	"context"

	"github.com/asjard/asjard/pkg/server/rest"
	pb "github.com/asjard/genproto/health"
	"google.golang.org/grpc"
)

type Health struct {
	pb.UnimplementedHealthServer
}

func init() {
	AddServerDefaultHandler("health", &Health{})
}

// Check 健康检查
func (Health) Check(ctx context.Context, in *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status: pb.HealthCheckResponse_SERVING,
	}, nil
}

func (Health) RestServiceDesc() *rest.ServiceDesc {
	return &pb.HealthRestServiceDesc
}

func (Health) GrpcServiceDesc() *grpc.ServiceDesc {
	return &pb.Health_ServiceDesc
}
