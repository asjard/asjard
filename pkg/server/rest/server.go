package rest

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/utils"
	"github.com/fasthttp/router"
	openapi_v3 "github.com/google/gnostic/openapiv3"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/proto"
)

const (
	// Protocol 协议名称
	Protocol = "rest"
)

// Handler .
type Handler interface {
	RestServiceDesc() *ServiceDesc
}

type MiddlewareFunc func(next fasthttp.RequestHandler) fasthttp.RequestHandler

// RestServer .
type RestServer struct {
	router       *router.Router
	server       fasthttp.Server
	openapi      *openapi_v3.Document
	interceptor  server.UnaryServerInterceptor
	conf         Config
	middlewares  []MiddlewareFunc
	errorHandler *ErrorHandlerAPI
}

var _ server.Server = &RestServer{}

func init() {
	server.AddServer(Protocol, New)
}

// New 初始化服务
func New(options *server.ServerOptions) (server.Server, error) {
	conf := defaultConfig()
	if err := config.GetWithUnmarshal(constant.ConfigServerRestPrefix, &conf); err != nil {
		return nil, err
	}
	if conf.CertFile != "" {
		conf.CertFile = filepath.Join(utils.GetCertDir(), conf.CertFile)
	}
	if conf.KeyFile != "" {
		conf.KeyFile = filepath.Join(utils.GetCertDir(), conf.KeyFile)
	}
	return MustNew(conf, options)
}

// MustNew 配置文件初始化
func MustNew(conf Config, options *server.ServerOptions) (server.Server, error) {
	r := router.New()
	r.SaveMatchedRoutePath = true
	corsMiddleware := NewCorsMiddleware(conf.Cors)
	// 不能删除这行，option请求走不到middleware中
	r.GlobalOPTIONS = corsMiddleware(func(ctx *fasthttp.RequestCtx) {})
	return &RestServer{
		router:       r,
		openapi:      &openapi_v3.Document{},
		interceptor:  options.Interceptor,
		conf:         conf,
		middlewares:  []MiddlewareFunc{corsMiddleware},
		errorHandler: &ErrorHandlerAPI{},
		server: fasthttp.Server{
			Name:                               runtime.GetAPP().App,
			Concurrency:                        conf.Options.Concurrency,
			ReadBufferSize:                     conf.Options.ReadBufferSize,
			WriteBufferSize:                    conf.Options.WriteBufferSize,
			ReadTimeout:                        conf.Options.ReadTimeout.Duration,
			WriteTimeout:                       conf.Options.WriteTimeout.Duration,
			IdleTimeout:                        conf.Options.IdleTimeout.Duration,
			MaxConnsPerIP:                      conf.Options.MaxConnsPerIP,
			MaxRequestsPerConn:                 conf.Options.MaxRequestsPerConn,
			MaxIdleWorkerDuration:              conf.Options.MaxIdleWorkerDuration.Duration,
			TCPKeepalivePeriod:                 conf.Options.TCPKeepalivePeriod.Duration,
			MaxRequestBodySize:                 conf.Options.MaxRequestBodySize,
			DisableKeepalive:                   conf.Options.DisableKeepalive,
			TCPKeepalive:                       conf.Options.TCPKeepalive,
			ReduceMemoryUsage:                  conf.Options.ReduceMemoryUsage,
			GetOnly:                            conf.Options.GetOnly,
			DisablePreParseMultipartForm:       conf.Options.DisablePreParseMultipartForm,
			LogAllErrors:                       conf.Options.LogAllErrors,
			SecureErrorLogMessage:              conf.Options.SecureErrorLogMessage,
			DisableHeaderNamesNormalizing:      conf.Options.DisableHeaderNamesNormalizing,
			SleepWhenConcurrencyLimitsExceeded: conf.Options.SleepWhenConcurrencyLimitsExceeded.Duration,
			NoDefaultServerHeader:              conf.Options.NoDefaultServerHeader,
			NoDefaultDate:                      conf.Options.NoDefaultDate,
			NoDefaultContentType:               conf.Options.NoDefaultContentType,
			KeepHijackedConns:                  conf.Options.KeepHijackedConns,
			CloseOnShutdown:                    conf.Options.CloseOnShutdown,
			StreamRequestBody:                  conf.Options.StreamRequestBody,
			Logger:                             &Logger{},
		},
	}, nil
}

// AddHandler .
func (s *RestServer) AddHandler(handler any) error {
	h, ok := handler.(Handler)
	if !ok {
		return fmt.Errorf("invlaid handler %T, must implement *rest.Handler", handler)
	}
	return s.addRouter(h)
}

// Start 启动rest服务
func (s *RestServer) Start(startErr chan error) error {
	s.router.NotFound = s.newHandler(_ErrorHandler_NotFound_RestHandler, s.errorHandler, DefaultWriterName)
	s.router.MethodNotAllowed = s.newHandler(_ErrorHandler_MethodNotAllowed_RestHandler, s.errorHandler, DefaultWriterName)
	s.server.ErrorHandler = func(ctx *fasthttp.RequestCtx, err error) {
		logger.Error("request error",
			"method", string(ctx.Method()),
			"path", string(ctx.Path()),
			"header", ctx.Request.Header.String(),
			"err", err)
		cc := NewContext(ctx, WithErrPage(s.conf.Doc.ErrPage))
		cc.WriteData(_ErrorHandler_Error_RestHandler(cc, s.errorHandler, s.interceptor))
	}
	if s.conf.Openapi.Enabled {
		// 添加openapi接口
		s.AddHandler(NewOpenAPI(s.conf.Openapi, s.openapi))
	}
	s.server.Handler = s.router.Handler
	if s.conf.Addresses.Listen == "" {
		return errors.New("config servces.rest.addresses.listen not found")
	}
	if strings.HasPrefix(s.conf.Addresses.Listen, "unix") {
		go func() {
			if err := s.server.ListenAndServeUNIX(s.conf.Addresses.Listen, os.ModeSocket); err != nil {
				startErr <- fmt.Errorf("start rest server with address %s fail %s",
					s.conf.Addresses.Listen, err.Error())
			}
		}()
	} else if s.conf.CertFile != "" && s.conf.KeyFile != "" {
		go func() {
			if err := s.server.ListenAndServeTLS(s.conf.Addresses.Listen, s.conf.CertFile, s.conf.KeyFile); err != nil {
				startErr <- fmt.Errorf("start rest server with address %s fail %s",
					s.conf.Addresses.Listen, err.Error())
			}
		}()

	} else {
		go func() {
			if err := s.server.ListenAndServe(s.conf.Addresses.Listen); err != nil {
				// return err
				startErr <- fmt.Errorf("start rest server with address %s fail %s",
					s.conf.Addresses.Listen, err.Error())
			}
		}()
	}
	logger.Debug("start rest server",
		"address", s.conf.Addresses.Listen)
	return nil
}

// Stop .
func (s *RestServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s.server.ShutdownWithContext(ctx)
}

// Protocol .
func (s *RestServer) Protocol() string {
	return Protocol
}

// Enabled .
func (s *RestServer) Enabled() bool {
	return s.conf.Enabled
}

// ListenAddresses 监听地址列表
func (s *RestServer) ListenAddresses() server.AddressConfig {
	return s.conf.Addresses
}

func (s *RestServer) addRouter(handler Handler) error {
	desc := handler.RestServiceDesc()
	if desc == nil {
		return nil
	}
	if s.conf.Openapi.Enabled && len(desc.OpenAPI) != 0 {
		document := &openapi_v3.Document{}
		if err := proto.Unmarshal(desc.OpenAPI, document); err != nil {
			return err
		}
		proto.Merge(s.openapi, document)
	}
	ht := reflect.TypeOf(desc.HandlerType).Elem()
	st := reflect.TypeOf(handler)
	if !st.Implements(ht) {
		return fmt.Errorf("found the handler of type %v that does not satisfy %v", st, ht)
	}
	for _, method := range desc.Methods {
		if method.Method != "" && method.Path != "" && method.Handler != nil {
			s.addRouterHandler(method.Method, method, handler, method.WriterName)
		}
	}
	return nil
}

func (s *RestServer) addRouterHandler(method string, methodDesc MethodDesc, svc Handler, writerName string) {
	s.router.Handle(method, methodDesc.Path,
		s.applyMiddleware(s.newHandler(methodDesc.Handler, svc, writerName),
			s.middlewares...))
}

func (s *RestServer) newHandler(methodHandler methodHandler, svc Handler, writerName string) fasthttp.RequestHandler {
	writer := GetWriter(writerName)
	return func(ctx *fasthttp.RequestCtx) {
		cc := NewContext(ctx, WithErrPage(s.conf.Doc.ErrPage), WithWriter(writer))
		reply, err := methodHandler(cc, svc, s.interceptor)
		cc.WriteData(reply, err)
	}
}

func (s *RestServer) applyMiddleware(h fasthttp.RequestHandler, middlewares ...MiddlewareFunc) fasthttp.RequestHandler {
	for i := 0; i < len(middlewares); i++ {
		h = middlewares[i](h)
	}
	return h
}
