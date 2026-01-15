package rest

import (
	"context"

	"github.com/asjard/asjard/core/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ErrorHandlerAPI provides standardized handlers for common HTTP error scenarios.
// It embeds UnimplementedErrorHandlerServer to satisfy the gRPC/Protobuf interface
// while only overriding specific error methods.
type ErrorHandlerAPI struct {
	UnimplementedErrorHandlerServer
}

// Error handles general internal server errors (HTTP 500).
func (ErrorHandlerAPI) Error(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, status.InternalServerError()
}

// NotFound handles requests to routes that do not exist (HTTP 404).
func (ErrorHandlerAPI) NotFound(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, status.PageNotFoundError()
}

// MethodNotAllowed handles requests where the URI exists but the HTTP verb is incorrect (HTTP 405).
func (ErrorHandlerAPI) MethodNotAllowed(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, status.MethodNotAllowedError()
}

// Panic handles runtime recovery. If the server recovers from a panic, this method
// ensures the client receives a valid Internal Server Error response.
func (ErrorHandlerAPI) Panic(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, status.InternalServerError()
}

// RestServiceDesc returns the RESTful service description used to register
// these system handlers within the routing table.
func (ErrorHandlerAPI) RestServiceDesc() *ServiceDesc {
	return &ErrorHandlerRestServiceDesc
}
