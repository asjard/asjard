/*
Package http provides the net/http implementation of the server interface.
It mirrors pkg/server/rest, but uses Go's standard library HTTP server and ServeMux
instead of fasthttp.
*/
package xhttp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/utils"
	openapi_v3 "github.com/google/gnostic/openapiv3"
	"google.golang.org/protobuf/proto"
)

const (
	// Protocol is the identifier for the HTTP server implementation.
	Protocol = "http"
)

// MiddlewareFunc defines the standard signature for HTTP middleware.
type MiddlewareFunc func(next http.Handler) http.Handler

// HttpServer manages the net/http instance, routing, and server lifecycle.
type HttpServer struct {
	mux          *http.ServeMux
	server       *http.Server
	openapi      *openapi_v3.Document
	interceptor  server.UnaryServerInterceptor
	conf         Config
	middlewares  []MiddlewareFunc
	errorHandler *rest.ErrorHandlerAPI
	handlers     []rest.Handler
}

var _ server.Server = &HttpServer{}

func init() {
	server.AddServer(Protocol, New)
}

// New initializes the HTTP server by loading configurations and setting up certificates.
func New(options *server.ServerOptions) (server.Server, error) {
	conf := DefaultConfig(Protocol)
	if err := config.GetWithUnmarshal(constant.ConfigServerHttpPrefix, &conf); err != nil {
		return nil, err
	}

	return MustNew(conf, options)
}

// MustNew performs the low-level setup of the net/http engine and global CORS middleware.
func MustNew(conf Config, options *server.ServerOptions) (server.Server, error) {
	if conf.CertFile != "" {
		conf.CertFile = filepath.Join(utils.GetCertDir(), conf.CertFile)
	}
	if conf.KeyFile != "" {
		conf.KeyFile = filepath.Join(utils.GetCertDir(), conf.KeyFile)
	}
	s := &HttpServer{
		mux:          http.NewServeMux(),
		openapi:      &openapi_v3.Document{},
		interceptor:  options.Interceptor,
		conf:         conf,
		middlewares:  []MiddlewareFunc{NewCorsMiddleware(conf.Cors)},
		errorHandler: &rest.ErrorHandlerAPI{},
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

// Handle registers a raw HTTP handler on the server mux.
func (s *HttpServer) Handle(pattern string, handler http.Handler) {
	if s.mux == nil || handler == nil || pattern == "" {
		return
	}
	s.mux.Handle(pattern, handler)
}

// AddHandler registers a service to the server and populates the router.
func (s *HttpServer) AddHandler(handler any) error {
	h, ok := handler.(rest.Handler)
	if !ok {
		return fmt.Errorf("invalid handler %T, must implement *rest.RestServiceDesc", handler)
	}
	s.handlers = append(s.handlers, h)
	return s.addRouter(h)
}

// Start launches the server listener (Unix, TLS, or Standard TCP) in a background routine.
func (s *HttpServer) Start(startErr chan error) error {
	if s.conf.Openapi.Enabled {
		if err := s.AddHandler(rest.NewOpenAPI(s.conf.Openapi, s.openapi)); err != nil {
			return err
		}
	}
	if s.conf.Routes.Enabled {
		if err := s.AddHandler(rest.NewRoutesAPI(s.handlers)); err != nil {
			return err
		}
	}

	router := &router{
		mux:              s.mux,
		notFound:         s.newNoInterceptorHandler(rest.ErrorHandler_NotFound_RestHandler, s.errorHandler, DefaultWriterName),
		methodNotAllowed: s.newNoInterceptorHandler(rest.ErrorHandler_MethodNotAllowed_RestHandler, s.errorHandler, DefaultWriterName),
	}
	s.server.Handler = s.applyMiddleware(router, s.middlewares...)

	if s.conf.Addresses.Listen == "" {
		return errors.New("config servers.http.addresses.listen not found")
	}

	network, address := listenNetwork(s.conf.Addresses.Listen)
	go func() {
		var err error
		switch {
		case network == "unix":
			_ = os.Remove(address)
			ln, listenErr := net.Listen("unix", address)
			if listenErr != nil {
				startErr <- fmt.Errorf("start http server with address %s fail %s", s.conf.Addresses.Listen, listenErr.Error())
				return
			}
			if chmodErr := os.Chmod(address, os.ModeSocket); chmodErr != nil {
				logger.Error("chmod unix socket fail", "address", address, "err", chmodErr)
			}
			err = s.server.Serve(ln)
		case s.conf.CertFile != "" && s.conf.KeyFile != "":
			err = s.server.ListenAndServeTLS(s.conf.CertFile, s.conf.KeyFile)
		default:
			ln, listenErr := net.Listen("tcp", address)
			if listenErr != nil {
				startErr <- fmt.Errorf("start http server with address %s fail %s", s.conf.Addresses.Listen, listenErr.Error())
				return
			}
			err = s.server.Serve(ln)
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			startErr <- fmt.Errorf("start http server with address %s fail %s", s.conf.Addresses.Listen, err.Error())
		}
	}()
	logger.Debug("start http server", "address", s.conf.Addresses.Listen)
	return nil
}

// Stop performs a graceful shutdown with a 3-second timeout.
func (s *HttpServer) Stop() {
	if s.server == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = s.server.Shutdown(ctx)
}

// Protocol returns the server protocol name.
func (s *HttpServer) Protocol() string {
	return Protocol
}

// Enabled checks if the HTTP server is enabled in config.
func (s *HttpServer) Enabled() bool {
	return s.conf.Enabled
}

// ListenAddresses returns the configured listening addresses.
func (s *HttpServer) ListenAddresses() server.AddressConfig {
	return s.conf.Addresses
}

func (s *HttpServer) addRouter(handler rest.Handler) error {
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

func (s *HttpServer) addRouterHandler(method string, methodDesc rest.MethodDesc, svc rest.Handler, writerName string) {
	pattern := routePattern(method, methodDesc.Path)
	s.mux.Handle(pattern, s.newHandler(methodDesc.Handler, svc, writerName))
}

func (s *HttpServer) newHandler(methodHandler rest.MethodHandler, svc rest.Handler, writerName string) http.Handler {
	writer := GetWriter(writerName)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.conf.Options.MaxRequestBodySize > 0 && r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, int64(s.conf.Options.MaxRequestBodySize))
		}
		cc := NewContext(NewRequestCtx(w, r), WithErrPage(s.conf.Doc.ErrPage), WithI18n(s.conf.I18n.Enabled), WithWriter(writer))
		reply, err := methodHandler(cc, svc, s.interceptor)
		cc.WriteData(reply, err)
	})
}

func (s *HttpServer) newNoInterceptorHandler(methodHandler rest.MethodHandler, svc rest.Handler, writerName string) http.Handler {
	writer := GetWriter(writerName)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cc := NewContext(NewRequestCtx(w, r), WithErrPage(s.conf.Doc.ErrPage), WithI18n(s.conf.I18n.Enabled), WithWriter(writer))
		reply, err := methodHandler(cc, svc, s.emptyInterceptor)
		cc.WriteData(reply, err)
	})
}

func (s *HttpServer) emptyInterceptor(ctx context.Context, req any, _ *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
	return handler(ctx, req)
}

func (s *HttpServer) applyMiddleware(h http.Handler, middlewares ...MiddlewareFunc) http.Handler {
	for i := range middlewares {
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
	h, pattern := r.mux.Handler(req)
	if pattern == "" {
		r.notFound.ServeHTTP(w, req)
		return
	}
	if isMethodNotAllowed(pattern, req.Method) {
		r.methodNotAllowed.ServeHTTP(w, req)
		return
	}
	h.ServeHTTP(w, req)
}
