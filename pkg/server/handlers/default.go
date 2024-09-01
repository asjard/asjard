package handlers

import (
	"context"
	"path/filepath"

	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server/handlers"
	"github.com/asjard/asjard/pkg/protobuf/requestpb"
	"github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/utils"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/types/known/emptypb"
)

type DefaultHandlersAPI struct {
	requestpb.UnimplementedDefaultHandlersServer
}

func init() {
	handlers.AddServerDefaultHandler("default", &DefaultHandlersAPI{})
}

func (api *DefaultHandlersAPI) Favicon(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	rtx, ok := ctx.(*rest.Context)
	if ok {
		faviconFile := filepath.Join(utils.GetHomeDir(), runtime.GetAPP().Favicon)
		if utils.IsPathExists(faviconFile) {
			rtx.Response.Header.Set(fasthttp.HeaderContentType, "image/x-icon")
			rtx.SendFile(faviconFile)
			return nil, nil
		}
	}
	return &emptypb.Empty{}, nil
}

func (api *DefaultHandlersAPI) RestServiceDesc() *rest.ServiceDesc {
	return &requestpb.DefaultHandlersRestServiceDesc
}

func (api *DefaultHandlersAPI) GrpcServiceDesc() *grpc.ServiceDesc {
	return &requestpb.DefaultHandlers_ServiceDesc
}
