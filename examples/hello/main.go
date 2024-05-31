package main

import (
	"context"
	"log"

	"github.com/asjard/asjard"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	pb "github.com/asjard/asjard/examples/protobuf/hello"
	mgrpc "github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
	"google.golang.org/grpc"
)

// Hello .
type Hello struct {
	pb.UnimplementedHelloServer
}

var _ pb.HelloServer = &Hello{}

// Say .
func (Hello) Say(ctx context.Context, in *pb.SayReq) (*pb.SayReq, error) {
	logger.Infof(config.GetString("servers.rest1.interceptors", ""))
	return in, nil
}

// RestServiceDesc .
func (Hello) RestServiceDesc() rest.ServiceDesc {
	return pb.HelloRestServiceDesc
}

// GrpcServiceDesc .
func (Hello) GrpcServiceDesc() *grpc.ServiceDesc {
	return &pb.Hello_ServiceDesc
}

func main() {
	server := asjard.New()
	server.AddHandler(rest.Protocol, &Hello{})
	server.AddHandler(mgrpc.Protocol, &Hello{})
	if err := server.Start(); err != nil {
		log.Println(err.Error())
	}
}
