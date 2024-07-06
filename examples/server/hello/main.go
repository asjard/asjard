package main

import (
	"bufio"
	"context"
	"fmt"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Hello 同一个方法既可以当做GRPC的handler，也可以当做http的handler
type Hello struct {
	pb.UnimplementedHelloServer
	conn pb.HelloClient
	exit chan struct{}
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

// Shutdown
func (c *Hello) Shutdown() {}

// Say .
func (c *Hello) Say(ctx context.Context, in *pb.SayReq) (*pb.SayReq, error) {
	// HTTP 调用GRPC
	logger.Debug("=============xxxxxxx==========", "T", fmt.Sprintf("%T", ctx))
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
	md, ok := metadata.FromIncomingContext(ctx)
	logger.Debug("===========", "md", md, "ok", ok)
	if in.Configs == nil {
		in.Configs = &pb.Configs{}
	}
	in.Configs.Timeout = config.GetString("examples.timeout", "")
	in.Configs.FieldInDifferentFileUnderSameSection = config.GetString("examples.fieldInDifferentFileUnderSameSection", "")
	in.Configs.AnotherFieldInDifferentFileUnderSameSection = config.GetString("examples.anotherFieldInDifferentFileUnderSameSection", "")
	return in, nil
}

func (c *Hello) Log(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	logger.Debug("------is stream request-------")
	rtx, ok := ctx.(*rest.Context)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "unsupport protocol")
	}
	rtx.SetContentType("text/event-stream")
	rtx.SetBodyStreamWriter(func(w *bufio.Writer) {
		for {
			select {
			case <-c.exit:
				return
			default:
				w.Write([]byte(fmt.Sprintf("data: %s\n\n", time.Now())))

				if err := w.Flush(); err != nil {
					logger.Debug("client disconnected", "err", err)
					return
				}

				time.Sleep(time.Second)
			}
		}
	})
	return nil, nil
}

// RestServiceDesc .
func (Hello) RestServiceDesc() *rest.ServiceDesc {
	return &pb.HelloRestServiceDesc
}

// GrpcServiceDesc .
func (Hello) GrpcServiceDesc() *grpc.ServiceDesc {
	return &pb.Hello_ServiceDesc
}

func main() {
	server := asjard.New()
	helloServer := &Hello{
		exit: server.Exit(),
	}
	server.AddHandler(rest.Protocol, helloServer)
	server.AddHandler(mgrpc.Protocol, helloServer)
	if err := server.Start(); err != nil {
		log.Println(err.Error())
	}
}
