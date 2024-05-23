package rest

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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
	ctx.Close()
}

func init() {
	server.AddServer(New)
}

// SetErrorHandler 设置错误处理
func SetErrorHandler(errHandler ErrorHandler) {
	errorHandler = func(ctx *Context, err error) {
		errHandler(ctx, err)
		ctx.Close()
	}
}

// New .
func New() (server.Server, error) {
	addressesMap := make(map[string]string)
	if err := config.GetWithUnmarshal("servers.rest.addresses", &addressesMap); err != nil {
		return nil, err
	}
	certFile := config.GetString("servers.rest.certFile", "")
	if certFile != "" {
		certFile = filepath.Join(utils.GetCertDir(), certFile)
	}
	keyFile := config.GetString("servers.rest.keyFile", "")
	if keyFile != "" {
		keyFile = filepath.Join(utils.GetCertDir(), keyFile)
	}
	server := &RestServer{
		addresses: addressesMap,
		router:    router.New(),
		certFile:  certFile,
		keyFile:   keyFile,
		server: fasthttp.Server{
			Name:                               config.GetString("servers.rest.name", constant.Framework),
			Concurrency:                        config.GetInt("servers.rest.Concurrency", fasthttp.DefaultConcurrency),
			ReadBufferSize:                     config.GetInt("servers.rest.ReadBufferSize", defaultReadBufferSize),
			WriteBufferSize:                    config.GetInt("servers.rest.ReadBufferSize", defaultWriteBufferSize),
			ReadTimeout:                        config.GetDuration("servers.rest.ReadTimeout", time.Second*3),
			WriteTimeout:                       config.GetDuration("servers.rest.WriteTimeout", time.Second*3),
			IdleTimeout:                        config.GetDuration("servers.rest.WriteTimeout", config.GetDuration("servers.rest.ReadTimeout", time.Second*3)),
			MaxConnsPerIP:                      config.GetInt("servers.rest.WriteTimeout", 0),
			MaxRequestsPerConn:                 config.GetInt("servers.rest.MaxRequestsPerConn", 0),
			MaxIdleWorkerDuration:              config.GetDuration("servers.rest.MaxIdleWorkerDuration", time.Minute*10),
			TCPKeepalivePeriod:                 config.GetDuration("servers.rest.TCPKeepalivePeriod", 0),
			MaxRequestBodySize:                 config.GetInt("servers.rest.MaxRequestBodySize", fasthttp.DefaultMaxRequestBodySize),
			DisableKeepalive:                   config.GetBool("servers.rest.DisableKeepalive", false),
			TCPKeepalive:                       config.GetBool("servers.rest.TCPKeepalive", false),
			ReduceMemoryUsage:                  config.GetBool("servers.rest.ReduceMemoryUsage", false),
			GetOnly:                            config.GetBool("servers.rest.GetOnly", false),
			DisablePreParseMultipartForm:       config.GetBool("servers.rest.DisablePreParseMultipartForm", true),
			LogAllErrors:                       config.GetBool("servers.rest.LogAllErrors", false),
			SecureErrorLogMessage:              config.GetBool("servers.rest.SecureErrorLogMessage", false),
			DisableHeaderNamesNormalizing:      config.GetBool("servers.rest.DisableHeaderNamesNormalizing", false),
			SleepWhenConcurrencyLimitsExceeded: config.GetDuration("servers.rest.SleepWhenConcurrencyLimitsExceeded", 0),
			NoDefaultServerHeader:              config.GetBool("servers.rest.NoDefaultServerHeader", false),
			NoDefaultDate:                      config.GetBool("servers.rest.NoDefaultDate", false),
			NoDefaultContentType:               config.GetBool("servers.rest.NoDefaultContentType", false),
			KeepHijackedConns:                  config.GetBool("servers.rest.KeepHijackedConns", false),
			CloseOnShutdown:                    config.GetBool("servers.rest.CloseOnShutdown", false),
			StreamRequestBody:                  config.GetBool("servers.rest.StreamRequestBody", false),
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
			ht := reflect.TypeOf(desc.HandlerType).Elem()
			st := reflect.TypeOf(handler)
			if !st.Implements(ht) {
				return fmt.Errorf("found the handler of type %v that does not satisfy %v", st, ht)
			}
			s.router.Handle(method.Method, method.Path, method.Handler(handler))
		}
	}
	return nil
}
