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
	// Protocol is the identifier for the REST server implementation.
	Protocol = "rest"
)

// Handler defines the interface for services that provide REST descriptors.
type Handler interface {
	RestServiceDesc() *ServiceDesc
}

// MiddlewareFunc defines the standard signature for REST middleware.
type MiddlewareFunc func(next fasthttp.RequestHandler) fasthttp.RequestHandler

// RestServer manages the fasthttp instance, routing, and server lifecycle.
type RestServer struct {
	router       *router.Router                // Handles request multiplexing.
	server       fasthttp.Server               // The high-performance HTTP engine.
	openapi      *openapi_v3.Document          // Aggregated API documentation.
	interceptor  server.UnaryServerInterceptor // Global interceptor for processing logic.
	conf         Config                        // Server-specific configurations.
	middlewares  []MiddlewareFunc              // Chain of global middlewares.
	errorHandler *ErrorHandlerAPI              // Standardized error response handler.
	handlers     []Handler                     // List of registered service handlers.
}

// Ensure RestServer satisfies the core server interface.
var _ server.Server = &RestServer{}

func init() {
	// Register the REST protocol to the framework's server registry.
	server.AddServer(Protocol, New)
}

// New initializes the REST server by loading configurations and setting up certificates.
func New(options *server.ServerOptions) (server.Server, error) {
	conf := defaultConfig()
	// Fetch configuration from the central config store.
	if err := config.GetWithUnmarshal(constant.ConfigServerRestPrefix, &conf); err != nil {
		return nil, err
	}
	// Resolve certificate paths relative to the system's cert directory.
	if conf.CertFile != "" {
		conf.CertFile = filepath.Join(utils.GetCertDir(), conf.CertFile)
	}
	if conf.KeyFile != "" {
		conf.KeyFile = filepath.Join(utils.GetCertDir(), conf.KeyFile)
	}
	return MustNew(conf, options)
}

// MustNew performs the low-level setup of the fasthttp engine and global CORS middleware.
func MustNew(conf Config, options *server.ServerOptions) (server.Server, error) {
	r := router.New()
	r.SaveMatchedRoutePath = true
	corsMiddleware := NewCorsMiddleware(conf.Cors)

	// Ensure GlobalOPTIONS handles preflight requests via CORS middleware.
	r.GlobalOPTIONS = corsMiddleware(func(ctx *fasthttp.RequestCtx) {})

	return &RestServer{
		router:       r,
		openapi:      &openapi_v3.Document{},
		interceptor:  options.Interceptor,
		conf:         conf,
		middlewares:  []MiddlewareFunc{corsMiddleware},
		errorHandler: &ErrorHandlerAPI{},
		server: fasthttp.Server{
			// Extensive performance tuning parameters mapped from configuration.
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

// AddHandler registers a service to the server and populates the router.
func (s *RestServer) AddHandler(handler any) error {
	h, ok := handler.(Handler)
	if !ok {
		return fmt.Errorf("invlaid handler %T, must implement *rest.Handler", handler)
	}
	s.handlers = append(s.handlers, h)
	return s.addRouter(h)
}

// Start launches the server listener (Unix, TLS, or Standard TCP) in a background routine.
func (s *RestServer) Start(startErr chan error) error {
	// Register system-level error handlers for the router.
	s.router.NotFound = s.newHandler(_ErrorHandler_NotFound_RestHandler, s.errorHandler, DefaultWriterName)
	s.router.MethodNotAllowed = s.newHandler(_ErrorHandler_MethodNotAllowed_RestHandler, s.errorHandler, DefaultWriterName)

	// Custom error handler for low-level protocol errors.
	s.server.ErrorHandler = func(ctx *fasthttp.RequestCtx, err error) {
		logger.L(ctx).Error("request error", "method", string(ctx.Method()), "path", string(ctx.Path()), "err", err.Error())
		cc := NewContext(ctx, WithErrPage(s.conf.Doc.ErrPage))
		cc.WriteData(_ErrorHandler_Error_RestHandler(cc, s.errorHandler, s.interceptor))
	}

	// Optionally enable built-in OpenAPI and Route discovery services.
	if s.conf.Openapi.Enabled {
		s.AddHandler(NewOpenAPI(s.conf.Openapi, s.openapi))
	}
	routesAPI := NewRoutesAPI(s.handlers)
	if s.conf.Routes.Enabled {
		s.AddHandler(routesAPI)
	}

	s.server.Handler = s.router.Handler
	if s.conf.Addresses.Listen == "" {
		return errors.New("config servces.rest.addresses.listen not found")
	}

	// Listener implementation selection.
	if strings.HasPrefix(s.conf.Addresses.Listen, "unix") {
		go func() {
			if err := s.server.ListenAndServeUNIX(s.conf.Addresses.Listen, os.ModeSocket); err != nil {
				startErr <- fmt.Errorf("start rest server with address %s fail %s", s.conf.Addresses.Listen, err.Error())
			}
		}()
	} else if s.conf.CertFile != "" && s.conf.KeyFile != "" {
		go func() {
			if err := s.server.ListenAndServeTLS(s.conf.Addresses.Listen, s.conf.CertFile, s.conf.KeyFile); err != nil {
				startErr <- fmt.Errorf("start rest server with address %s fail %s", s.conf.Addresses.Listen, err.Error())
			}
		}()
	} else {
		go func() {
			if err := s.server.ListenAndServe(s.conf.Addresses.Listen); err != nil {
				startErr <- fmt.Errorf("start rest server with address %s fail %s", s.conf.Addresses.Listen, err.Error())
			}
		}()
	}
	logger.Debug("start rest server", "address", s.conf.Addresses.Listen)
	return nil
}

// Stop performs a graceful shutdown with a 3-second timeout.
func (s *RestServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s.server.ShutdownWithContext(ctx)
}

// Protocol returns the server protocol name.
func (s *RestServer) Protocol() string {
	return Protocol
}

// Enabled checks if the REST server is enabled in config.
func (s *RestServer) Enabled() bool {
	return s.conf.Enabled
}

// ListenAddresses returns the configured listening addresses.
func (s *RestServer) ListenAddresses() server.AddressConfig {
	return s.conf.Addresses
}

// addRouter parses the service descriptor to register paths and merge OpenAPI docs.
func (s *RestServer) addRouter(handler Handler) error {
	desc := handler.RestServiceDesc()
	if desc == nil {
		return nil
	}
	// Merge individual service OpenAPI specs into the global document.
	if s.conf.Openapi.Enabled && len(desc.OpenAPI) != 0 {
		document := &openapi_v3.Document{}
		if err := proto.Unmarshal(desc.OpenAPI, document); err != nil {
			return err
		}
		proto.Merge(s.openapi, document)
	}
	// Type validation using reflection.
	ht := reflect.TypeOf(desc.HandlerType).Elem()
	st := reflect.TypeOf(handler)
	if !st.Implements(ht) {
		return fmt.Errorf("found the handler of type %v that does not satisfy %v", st, ht)
	}
	// Register each method path to the router.
	for _, method := range desc.Methods {
		if method.Method != "" && method.Path != "" && method.Handler != nil {
			s.addRouterHandler(method.Method, method, handler, method.WriterName)
		}
	}
	return nil
}

// addRouterHandler applies middleware and registers a handler to a specific route.
func (s *RestServer) addRouterHandler(method string, methodDesc MethodDesc, svc Handler, writerName string) {
	s.router.Handle(method, methodDesc.Path,
		s.applyMiddleware(s.newHandler(methodDesc.Handler, svc, writerName),
			s.middlewares...))
}

// newHandler wraps the business logic in a REST context and response writer.
func (s *RestServer) newHandler(methodHandler methodHandler, svc Handler, writerName string) fasthttp.RequestHandler {
	writer := GetWriter(writerName)
	return func(ctx *fasthttp.RequestCtx) {
		cc := NewContext(ctx, WithErrPage(s.conf.Doc.ErrPage), WithWriter(writer))
		reply, err := methodHandler(cc, svc, s.interceptor)
		cc.WriteData(reply, err)
	}
}

// applyMiddleware chains multiple middleware functions into a single handler.
func (s *RestServer) applyMiddleware(h fasthttp.RequestHandler, middlewares ...MiddlewareFunc) fasthttp.RequestHandler {
	for i := 0; i < len(middlewares); i++ {
		h = middlewares[i](h)
	}
	return h
}
