package xhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/server"
	_ "github.com/asjard/asjard/pkg/config/mem"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

type echoServer interface {
	Hello(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error)
}

type echoAPI struct{}

func (echoAPI) Hello(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return in, nil
}

func (e *echoAPI) RestServiceDesc() *rest.ServiceDesc {
	return &rest.ServiceDesc{
		Name:        "Echo",
		ServiceName: "api.v1.Echo",
		HandlerType: (*echoServer)(nil),
		Methods: []rest.MethodDesc{
			{
				MethodName: "Hello",
				Name:       "Hello",
				Method:     http.MethodGet,
				Path:       "/hello/{name}",
				Handler: func(ctx context.Context, srv any, interceptor server.UnaryServerInterceptor) (any, error) {
					in := new(emptypb.Empty)
					if interceptor == nil {
						return srv.(echoServer).Hello(ctx, in)
					}
					info := &server.UnaryServerInfo{
						Server:     srv,
						FullMethod: "/api.v1.Echo/Hello",
						Protocol:   Protocol,
					}
					handler := func(ctx context.Context, req any) (any, error) {
						return srv.(echoServer).Hello(ctx, in)
					}
					return interceptor(ctx, in, info, handler)
				},
			},
			{
				MethodName: "Create",
				Name:       "Create",
				Method:     http.MethodPost,
				Path:       "/hello/{name}",
				Handler: func(ctx context.Context, srv any, interceptor server.UnaryServerInterceptor) (any, error) {
					return srv.(echoServer).Hello(ctx, new(emptypb.Empty))
				},
			},
		},
	}
}

func TestDefaultHTTPTimeouts(t *testing.T) {
	conf := DefaultConfig(Protocol)
	require.Equal(t, 30*time.Second, conf.Options.ReadTimeout.Duration)
	require.Equal(t, 10*time.Second, conf.Options.ReadHeaderTimeout.Duration)
	require.Equal(t, 60*time.Second, conf.Options.IdleTimeout.Duration)
}

func TestHttpServerContract(t *testing.T) {
	conf := DefaultConfig(Protocol)
	conf.Enabled = true
	conf.Addresses.Listen = "127.0.0.1:0"
	created, err := MustNew(conf, &server.ServerOptions{})
	require.NoError(t, err)
	s := created.(*HttpServer)
	require.Equal(t, Protocol, s.Protocol())
	require.True(t, s.Enabled())
	require.Equal(t, conf.Addresses, s.ListenAddresses())
	require.Error(t, s.AddHandler(struct{}{}))
	s.Stop()
}

func TestHttpStartRequiresAddress(t *testing.T) {
	created, err := MustNew(Config{}, &server.ServerOptions{})
	require.NoError(t, err)
	require.Error(t, created.Start(make(chan error, 1)))
	created.Stop()
}

func TestHttpRoutingAndErrors(t *testing.T) {
	s := newTestServer(t)
	require.NoError(t, s.AddHandler(&echoAPI{}))
	h := testHandler(s)

	t.Run("path param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/hello/codex", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Header().Get(headerContentType), MIME_JSON)
		var payload map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
		require.Equal(t, true, payload["success"])
	})

	t.Run("post path param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/hello/codex", nil)
		req.Header.Set(headerContentType, MIME_JSON)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/missing", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		var payload map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
		require.Equal(t, false, payload["success"])
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/hello/codex", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		var payload map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
		require.Equal(t, false, payload["success"])
	})
}

func TestOpenAPIHandlerRegistration(t *testing.T) {
	s := newTestServer(t)
	require.NoError(t, s.AddHandler(rest.NewOpenAPI(s.conf.Openapi, s.openapi)))
}

func TestCorsMiddleware(t *testing.T) {
	s := newTestServer(t)
	require.NoError(t, s.AddHandler(&echoAPI{}))
	h := testHandler(s)

	req := httptest.NewRequest(http.MethodOptions, "http://localhost/hello/codex", nil)
	req.Host = "localhost"
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, "https://example.com", rec.Header().Get(headerAccessControlAllowOrigin))
}

func newTestServer(t *testing.T) *HttpServer {
	t.Helper()
	conf := DefaultConfig(Protocol)
	conf.Enabled = true
	conf.Addresses.Listen = "127.0.0.1:0"
	created, err := MustNew(conf, &server.ServerOptions{})
	require.NoError(t, err)
	return created.(*HttpServer)
}

func testHandler(s *HttpServer) http.Handler {
	return s.applyMiddleware(&router{
		mux:              s.mux,
		notFound:         s.newNoInterceptorHandler(rest.ErrorHandler_NotFound_RestHandler, s.errorHandler, DefaultWriterName),
		methodNotAllowed: s.newNoInterceptorHandler(rest.ErrorHandler_MethodNotAllowed_RestHandler, s.errorHandler, DefaultWriterName),
	}, s.middlewares...)
}
