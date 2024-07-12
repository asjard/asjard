package ajerr

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// InternalServerError 系统内部错误
	InternalServerError   = status.Error(codes.Internal, "internal server error")
	PageNotFoundError     = status.Error(codes.NotFound, "page not found")
	MethodNotAllowedError = status.Error(codes.Unimplemented, "method not allowed")
)
