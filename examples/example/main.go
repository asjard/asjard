package main

import (
	"context"

	"github.com/asjard/asjard"
	"github.com/asjard/asjard/examples/example/hellopb"
	mgrpc "github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
	"google.golang.org/grpc"
)

// HelloAPI hello相关接口
type HelloAPI struct {
	hellopb.UnimplementedHelloServer
}

func (api *HelloAPI) Say(ctx context.Context, in *hellopb.SayReq) (*hellopb.SayResp, error) {
	return &hellopb.SayResp{
		Message: "hello " + in.Name,
	}, nil
}

// 如果需要提供grpc服务则需要实现此方法
func (api *HelloAPI) GrpcServiceDesc() *grpc.ServiceDesc {
	return &hellopb.Hello_ServiceDesc
}

// 如果需要提供rest服务则需要实现此方法
func (api *HelloAPI) RestServiceDesc() *rest.ServiceDesc {
	return &hellopb.HelloRestServiceDesc
}

func main() {
	server := asjard.New()
	// 同时提供grpc和rest服务
	server.AddHandler(&HelloAPI{}, rest.Protocol, mgrpc.Protocol)
	// 启动服务
	if err := server.Start(); err != nil {
		panic(err)
	}
}
