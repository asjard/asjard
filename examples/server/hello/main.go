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
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/status"
	pb "github.com/asjard/asjard/examples/protobuf/hello"
	_ "github.com/asjard/asjard/pkg/client/grpc"
	_ "github.com/asjard/asjard/pkg/registry/etcd"
	mgrpc "github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Hello 同一个方法既可以当做GRPC的handler，也可以当做http的handler
type Hello struct {
	pb.UnimplementedHelloServer
	app  runtime.APP
	conn pb.HelloClient
	exit <-chan struct{}
}

var _ pb.HelloServer = &Hello{}

// Bootstrap .
func (c *Hello) Bootstrap() error {
	conn, err := client.NewClient(mgrpc.Protocol, runtime.GetAPP().Instance.Name).Conn()
	if err != nil {
		return err
	}
	c.conn = pb.NewHelloClient(conn)
	c.app = runtime.GetAPP()
	return nil
}

// Shutdown
func (c *Hello) Shutdown() {}

func (c *Hello) Hello(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

// Say .
func (c *Hello) Say(ctx context.Context, in *pb.SayReq) (*pb.SayReq, error) {
	// HTTP 调用GRPC
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
	// time.Sleep(500 * time.Millisecond)
	md, ok := metadata.FromIncomingContext(ctx)
	logger.Debug("===========", "md", md, "ok", ok)
	// return nil, ajerr.InternalServerError
	if in.Configs == nil {
		in.Configs = &pb.Configs{}
	}
	in.Configs.Timeout = config.GetString("examples.timeout", "")
	in.Configs.FieldInDifferentFileUnderSameSection = config.GetString("examples.fieldInDifferentFileUnderSameSection", "")
	in.Configs.AnotherFieldInDifferentFileUnderSameSection = config.GetString("examples.anotherFieldInDifferentFileUnderSameSection", "")
	in.App = c.app.App
	in.Region = c.app.Region
	in.Az = c.app.AZ
	in.Instance = &pb.SayReq_Instance{
		Id:       c.app.Instance.ID,
		Name:     c.app.Instance.Name,
		Version:  c.app.Instance.Version,
		Metadata: c.app.Instance.MetaData,
	}
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
				logger.Debug("------server done----------")
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
	server.AddHandler(helloServer, rest.Protocol, mgrpc.Protocol)
	if err := server.Start(); err != nil {
		log.Println(err.Error())
	}
}
