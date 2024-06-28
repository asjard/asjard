package ajerr

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// InternalServerError 系统内部错误
	InternalServerError = status.Error(codes.Internal, "internal server error")
)
