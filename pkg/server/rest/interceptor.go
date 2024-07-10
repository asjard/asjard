package rest

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/utils"
	"github.com/google/uuid"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const (
	// HeaderResponseRequestMethod 请求方法返回头
	HeaderResponseRequestMethod = "x-request-method"
	// HeaderResponseRequestID 请求ID返回头
	HeaderResponseRequestID = "x-request-id"
)

func init() {
	// 请求参数自动解析
	server.AddInterceptor(NewReadEntityInterceptor, Protocol)
	// 统一添加返回头
	server.AddInterceptor(NewResponseHeaderInterceptor, Protocol)
	// 跨域请求
	server.AddInterceptor(NewCorsInterceptor, Protocol)
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

// Cors 跨域请求
type Cors struct {
	conf CorsConfig
}

// Cors 跨域配置
type CorsConfig struct {
	allowAllOrigins bool
	// 允许通过的源
	AllowOrigins []string `json:"allowOrigins"`
	// 允许通过的请求方式
	AllowMethods     []string           `json:"allowMethods"`
	AllowHeaders     []string           `json:"allowHeader"`
	AllowCredentials bool               `json:"allowCredentials"`
	ExposeHeaders    []string           `json:"exposeHeaders"`
	MaxAge           utils.JSONDuration `json:"maxAge"`
}

// NewCorsInterceptor 跨域拦截器初始化
func NewCorsInterceptor() server.ServerInterceptor {
	conf := CorsConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}
	config.GetWithUnmarshal(constant.ConfigServerInterceptorPrefix+".cors", &conf)
	for _, origin := range conf.AllowOrigins {
		if origin == "*" {
			conf.allowAllOrigins = true
		}
	}
	return &Cors{
		conf: conf,
	}
}

func (Cors) Name() string {
	return "cors"
}

// Interceptor 跨域请求拦截器实现
func (c Cors) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		rtx, ok := ctx.(*Context)
		if !ok {
			return handler(ctx, req)
		}
		origin := string(rtx.Request.Header.Peek("Origin"))
		if len(origin) == 0 {
			return handler(ctx, req)
		}
		host := string(rtx.Host())
		if origin == "http://"+host || origin == "https://"+host {
			return handler(ctx, req)
		}
		if !c.isOriginValid(origin) {
			return nil, status.Error(codes.PermissionDenied, "forbidden origin")
		}
		if c.conf.allowAllOrigins {
			rtx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		} else {
			rtx.Response.Header.Set("Access-Control-Allow-Origin", origin)
			rtx.Response.Header.Set("Vary", "Origin")
		}
		if string(rtx.Method()) == http.MethodOptions {
			if c.conf.AllowCredentials {
				rtx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
			}
			if len(c.conf.AllowMethods) > 0 {
				rtx.Response.Header.Set("Access-Control-Allow-Methods", strings.Join(c.conf.AllowMethods, ","))
			}
			if len(c.conf.AllowHeaders) > 0 {
				rtx.Response.Header.Set("Access-Control-Allow-Headers", strings.Join(c.conf.AllowHeaders, ","))
			}
			if c.conf.MaxAge.Duration != 0 {
				rtx.Response.Header.Set("Access-Control-Max-Age", strconv.FormatInt(int64(c.conf.MaxAge.Duration/time.Second), 10))
			}
			if !c.conf.allowAllOrigins {
				rtx.Response.Header.Add("Vary", "Access-Control-Request-Method")
				rtx.Response.Header.Add("Vary", "Access-Control-Request-Headers")
			}
			rtx.SetStatusCode(http.StatusOK)
			return nil, nil
		} else {
			if c.conf.AllowCredentials {
				rtx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
			}
			if len(c.conf.ExposeHeaders) > 0 {
				rtx.Response.Header.Set("Access-Control-Expose-Headers", strings.Join(c.conf.ExposeHeaders, ","))
			}
		}
		return handler(ctx, req)
	}
}

func (c Cors) isOriginValid(origin string) bool {
	if c.conf.allowAllOrigins {
		return true
	}
	for _, value := range c.conf.AllowOrigins {
		if value == origin {
			return true
		}
	}
	return false
}
