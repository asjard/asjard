package main

import (
	"context"
	"log"
	"sync/atomic"

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

// Hello .
type Hello struct {
	pb.UnimplementedHelloServer
	conn pb.HelloClient
}

var _ pb.HelloServer = &Hello{}

var count int32 = 0

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
func (c *Hello) Say(ctx context.Context, in *pb.SayReq) (*pb.SayReq, error) {
	logger.Debug("----------------", "count", count)
	if atomic.LoadInt32(&count) >= 3 {
		atomic.StoreInt32(&count, 0)
		return in, nil
	}
	resp, err := c.conn.Call(ctx, in)
	if err != nil {
		logger.Error("call fail",
			"err", err.Error())
		return in, nil
	}
	atomic.AddInt32(&count, 1)
	return resp, nil
}

// Call .
func (c *Hello) Call(ctx context.Context, in *pb.SayReq) (*pb.SayReq, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	logger.Debug("---------------", "md", md, "ok", ok)
	// _, err := c.conn.Say(ctx, in)
	// if err != nil {
	// 	logger.Error("call fail",
	// 		"err", err.Error())
	// 	return in, nil
	// }
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
	server.AddHandler(rest.Protocol, &Hello{})
	server.AddHandler(mgrpc.Protocol, &Hello{})
	if err := server.Start(); err != nil {
		log.Println(err.Error())
	}
}
