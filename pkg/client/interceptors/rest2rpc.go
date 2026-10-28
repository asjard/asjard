package interceptors

import (
	"context"
	"strings"
	"sync"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/pkg/client/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/utils"
	"github.com/spf13/cast"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/metadata"
)

const (
	// Rest2RpcInterceptorName is the unique identifier for this bridging interceptor.
	Rest2RpcInterceptorName = "rest2Rpc"
)

// Rest2Rpc handles the conversion of a REST-based context into a gRPC-compatible context.
// This should typically be positioned at the very beginning of the interceptor chain.
type Rest2Rpc struct {
	cfg rest2RpcConfig
	cm  sync.RWMutex // Protects the configuration during dynamic updates.
}

// rest2RpcContextConfig defines which headers are allowed to pass through the bridge.
type rest2RpcConfig struct {
	allowAllHeaders bool
	// AllowHeaders contains user-defined headers to be forwarded.
	AllowHeaders utils.JSONStrings `json:"allowHeaders"`
	// BuiltInAllowHeaders contains framework-default headers (Tracing, Auth, etc.).
	BuiltInAllowHeaders utils.JSONStrings `json:"builtInAllowHeaders"`
}

var defaultRest2RpcConfig = rest2RpcConfig{
	BuiltInAllowHeaders: utils.JSONStrings{
		"x-request-region",
		"x-request-az",
		"x-request-id",
		"x-request-instance",
		"Traceparent", // W3C Trace Context
		fasthttp.HeaderXForwardedFor,
		fasthttp.HeaderAuthorization,
	},
}

// complete merges built-in defaults with user-provided configuration.
func (r rest2RpcConfig) complete() rest2RpcConfig {
	r.AllowHeaders = r.BuiltInAllowHeaders.Merge(r.AllowHeaders)
	return r
}

func init() {
	// Register the interceptor specifically for gRPC client protocols.
	client.AddInterceptor(Rest2RpcInterceptorName, NewRest2Rpc, grpc.Protocol)
}

// NewRest2Rpc initializes the interceptor and starts a configuration watcher.
func NewRest2Rpc() (client.ClientInterceptor, error) {
	rest2Rpc := Rest2Rpc{
		cfg: defaultRest2RpcConfig,
	}
	// Fetch configuration from the provider and attach a reload listener.
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientRest2RpcContextPrefix,
		&rest2Rpc.cfg,
		config.WithWatch(rest2Rpc.watch)); err != nil {
		logger.Error("get interceptor config fail", "interceptor", Rest2RpcInterceptorName, "err", err)
		return nil, err
	}
	rest2Rpc.cfg = rest2Rpc.cfg.complete()
	return &rest2Rpc, nil
}

// Name returns the interceptor's registration name.
func (*Rest2Rpc) Name() string {
	return Rest2RpcInterceptorName
}

// Interceptor provides the translation logic from HTTP headers to gRPC metadata.
func (r *Rest2Rpc) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) error {
		// Attempt to type-assert the context to an HTTP context.
		rtx, ok := ctx.(rest.RestContext)
		if !ok {
			return invoker(ctx, method, req, reply, cc)
		}

		// Initialize gRPC Metadata.
		var md metadata.MD
		if existMd, ok := metadata.FromOutgoingContext(ctx); ok {
			md = existMd.Copy()
		} else {
			md = make(metadata.MD)
		}
		r.cm.RLock()
		defer r.cm.RUnlock()

		// Map allowed HTTP headers/params to gRPC metadata.
		for _, k := range r.cfg.AllowHeaders {
			key := strings.ToLower(k)
			// First, check standard headers.
			if v := rtx.Header(key); v != "" {
				md.Set(key, v)
			}
			// Then, check user-defined context parameters (overrides header if present).
			if v := rtx.UserValue(key); v != nil {
				md.Set(key, cast.ToString(v))
			}
		}

		// Attach outgoing metadata to the incoming context so cancel/deadline
		// still propagate to the gRPC transport. Using rtx.Context() here would
		// drop cancel when it is context.Background() (REST default).
		return invoker(metadata.NewOutgoingContext(ctx, md),
			method, req, reply, cc)
	}
}

// watch handles dynamic configuration updates (e.g., adding a new allowed header at runtime).
func (r *Rest2Rpc) watch(event *config.Event) {
	conf := defaultRest2RpcConfig
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientRest2RpcContextPrefix, &conf); err == nil {
		r.cm.Lock()
		r.cfg = conf.complete()
		r.cm.Unlock()
	}
}
