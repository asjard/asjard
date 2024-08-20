package handlers

import (
	"context"
	"fmt"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server/handlers"
	_ "github.com/asjard/asjard/pkg/client/grpc"
	"github.com/asjard/asjard/pkg/protobuf/healthpb"
	"github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
)

type Health struct {
	healthpb.UnimplementedHealthServer
}

func init() {
	handlers.AddServerDefaultHandler("health", &Health{})
}

// Check 健康检查
func (Health) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	if in.Service != "" {
		conn, err := client.NewClient(grpc.Protocol, config.GetString(fmt.Sprintf("asjard.topology.services.%s.name", in.Service), in.Service)).Conn()
		if err == nil {
			out := new(healthpb.HealthCheckResponse)
			err = conn.Invoke(ctx, healthpb.Health_Check_FullMethodName, &healthpb.HealthCheckRequest{}, out)
			if err == nil {
				return out, nil
			}
		}
		if err != nil {
			logger.Error("health check fail", "service", in.Service, "err", err)
			return &healthpb.HealthCheckResponse{
				Status:  healthpb.HealthCheckResponse_NOT_SERVING,
				Service: in.Service,
			}, nil
		}
	}
	return &healthpb.HealthCheckResponse{
		Status:  healthpb.HealthCheckResponse_SERVING,
		Service: runtime.GetAPP().Instance.Name,
	}, nil
}

func (Health) RestServiceDesc() *rest.ServiceDesc {
	return &healthpb.HealthRestServiceDesc
}

func (Health) GrpcServiceDesc() *grpc.ServiceDesc {
	return &healthpb.Health_ServiceDesc
}
