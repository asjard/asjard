package rest

import (
	"context"

	"github.com/asjard/asjard/core/server"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

const (
	// HeaderResponseRequestMethod 请求方法返回头
	HeaderResponseRequestMethod = "x-request-method"
	// HeaderResponseRequestID 请求ID返回头
	HeaderResponseRequestID = "x-request-id"

	RestReadEntityInterceptorName     = "restReadEntity"
	RestResponseHeaderInterceptorName = "restResponseHeader"
)

func init() {
	// 请求参数自动解析
	server.AddInterceptor(RestReadEntityInterceptorName, NewReadEntityInterceptor, Protocol)
	// 统一添加返回头
	server.AddInterceptor(RestResponseHeaderInterceptorName, NewResponseHeaderInterceptor, Protocol)
}

// NewReadEntityInterceptor 初始化序列化参数拦截器
func NewReadEntityInterceptor() server.ServerInterceptor {
	return &ReadEntity{}
}

// NewResponseHeaderInterceptor 初始化返回请求头拦截器
func NewResponseHeaderInterceptor() server.ServerInterceptor {
	return &ResponseHeader{}
}

// ReadEntity 解析参数到请求参数中
type ReadEntity struct{}

// Name .
func (r *ReadEntity) Name() string {
	return RestReadEntityInterceptorName
}

// Interceptor .
func (r *ReadEntity) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		rc := ctx.(*Context)
		if err := rc.ReadEntity(req.(proto.Message)); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// ResponseHeader 添加返回头
type ResponseHeader struct{}

// Name .
func (ResponseHeader) Name() string {
	return RestResponseHeaderInterceptorName
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
