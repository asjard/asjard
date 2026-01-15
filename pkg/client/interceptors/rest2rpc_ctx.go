package interceptors

import (
	"context"
	"sync"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/pkg/client/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/utils"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/metadata"
)

const (
	// Rest2RpcContextInterceptorName is the unique identifier for this bridging interceptor.
	Rest2RpcContextInterceptorName = "rest2RpcContext"
)

// Rest2RpcContext handles the conversion of a REST-based context into a gRPC-compatible context.
// This should typically be positioned at the very beginning of the interceptor chain.
type Rest2RpcContext struct {
	cfg rest2RpcContextConfig
	cm  sync.RWMutex // Protects the configuration during dynamic updates.
}

// rest2RpcContextConfig defines which headers are allowed to pass through the bridge.
type rest2RpcContextConfig struct {
	allowAllHeaders bool
	// AllowHeaders contains user-defined headers to be forwarded.
	AllowHeaders utils.JSONStrings `json:"allowHeaders"`
	// BuiltInAllowHeaders contains framework-default headers (Tracing, Auth, etc.).
	BuiltInAllowHeaders utils.JSONStrings `json:"builtInAllowHeaders"`
}

var defaultRest2RpcContextConfig = rest2RpcContextConfig{
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
func (r rest2RpcContextConfig) complete() rest2RpcContextConfig {
	r.AllowHeaders = r.BuiltInAllowHeaders.Merge(r.AllowHeaders)
	return r
}

func init() {
	// Register the interceptor specifically for gRPC client protocols.
	client.AddInterceptor(Rest2RpcContextInterceptorName, NewRest2RpcContext, grpc.Protocol)
}

// NewRest2RpcContext initializes the interceptor and starts a configuration watcher.
func NewRest2RpcContext() (client.ClientInterceptor, error) {
	rest2RpcContext := Rest2RpcContext{
		cfg: defaultRest2RpcContextConfig,
	}
	// Fetch configuration from the provider and attach a reload listener.
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientRest2RpcContextPrefix,
		&rest2RpcContext.cfg,
		config.WithWatch(rest2RpcContext.watch)); err != nil {
		logger.Error("get interceptor config fail", "interceptor", "rest2RpcContext", "err", err)
		return nil, err
	}
	rest2RpcContext.cfg = rest2RpcContext.cfg.complete()
	return &rest2RpcContext, nil
}

// Name returns the interceptor's registration name.
func (*Rest2RpcContext) Name() string {
	return Rest2RpcContextInterceptorName
}

// Interceptor provides the translation logic from HTTP headers to gRPC metadata.
func (r *Rest2RpcContext) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) error {
		// Attempt to type-assert the context to a REST context.
		// If the source is not REST (e.g., RPC to RPC), skip this logic.
		rtx, ok := ctx.(*rest.Context)
		if !ok {
			return invoker(ctx, method, req, reply, cc)
		}

		// Initialize gRPC Metadata.
		md := make(metadata.MD)
		r.cm.RLock()
		defer r.cm.RUnlock()

		// Map allowed HTTP headers/params to gRPC metadata.
		for _, k := range r.cfg.AllowHeaders {
			// First, check standard HTTP headers.
			md[k] = rtx.GetHeaderParam(k)
			// Then, check user-defined context parameters (overrides header if present).
			v := rtx.GetUserParam(k)
			if len(v) != 0 {
				md[k] = v
			}
		}

		// Create a new gRPC outgoing context and continue the invocation chain.
		// Note: context.Background() is used here to reset the context type to standard
		// while carrying the new metadata.
		return invoker(metadata.NewOutgoingContext(context.Background(), md),
			method, req, reply, cc)
	}
}

// watch handles dynamic configuration updates (e.g., adding a new allowed header at runtime).
func (r *Rest2RpcContext) watch(event *config.Event) {
	conf := defaultRest2RpcContextConfig
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientRest2RpcContextPrefix, &conf); err == nil {
		r.cm.Lock()
		r.cfg = conf.complete()
		r.cm.Unlock()
	}
}
