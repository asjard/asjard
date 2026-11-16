package rest

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/utils"
	openapi_v3 "github.com/google/gnostic/openapiv3"
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
type MiddlewareFunc func(next http.Handler) http.Handler

// RestServer manages the rest instance, routing, and server lifecycle.
type RestServer struct {
	mux          *http.ServeMux                // Handles request multiplexing.
	server       *http.Server                  // The high-performance HTTP engine.
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
	conf := DefaultConfig(Protocol)
	// Fetch configuration from the central config store.
	if err := config.GetWithUnmarshal(constant.ConfigServerRestPrefix, &conf); err != nil {
		return nil, err
	}
	return MustNew(conf, options)
}

// MustNew performs the low-level setup of the rest engine and global CORS middleware.
func MustNew(conf Config, options *server.ServerOptions) (server.Server, error) {
	if conf.CertFile != "" {
		conf.CertFile = filepath.Join(utils.GetCertDir(), conf.CertFile)
	}
	if conf.KeyFile != "" {
		conf.KeyFile = filepath.Join(utils.GetCertDir(), conf.KeyFile)
	}

	s := &RestServer{
		mux:          http.NewServeMux(),
		openapi:      &openapi_v3.Document{},
		interceptor:  options.Interceptor,
		conf:         conf,
		middlewares:  []MiddlewareFunc{NewCorsMiddleware(conf.Cors)},
		errorHandler: &ErrorHandlerAPI{},
		server: &http.Server{
			Addr:              conf.Addresses.Listen,
			ReadTimeout:       conf.Options.ReadTimeout.Duration,
			ReadHeaderTimeout: conf.Options.ReadHeaderTimeout.Duration,
			WriteTimeout:      conf.Options.WriteTimeout.Duration,
			IdleTimeout:       conf.Options.IdleTimeout.Duration,
			MaxHeaderBytes:    conf.Options.MaxHeaderBytes,
			ErrorLog:          log.New(logWriter{}, "", 0),
		},
	}
	if conf.Options.DisableKeepalive {
		s.server.SetKeepAlivesEnabled(false)
	}
	return s, nil
}

// AddHandler registers a service to the server and populates the router.
func (s *RestServer) AddHandler(handler any) error {
	h, ok := handler.(Handler)
	if !ok {
		return fmt.Errorf("invlaid handler %T, must implement *rest.RestServiceDesc", handler)
	}
	s.handlers = append(s.handlers, h)
	return s.addRouter(h)
}

// Start launches the server listener in a background routine.
func (s *RestServer) Start(startErr chan error) error {
	// Optionally enable built-in OpenAPI and Route discovery services.
	if s.conf.Openapi.Enabled {
		s.AddHandler(NewOpenAPI(s.conf.Openapi, s.openapi))
	}
	if s.conf.Routes.Enabled {
		s.AddHandler(NewRoutesAPI(s.handlers))
	}

	router := &router{
		mux:              s.mux,
		notFound:         s.newNoInterceptorHandler(_ErrorHandler_NotFound_RestHandler, s.errorHandler, DefaultWriterName),
		methodNotAllowed: s.newNoInterceptorHandler(_ErrorHandler_MethodNotAllowed_RestHandler, s.errorHandler, DefaultWriterName),
	}
	s.server.Handler = s.applyMiddleware(router, s.middlewares...)

	if s.conf.Addresses.Listen == "" {
		return errors.New("config addresses.listen not found")
	}

	go func() {
		if s.conf.CertFile != "" && s.conf.KeyFile != "" {
			if err := s.server.ListenAndServeTLS(s.conf.CertFile, s.conf.KeyFile); err != nil {
				startErr <- fmt.Errorf("start https server with address %s fail %s", s.conf.Addresses.Listen, err.Error())
				return
			}
		} else if err := s.server.ListenAndServe(); err != nil {
			startErr <- fmt.Errorf("start http server with address %s fail %s", s.conf.Addresses.Listen, err.Error())
			return
		}
	}()
	logger.Debug("start rest server", "address", s.conf.Addresses.Listen)
	return nil
}

// Stop performs a graceful shutdown with a 3-second timeout.
func (s *RestServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s.server.Shutdown(ctx)
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

// Handle registers a raw HTTP handler on the server mux.
func (s *RestServer) Handle(pattern string, handler http.Handler) {
	if s.mux == nil || handler == nil || pattern == "" {
		return
	}
	s.mux.Handle(pattern, handler)
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
	pattern := methodDesc.Method + " " + methodDesc.Path
	s.mux.Handle(pattern, s.applyMiddleware(s.newHandler(methodDesc.Handler, svc, writerName),
		s.middlewares...))
}

// newHandler wraps the business logic in a REST context and response writer.
func (s *RestServer) newHandler(methodHandler methodHandler, svc Handler, writerName string) http.Handler {
	writer := GetWriter(writerName)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cc := NewContext(r, w, WithErrPage(s.conf.Doc.ErrPage), WithWriter(writer))
		reply, err := methodHandler(cc, svc, s.interceptor)
		cc.WriteData(reply, err)
	})
}

// newNoInterceptorHandler wraps the business logic in a REST context and response writer without interceptor.
func (s *RestServer) newNoInterceptorHandler(methodHandler methodHandler, svc Handler, writerName string) http.Handler {
	writer := GetWriter(writerName)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cc := NewContext(r, w, WithErrPage(s.conf.Doc.ErrPage), WithWriter(writer))
		reply, err := methodHandler(cc,
			svc,
			s.emptyInterceptor)
		cc.WriteData(reply, err)
	})
}

func (s *RestServer) emptyInterceptor(ctx context.Context, req any, _ *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
	return handler(ctx, req)
}

// applyMiddleware chains multiple middleware functions into a single handler.
func (s *RestServer) applyMiddleware(h http.Handler, middlewares ...MiddlewareFunc) http.Handler {
	for i := range len(middlewares) {
		h = middlewares[i](h)
	}
	return h
}

type router struct {
	mux              *http.ServeMux
	notFound         http.Handler
	methodNotAllowed http.Handler
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	_, pattern := r.mux.Handler(req)
	if pattern == "" {
		r.notFound.ServeHTTP(w, req)
		return
	}
	if r.isMethodNotAllowed(pattern, req.Method) {
		r.methodNotAllowed.ServeHTTP(w, req)
		return
	}
	// ServeMux.Handler does not populate named path wildcards.
	// ServeHTTP fills Request.Pattern and PathValue for {param} matches.
	r.mux.ServeHTTP(w, req)
}
func (r *router) isMethodNotAllowed(pattern, method string) bool {
	routeMethod, _, found := strings.Cut(pattern, " ")
	if !found || strings.HasPrefix(routeMethod, "/") {
		return false
	}
	if routeMethod == method {
		return false
	}
	// ServeMux serves GET handlers for HEAD requests.
	if method == http.MethodHead && routeMethod == http.MethodGet {
		return false
	}
	return true
}
