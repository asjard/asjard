package handlers

import (
	"context"
	"fmt"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server/handlers"
	_ "github.com/asjard/asjard/pkg/client/grpc" // Side-effect import to register gRPC client
	"github.com/asjard/asjard/pkg/protobuf/healthpb"
	"github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
)

// Health implements the healthpb.HealthServer interface.
type Health struct {
	healthpb.UnimplementedHealthServer
}

func init() {
	// Automatically register the health handler for both gRPC and REST protocols.
	// This ensures the service is discoverable by load balancers and monitoring tools.
	handlers.AddServerDefaultHandler("health", &Health{}, grpc.Protocol, rest.Protocol)
}

// Check performs a health check on the current service or a specified downstream service.
func (Health) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	// If a specific service name is provided in the request, perform a downstream check.
	if in.Service != "" {
		// Look up the service name in the application topology configuration.
		serviceName := config.GetString(fmt.Sprintf("asjard.topology.services.%s.name", in.Service), in.Service)

		// Establish a gRPC client connection to the target downstream service.
		conn, err := client.NewClient(grpc.Protocol, serviceName).Conn()
		if err == nil {
			out := new(healthpb.HealthCheckResponse)
			// Forward the health check request to the downstream service.
			err = conn.Invoke(ctx, healthpb.Health_Check_FullMethodName, &healthpb.HealthCheckRequest{}, out)
			if err == nil {
				return out, nil
			}
		}

		// If the downstream check fails, log the error and return a NOT_SERVING status.
		if err != nil {
			logger.Error("health check fail", "service", in.Service, "err", err)
			return &healthpb.HealthCheckResponse{
				Status:  healthpb.HealthCheckResponse_NOT_SERVING,
				Service: in.Service,
			}, nil
		}
	}

	// Default behavior: Return SERVING for the current instance.
	return &healthpb.HealthCheckResponse{
		Status:  healthpb.HealthCheckResponse_SERVING,
		Service: runtime.GetAPP().Instance.Name,
	}, nil
}

// RestServiceDesc returns the RESTful service description for the health check.
func (Health) RestServiceDesc() *rest.ServiceDesc {
	return &healthpb.HealthRestServiceDesc
}

// GrpcServiceDesc returns the gRPC service description for the health check.
func (Health) GrpcServiceDesc() *grpc.ServiceDesc {
	return &healthpb.Health_ServiceDesc
}
