package rest

import (
	"context"

	"github.com/asjard/asjard/core/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ErrorHandlerAPI struct {
	UnimplementedErrorHandlerServer
}

func (ErrorHandlerAPI) Error(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, status.InternalServerError
}

func (ErrorHandlerAPI) NotFound(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, status.PageNotFoundError
}

func (ErrorHandlerAPI) MethodNotAllowed(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, status.MethodNotAllowedError
}

func (ErrorHandlerAPI) Panic(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, status.InternalServerError
}

func (ErrorHandlerAPI) RestServiceDesc() *ServiceDesc {
	return &ErrorHandlerRestServiceDesc
}
