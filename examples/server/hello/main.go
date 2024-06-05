package main

import (
	"context"
	"log"

	"github.com/asjard/asjard"
	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	pb "github.com/asjard/asjard/examples/protobuf/hello"
	_ "github.com/asjard/asjard/pkg/client/grpc"
	mgrpc "github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
	"google.golang.org/grpc"
)

// Hello .
type Hello struct {
	pb.UnimplementedHelloServer
	conn pb.HelloClient
}

var _ pb.HelloServer = &Hello{}

// Bootstrap .
func (c *Hello) Bootstrap() error {
	conn, err := client.NewClient(mgrpc.Protocol, "helloGrpc").Conn()
	if err != nil {
		return err
	}
	c.conn = pb.NewHelloClient(conn)
	return nil
}

func (c *Hello) Shutdown() {
}

// Say .
func (c Hello) Say(ctx context.Context, in *pb.SayReq) (*pb.SayReq, error) {
	resp, err := c.conn.Call(ctx, in)
	if err != nil {
		logger.Errorf("call fail[%s]", err.Error())
		return in, nil
	}
	return resp, nil
}

// Call .
func (Hello) Call(ctx context.Context, in *pb.SayReq) (*pb.SayReq, error) {
	in.RegionId = "changed by call " + config.GetString("testEnv", "")
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
	helloClient := &Hello{}
	bootstrap.AddBootstrap(helloClient)
	server.AddHandler(rest.Protocol, helloClient)
	server.AddHandler(mgrpc.Protocol, helloClient)
	if err := server.Start(); err != nil {
		log.Println(err.Error())
	}
}
