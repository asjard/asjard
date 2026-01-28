package apiv1

import (
	"context"

	pb "github.com/asjard/asjard/_examples/protos-repo/example/api/v1/sample"
	"github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
)

type SampleAPI struct {
	pb.UnimplementedSampleServer
}

func NewSampleAPI() *SampleAPI {
	return &SampleAPI{}
}

func (api *SampleAPI) Start() error { return nil }
func (api *SampleAPI) Stop()        {}

// GRPC服务
func (api *SampleAPI) GrpcServiceDesc() *grpc.ServiceDesc { return &pb.Sample_ServiceDesc }

// HTTP服务
func (api *SampleAPI) RestServiceDesc() *rest.ServiceDesc { return &pb.SampleRestServiceDesc }

func (api *SampleAPI) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{
		Message: "hello " + in.Name,
	}, nil
}
