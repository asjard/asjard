package rest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/status"
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
	RestServiceDesc() ServiceDesc
}

// Writer 结果输出
type Writer func(ctx *Context, data any, err error)

// RestServer .
type RestServer struct {
	addresses   map[string]string
	router      *router.Router
	server      fasthttp.Server
	certFile    string
	keyFile     string
	enabled     bool
	interceptor server.UnaryServerInterceptor
}

var _ server.Server = &RestServer{}

func init() {
	server.AddServer(Protocol, New)
}

// New 初始化服务
// TODO 使用options的方式带参数, 配置写在结构体中，并通过GetWithUnmarshal反序列化进去
func New(interceptor server.UnaryServerInterceptor) (server.Server, error) {
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
		addresses:   addressesMap,
		router:      router.New(),
		certFile:    certFile,
		keyFile:     keyFile,
		enabled:     config.GetBool("servers.rest.enabled", false),
		interceptor: interceptor,
		server: fasthttp.Server{
			Name:            runtime.APP,
			Concurrency:     config.GetInt("servers.rest.options.Concurrency", fasthttp.DefaultConcurrency),
			ReadBufferSize:  config.GetInt("servers.rest.options.ReadBufferSize", defaultReadBufferSize),
			WriteBufferSize: config.GetInt("servers.rest.options.ReadBufferSize", defaultWriteBufferSize),
			ReadTimeout:     config.GetDuration("servers.rest.options.ReadTimeout", time.Second*3),
			WriteTimeout:    config.GetDuration("servers.rest.options.WriteTimeout", time.Second*3),
			IdleTimeout: config.GetDuration("servers.rest.options.WriteTimeout",
				config.GetDuration("servers.rest.options.ReadTimeout", time.Second*3)),
			MaxConnsPerIP:                      config.GetInt("servers.rest.options.WriteTimeout", 0),
			MaxRequestsPerConn:                 config.GetInt("servers.rest.options.MaxRequestsPerConn", 0),
			MaxIdleWorkerDuration:              config.GetDuration("servers.rest.options.MaxIdleWorkerDuration", time.Minute*10),
			TCPKeepalivePeriod:                 config.GetDuration("servers.rest.options.TCPKeepalivePeriod", 0),
			MaxRequestBodySize:                 config.GetInt("servers.rest.options.MaxRequestBodySize", fasthttp.DefaultMaxRequestBodySize),
			DisableKeepalive:                   config.GetBool("servers.rest.options.DisableKeepalive", false),
			TCPKeepalive:                       config.GetBool("servers.rest.options.TCPKeepalive", false),
			ReduceMemoryUsage:                  config.GetBool("servers.rest.options.ReduceMemoryUsage", false),
			GetOnly:                            config.GetBool("servers.rest.options.GetOnly", false),
			DisablePreParseMultipartForm:       config.GetBool("servers.rest.options.DisablePreParseMultipartForm", true),
			LogAllErrors:                       config.GetBool("servers.rest.options.LogAllErrors", false),
			SecureErrorLogMessage:              config.GetBool("servers.rest.options.SecureErrorLogMessage", false),
			DisableHeaderNamesNormalizing:      config.GetBool("servers.rest.options.DisableHeaderNamesNormalizing", false),
			SleepWhenConcurrencyLimitsExceeded: config.GetDuration("servers.rest.options.SleepWhenConcurrencyLimitsExceeded", 0),
			NoDefaultServerHeader:              config.GetBool("servers.rest.options.NoDefaultServerHeader", false),
			NoDefaultDate:                      config.GetBool("servers.rest.options.NoDefaultDate", false),
			NoDefaultContentType:               config.GetBool("servers.rest.options.NoDefaultContentType", false),
			KeepHijackedConns:                  config.GetBool("servers.rest.options.KeepHijackedConns", false),
			CloseOnShutdown:                    config.GetBool("servers.rest.options.CloseOnShutdown", false),
			StreamRequestBody:                  config.GetBool("servers.rest.options.StreamRequestBody", false),
			Logger:                             &Logger{},
		},
	}
	return server, nil
}

// AddHandler .
func (s *RestServer) AddHandler(handler any) error {
	h, ok := handler.(Handler)
	if !ok {
		return fmt.Errorf("invlaid handler, %v must implement *rest.Handler", reflect.TypeOf(handler))
	}
	if err := s.addRouter(h); err != nil {
		return err
	}
	return nil
}

// Start .
func (s *RestServer) Start(startErr chan error) error {
	s.server.ErrorHandler = func(ctx *fasthttp.RequestCtx, err error) {
		logger.Error("request fail",
			"method", ctx.Method(),
			"path", ctx.Path(),
			"err", err)
		NewContext(ctx).Write(nil, status.ErrInterServerError)
	}
	s.router.NotFound = func(ctx *fasthttp.RequestCtx) {
		NewContext(ctx).Write(nil, status.ErrNotFound)
	}
	s.router.MethodNotAllowed = func(ctx *fasthttp.RequestCtx) {
		NewContext(ctx).Write(nil, status.ErrMethodNotAllowed)
	}
	s.router.PanicHandler = func(ctx *fasthttp.RequestCtx, err any) {
		logger.Error("request panic",
			"method", ctx.Method(),
			"path", ctx.Path(),
			"err", err,
			"stack", string(debug.Stack()))
		NewContext(ctx).Write(nil, status.ErrInterServerError)
	}
	s.server.Handler = s.router.Handler
	address, ok := s.addresses[constant.ServerListenAddressName]
	if !ok {
		return errors.New("config servces.rest.addresses.listen not found")
	}
	if strings.HasPrefix(address, "unix") {
		go func() {
			if err := s.server.ListenAndServeUNIX(address, os.ModeSocket); err != nil {
				startErr <- fmt.Errorf("start rest server with address %s fail %s",
					address, err.Error())
			}
		}()
	} else if s.certFile != "" && s.keyFile != "" {
		go func() {
			if err := s.server.ListenAndServeTLS(address, s.certFile, s.keyFile); err != nil {
				startErr <- fmt.Errorf("start rest server with address %s fail %s",
					address, err.Error())
			}
		}()

	} else {
		go func() {
			if err := s.server.ListenAndServe(address); err != nil {
				// return err
				startErr <- fmt.Errorf("start rest server with address %s fail %s",
					address, err.Error())
			}
		}()
	}
	logger.Debug("start rest server",
		"address", address)
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

// Enabled .
func (s *RestServer) Enabled() bool {
	return s.enabled
}

// ListenAddresses 监听地址列表
func (s *RestServer) ListenAddresses() map[string]string {
	return s.addresses
}

func (s *RestServer) addRouter(handler Handler) error {
	desc := handler.RestServiceDesc()
	for _, method := range desc.Methods {
		if method.Method != "" && method.Path != "" && method.Handler != nil {
			ht := reflect.TypeOf(desc.HandlerType).Elem()
			st := reflect.TypeOf(handler)
			if !st.Implements(ht) {
				return fmt.Errorf("found the handler of type %v that does not satisfy %v", st, ht)
			}
			s.addRouterHandler(method.Method, method, handler)
		}
	}
	return nil
}

func (s *RestServer) addRouterHandler(method string, methodDesc MethodDesc, svc Handler) {
	s.router.Handle(method, methodDesc.Path, func(ctx *fasthttp.RequestCtx) {
		cc := NewContext(ctx)
		reply, err := methodDesc.Handler(cc, svc, s.interceptor)
		cc.Write(reply, err)
	})
}
