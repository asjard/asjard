package main

import (
	"context"
	"log"
	"time"

	"github.com/asjard/asjard"
	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	pb "github.com/asjard/asjard/examples/protobuf/hello"
	_ "github.com/asjard/asjard/pkg/client/grpc"
	mgrpc "github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Hello 同一个方法既可以当做GRPC的handler，也可以当做http的handler
type Hello struct {
	pb.UnimplementedHelloServer
	conn  pb.HelloClient
	conn1 pb.HelloClient
}

var _ pb.HelloServer = &Hello{}

// Bootstrap .
func (c *Hello) Bootstrap() error {
	conn, err := client.NewClient(mgrpc.Protocol, "helloGrpc").Conn()
	if err != nil {
		return err
	}
	c.conn = pb.NewHelloClient(conn)
	conn1, err := client.NewClient(mgrpc.Protocol, "helloGrpc1").Conn()
	if err != nil {
		return err
	}
	c.conn1 = pb.NewHelloClient(conn1)
	return nil
}

// Shutdown
func (c *Hello) Shutdown() {}

// Say .
func (c *Hello) Say(ctx context.Context, in *pb.SayReq) (*pb.SayReq, error) {
	// HTTP 调用GRPC
	logger.Debug("=============xxxxxxx==========")
	resp, err := c.conn.Call(ctx, in)
	if err != nil {
		logger.Error("call call fail",
			"err", err.Error())
		return nil, err
	}
	return resp, err
}

// Call .
func (c *Hello) Call(ctx context.Context, in *pb.SayReq) (*pb.SayReq, error) {
	time.Sleep(config.GetDuration("sleep", 2*time.Millisecond))
	md, ok := metadata.FromIncomingContext(ctx)
	logger.Debug("===========", "md", md, "ok", ok)
	in.RegionId = "timeout: " + config.GetString("sleep", "")
	// resp, err := c.conn.Say(ctx, in)
	// if err != nil {
	// 	return in, err
	// }
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
	helloServer := &Hello{}
	server.AddHandler(rest.Protocol, helloServer)
	server.AddHandler(mgrpc.Protocol, helloServer)
	if err := server.Start(); err != nil {
		log.Println(err.Error())
	}
}
