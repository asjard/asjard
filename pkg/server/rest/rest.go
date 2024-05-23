package rest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/utils"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

const (
	// Protocol 协议名称
	Protocol = "rest"
)

const (
	defaultReadBufferSize  = 4096
	defaultWriteBufferSize = 4096
)

// Handler .
type Handler interface {
	// Routers() []*Router
	// Groups() []*Group
	ServiceDesc() ServiceDesc
}

// ErrorHandler 错误处理
type ErrorHandler func(ctx *Context, err error)

// RestServer .
type RestServer struct {
	addresses map[string]string
	router    *router.Router
	server    fasthttp.Server
	certFile  string
	keyFile   string
}

var _ server.Server = &RestServer{}

var errorHandler ErrorHandler = func(ctx *Context, err error) {
	DefaultErrorHandler(ctx, err)
	contextPool.Put(ctx)
}

func init() {
	server.AddServer(New)
}

// SetErrorHandler 设置错误处理
func SetErrorHandler(errHandler ErrorHandler) {
	errorHandler = func(ctx *Context, err error) {
		errHandler(ctx, err)
		contextPool.Put(ctx)
	}
}

// New .
func New() (server.Server, error) {
	addressesMap := make(map[string]string)
	if err := config.GetWithUnmarshal("servers.http.addresses", &addressesMap); err != nil {
		return nil, err
	}
	certFile := config.GetString("servers.http.certFile", "")
	if certFile != "" {
		certFile = filepath.Join(utils.GetCertDir(), certFile)
	}
	keyFile := config.GetString("servers.http.keyFile", "")
	if keyFile != "" {
		keyFile = filepath.Join(utils.GetCertDir(), keyFile)
	}
	server := &RestServer{
		addresses: addressesMap,
		router:    router.New(),
		certFile:  certFile,
		keyFile:   keyFile,
		server: fasthttp.Server{
			Name:                               config.GetString("servers.http.name", constant.Framework),
			Concurrency:                        config.GetInt("servers.http.Concurrency", fasthttp.DefaultConcurrency),
			ReadBufferSize:                     config.GetInt("servers.http.ReadBufferSize", defaultReadBufferSize),
			WriteBufferSize:                    config.GetInt("servers.http.ReadBufferSize", defaultWriteBufferSize),
			ReadTimeout:                        config.GetDuration("servers.http.ReadTimeout", time.Second*3),
			WriteTimeout:                       config.GetDuration("servers.http.WriteTimeout", time.Second*3),
			IdleTimeout:                        config.GetDuration("servers.http.WriteTimeout", config.GetDuration("servers.http.ReadTimeout", time.Second*3)),
			MaxConnsPerIP:                      config.GetInt("servers.http.WriteTimeout", 0),
			MaxRequestsPerConn:                 config.GetInt("servers.http.MaxRequestsPerConn", 0),
			MaxIdleWorkerDuration:              config.GetDuration("servers.http.MaxIdleWorkerDuration", time.Minute*10),
			TCPKeepalivePeriod:                 config.GetDuration("servers.http.TCPKeepalivePeriod", 0),
			MaxRequestBodySize:                 config.GetInt("servers.http.MaxRequestBodySize", fasthttp.DefaultMaxRequestBodySize),
			DisableKeepalive:                   config.GetBool("servers.http.DisableKeepalive", false),
			TCPKeepalive:                       config.GetBool("servers.http.TCPKeepalive", false),
			ReduceMemoryUsage:                  config.GetBool("servers.http.ReduceMemoryUsage", false),
			GetOnly:                            config.GetBool("servers.http.GetOnly", false),
			DisablePreParseMultipartForm:       config.GetBool("servers.http.DisablePreParseMultipartForm", true),
			LogAllErrors:                       config.GetBool("servers.http.LogAllErrors", false),
			SecureErrorLogMessage:              config.GetBool("servers.http.SecureErrorLogMessage", false),
			DisableHeaderNamesNormalizing:      config.GetBool("servers.http.DisableHeaderNamesNormalizing", false),
			SleepWhenConcurrencyLimitsExceeded: config.GetDuration("servers.http.SleepWhenConcurrencyLimitsExceeded", 0),
			NoDefaultServerHeader:              config.GetBool("servers.http.NoDefaultServerHeader", false),
			NoDefaultDate:                      config.GetBool("servers.http.NoDefaultDate", false),
			NoDefaultContentType:               config.GetBool("servers.http.NoDefaultContentType", false),
			KeepHijackedConns:                  config.GetBool("servers.http.KeepHijackedConns", false),
			CloseOnShutdown:                    config.GetBool("servers.http.CloseOnShutdown", false),
			StreamRequestBody:                  config.GetBool("servers.http.StreamRequestBody", false),
			Logger:                             &Logger{},
		},
	}
	return server, nil
}

// AddHandler .
func (s *RestServer) AddHandler(handler any) error {
	h, ok := handler.(Handler)
	if !ok {
		return fmt.Errorf("invlaid handler, must implement *rest.Handler")
	}
	if err := s.addRouter(h); err != nil {
		return err
	}
	return nil
}

// Start .
func (s *RestServer) Start() error {
	s.server.ErrorHandler = func(ctx *fasthttp.RequestCtx, err error) {
		logger.Errorf("request %s %s err: %v", ctx.Method(), ctx.Path(), err)
		errorHandler(NewContext(ctx), ErrInterServerError)
	}
	s.router.NotFound = func(ctx *fasthttp.RequestCtx) {
		errorHandler(NewContext(ctx), ErrNotFound)
	}
	s.router.MethodNotAllowed = func(ctx *fasthttp.RequestCtx) {
		errorHandler(NewContext(ctx), ErrMethodNotAllowed)
	}
	s.router.PanicHandler = func(ctx *fasthttp.RequestCtx, err interface{}) {
		logger.Errorf("request %s %s err: %v", ctx.Method(), ctx.Path(), err)
		errorHandler(NewContext(ctx), ErrInterServerError)
	}
	s.server.Handler = s.router.Handler
	for _, address := range s.addresses {
		if strings.HasPrefix(address, "unix") {
			if err := s.server.ListenAndServeUNIX(address, os.ModeSocket); err != nil {
				return err
			}
		} else if s.certFile != "" && s.keyFile != "" {
			if err := s.server.ListenAndServeTLS(address, s.certFile, s.keyFile); err != nil {
				return err
			}
		} else {
			if err := s.server.ListenAndServe(address); err != nil {
				return err
			}
		}
	}
	return nil

}

// Stop .
func (s *RestServer) Stop() {
	s.server.Shutdown()
}

// Protocol .
func (s *RestServer) Protocol() string {
	return Protocol
}

// ListenAddresses 监听地址列表
func (s *RestServer) ListenAddresses() []*server.EndpointAddress {
	var addresses []*server.EndpointAddress
	for name, address := range s.addresses {
		addresses = append(addresses, &server.EndpointAddress{
			Name:    name,
			Address: address,
		})
	}
	return addresses
}

func (s *RestServer) addRouter(handler Handler) error {
	desc := handler.ServiceDesc()
	for _, method := range desc.Methods {
		if method.Method != "" && method.Path != "" && method.Handler != nil {
			s.router.Handle(method.Method, method.Path, method.Handler(handler))
		}
	}
	return nil
}
