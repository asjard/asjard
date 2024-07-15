package rest

import (
	"context"

	"github.com/asjard/asjard/pkg/ajerr"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ErrorHandlerAPI struct {
	UnimplementedErrorHandlerServer
}

func (ErrorHandlerAPI) NotFound(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, ajerr.PageNotFoundError
}

func (ErrorHandlerAPI) MethodNotAllowed(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, ajerr.MethodNotAllowedError
}

func (ErrorHandlerAPI) Panic(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, ajerr.InternalServerError
}

func (ErrorHandlerAPI) RestServiceDesc() *ServiceDesc {
	return &ErrorHandlerRestServiceDesc
}
