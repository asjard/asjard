package apiv1

import (
	"context"
	"protos-repo/example/api/v1/hello"

	"github.com/asjard/asjard/pkg/server/grpc"
)

type HelloAPI struct {
	hello.UnimplementedHelloServer
}

func NewHelloAPI() *HelloAPI {
	return &HelloAPI{}
}

func (api *HelloAPI) GrpcServiceDesc() *grpc.ServiceDesc { return &hello.Hello_ServiceDesc }

func (api *HelloAPI) Say(ctx context.Context, in *hello.SayReq) (*hello.SayResp, error) {
	return &hello.SayResp{
		Message: "hello " + in.Name,
	}, nil
}
