package rest

import (
	"context"

	"github.com/asjard/asjard/core/server"
	"github.com/google/uuid"
)

const (
	// HeaderResponseRequestMethod 请求方法返回头
	HeaderResponseRequestMethod = "x-request-method"
	// HeaderResponseRequestID 请求ID返回头
	HeaderResponseRequestID = "x-request-id"
)

// ReadEntity 解析参数到请求参数中
type ReadEntity struct{}

func init() {
	server.AddInterceptor(&ReadEntity{})
	server.AddInterceptor(&ResponseHeader{})
}

// Name .
func (r *ReadEntity) Name() string {
	return "restReadEntity"
}

// Interceptor .
func (r *ReadEntity) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		rc := ctx.(*Context)
		rc.ReadEntity(req)
		return handler(ctx, req)
	}
}

// ResponseHeader 添加返回头
type ResponseHeader struct{}

// Name .
func (ResponseHeader) Name() string {
	return "restResponseHeader"
}

// Interceptor .
func (ResponseHeader) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		rc := ctx.(*Context)
		rc.Response.Header.Add(HeaderResponseRequestID, uuid.NewString())
		if info != nil {
			rc.Response.Header.Add(HeaderResponseRequestMethod, info.FullMethod)
		}
		return handler(ctx, req)
	}
}
