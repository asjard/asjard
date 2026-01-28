## 接口实现

- 如需添加拦截器，则需实现如下功能
- 然后通过`client.AddInterceptor`注册拦截器，之后服务启动时将会通过配置调用相应的拦截器

```go

// UnaryServerInfo contains metadata about a single (unary) RPC call.
// This is passed to interceptors to provide context about the server and method being called.
type UnaryServerInfo struct {
	// Server is the underlying service implementation.
	Server any
	// FullMethod is the path to the RPC (e.g., "/user.UserService/GetUser").
	FullMethod string
	// Protocol identifies the transport (e.g., "grpc", "rest").
	Protocol string
}

// UnaryHandler is the signature of the final business logic or the next step in the chain.
type UnaryHandler func(ctx context.Context, req any) (any, error)

// UnaryServerInterceptor is a middleware function that wraps the request execution.
// It can modify the context/request before execution or the response/error after.
type UnaryServerInterceptor func(ctx context.Context, req any, info *UnaryServerInfo, handler UnaryHandler) (resp any, err error)

// ServerInterceptor is the interface for pluggable middleware components.
type ServerInterceptor interface {
	// Name returns the unique identifier for the interceptor (e.g., "logger").
	Name() string
	// Interceptor returns the actual function that performs the wrapping.
	Interceptor() UnaryServerInterceptor
}

```

## 已支持的实现

- [accessLog](interceptor-server-accessLog.md)
- [i18n](interceptor-server-i18n.md)
- [监控](interceptor-server-metrics.md)
- [panic日志](inteceptor-server-panic.md)
- [限速](inteceptor-server-ratelimit.md)
- [请求参数解析](inteceptor-server-restReadEntity.md)
- [链路追踪](inteceptor-server-trace.md)
- [参数校验](inteceptor-server-validate.md)

## 配置

```yaml
asjard:
  ## client configurations
  servers:
    ## 全局通用协议无关拦截器
    # interceptors: ""
    ## 全局内建默认拦截器
    # builtInInterceptors: panic,rest2RpcContext,validate,errLog,slowLog,cycleChainInterceptor
    ## 指定协议的配置
    grpc:
      ## grpc的拦截器
      # interceptors: ""
```

更新默认拦截器:

```yaml
asjard:
  ## client configurations
  servers:
    ## 全局通用协议无关拦截器
    # interceptors: ""
    ## 全局内建默认拦截器
    ## 局部更新语法
    ## -errLog 删除errLog拦截器
    ## +errLog:customeLog 在errLog拦截器前添加customeLog拦截器
    ## errLog+:customeLog 在errLog拦截器之后添加customeLog拦截器
    ## =errLog:customeLog 将errLog拦截器替换为customeLog拦截器
    ## customeLog 在默认拦截器尾部添加customeLog拦截器
    # builtInInterceptors: -errLog,customeLog
```

## 自定义实现

```go
package interceptors

import (
	"context"
	"runtime/debug"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/core/status"
)

const (
	// PanicInterceptorName is the unique identifier for this interceptor.
	PanicInterceptorName = "panic"
)

func init() {
	// Register the panic recovery interceptor globally for all server protocols.
	server.AddInterceptor(PanicInterceptorName, NewPanic)
}

// Panic represents the recovery interceptor component.
type Panic struct{}

// NewPanic initializes the panic interceptor.
func NewPanic() (server.ServerInterceptor, error) {
	return &Panic{}, nil
}

// Interceptor returns the middleware function that handles panic recovery.
func (*Panic) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		// Use defer to ensure the recovery logic runs even if the handler panics.
		defer func() {
			if rcv := recover(); rcv != nil {
				// 1. Collect diagnostic information about the crash.
				args := []any{
					"err", rcv, // The actual panic object/message.
					"req", req, // The request payload that triggered the panic.
					"method", info.FullMethod, // The endpoint being called.
					"protocol", info.Protocol, // gRPC or REST.
					"stack", string(debug.Stack()), // The full goroutine stack trace.
				}

				// 2. Log the incident with Error level for alerting and debugging.
				logger.L(ctx).Error("request panic", args...)

				// 3. Mask the internal crash from the client by returning a
				// standardized 500 Internal Server Error.
				err = status.InternalServerError()
			}
		}()

		// Execute the next interceptor or the final business logic handler.
		return handler(ctx, req)
	}
}

// Name returns the interceptor's unique name.
func (*Panic) Name() string {
	return PanicInterceptorName
}

```
