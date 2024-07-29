package main

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"github.com/asjard/asjard"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	pb "github.com/asjard/asjard/examples/protobuf/hello"
	_ "github.com/asjard/asjard/pkg/registry/etcd"
	"github.com/asjard/asjard/pkg/server/rest"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ pb.HelloServer = &pb.HelloAPI{}
var _ rest.Handler = &pb.HelloAPI{}

type Rewrite struct {
	*pb.HelloAPI
	exit <-chan struct{}
}

func (api Rewrite) Log(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	logger.Debug("------is stream request-------")
	rtx, ok := ctx.(*rest.Context)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "unsupport protocol")
	}
	rtx.SetContentType("text/event-stream")
	rtx.SetBodyStreamWriter(func(w *bufio.Writer) {
		for {
			select {
			case <-api.exit:
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

func main() {
	server := asjard.New()
	server.AddHandler(&Rewrite{
		HelloAPI: &pb.HelloAPI{},
		exit:     server.Exit(),
	}, rest.Protocol)
	if err := server.Start(); err != nil {
		panic(err)
	}
}
