package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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

// Writer 结果输出
type Writer func(ctx *Context, data any, err error)

// RestServer .
type RestServer struct {
	router      *router.Router
	server      fasthttp.Server
	openapi     *openapi_v3.Document
	interceptor server.UnaryServerInterceptor
	conf        ServerConfig
}

var _ server.Server = &RestServer{}

func init() {
	server.AddServer(Protocol, New)
}

func MustNew(conf ServerConfig, options *server.ServerOptions) (server.Server, error) {
	return &RestServer{
		router:      router.New(),
		openapi:     &openapi_v3.Document{},
		interceptor: options.Interceptor,
		conf:        conf,
		server: fasthttp.Server{
			Name:                               runtime.APP,
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

// New 初始化服务
func New(options *server.ServerOptions) (server.Server, error) {
	conf := defaultConfig
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
			"header", ctx.Request.Header.String(),
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
			"header", ctx.Request.Header.String(),
			"err", err,
			"stack", string(debug.Stack()))
		NewContext(ctx).Write(nil, status.ErrInterServerError)
	}
	s.router.GlobalOPTIONS = func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
		ctx.Response.Header.SetBytesV("Access-Control-Allow-Origin", ctx.Request.Header.Peek("Origin"))
	}
	if s.conf.Openapi {
		// 合并openapi并且
		s.addOpenapiHandler()
	}
	s.server.Handler = s.router.Handler
	address, ok := s.conf.Addresses[constant.ServerListenAddressName]
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
	} else if s.conf.CertFile != "" && s.conf.KeyFile != "" {
		go func() {
			if err := s.server.ListenAndServeTLS(address, s.conf.CertFile, s.conf.KeyFile); err != nil {
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
func (s *RestServer) ListenAddresses() map[string]string {
	return s.conf.Addresses
}

func (s *RestServer) addRouter(handler Handler) error {
	desc := handler.RestServiceDesc()
	if desc == nil {
		return nil
	}
	if s.conf.Openapi && len(desc.OpenAPI) != 0 {
		document := &openapi_v3.Document{}
		if err := proto.Unmarshal(desc.OpenAPI, document); err != nil {
			return err
		}
		proto.Merge(s.openapi, document)
	}
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
		cc := NewContext(ctx, WithErrPage(s.conf.Doc.ErrPage))
		reply, err := methodDesc.Handler(cc, svc, s.interceptor)
		cc.Write(reply, err)
	})
}

func (s *RestServer) addOpenapiHandler() {
	s.openapi.Info = &openapi_v3.Info{
		Title:       runtime.APP,
		Description: config.GetString(constant.ConfigDesc, ""),
		Version:     runtime.Version,
	}
	props := make([]*openapi_v3.NamedSchemaOrReference, 0, len(s.openapi.Components.Schemas.AdditionalProperties))
	propMap := make(map[string]struct{})
	for _, prop := range s.openapi.Components.Schemas.AdditionalProperties {
		if _, ok := propMap[prop.Name]; !ok {
			props = append(props, prop)
		}
		propMap[prop.Name] = struct{}{}
	}
	s.openapi.Components.Schemas.AdditionalProperties = props
	s.router.Handle(http.MethodGet, "/openapi.yml", func(ctx *fasthttp.RequestCtx) {
		b, err := s.openapi.YAMLValue("")
		if err != nil {
			return
		}
		ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
		ctx.Response.Header.SetBytesV("Access-Control-Allow-Origin", ctx.Request.Header.Peek("Origin"))
		ctx.Write(b)
	})
	address := s.conf.Addresses[constant.ServerAdvertiseAddressName]
	if address == "" {
		address = s.conf.Addresses[constant.ServerListenAddressName]
	}
	s.router.Handle(http.MethodGet, "/openapi", func(ctx *fasthttp.RequestCtx) {
		ctx.Redirect(fmt.Sprintf("https://authress-engineering.github.io/openapi-explorer/?url=http://%s/openapi.yml#?route=overview", address), http.StatusMovedPermanently)
	})
}
