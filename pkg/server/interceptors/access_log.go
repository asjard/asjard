package interceptors

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/client/grpc"
	"github.com/asjard/asjard/pkg/protobuf/healthpb"
	"github.com/asjard/asjard/pkg/protobuf/requestpb"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/utils"
)

const (
	// AccessLogInterceptorName is the unique identifier for this interceptor.
	AccessLogInterceptorName = "accessLog"
)

// AccessLog handles the logging of every RPC/REST request passing through the server.
type AccessLog struct {
	cfg    *accessLogConfig
	m      sync.RWMutex   // Protects the configuration during hot-reloads.
	logger *logger.Logger // Localized logger instance specific to access logs.
}

// accessLogConfig defines the settings for request logging.
type accessLogConfig struct {
	Enabled bool `json:"enabled"`
	logger.Config

	// SkipMethods allows excluding specific protocols or methods from logging.
	// Format examples:
	// - "grpc" (skips all gRPC)
	// - "/health.Health/Check" (skips specific method)
	// - "rest:///favicon.ico" (skips specific protocol method)
	SkipMethods    utils.JSONStrings   `json:"skipMethods"`
	skipMethodsMap map[string]struct{} // Map for O(1) lookup performance.
}

var defaultAccessLogConfig = accessLogConfig{
	Config: logger.DefaultConfig,
	SkipMethods: utils.JSONStrings{
		grpc.Protocol,
		healthpb.Health_Check_FullMethodName,             // Skip health checks to reduce noise.
		requestpb.DefaultHandlers_Favicon_FullMethodName, // Skip favicon requests.
	},
}

func init() {
	// Register the interceptor factory with the global server manager.
	server.AddInterceptor(AccessLogInterceptorName, NewAccessLogInterceptor)
}

// NewAccessLogInterceptor initializes the access log component.
func NewAccessLogInterceptor() (server.ServerInterceptor, error) {
	accessLog := &AccessLog{}
	if err := accessLog.loadAndWatch(); err != nil {
		return nil, err
	}
	return accessLog, nil
}

// Name returns the interceptor's unique name.
func (*AccessLog) Name() string {
	return AccessLogInterceptorName
}

// Interceptor returns the actual middleware function that wraps request execution.
func (al *AccessLog) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		logger.L(ctx).Debug("start server interceptor", "interceptor", al.Name(), "full_method", info.FullMethod, "protocol", info.Protocol)
		// 1. Check if this specific request should be ignored based on skip rules.
		if al.skipped(info.Protocol, info.FullMethod) {
			return handler(ctx, req)
		}

		now := time.Now()
		var fields []any
		fields = append(fields, []any{"protocol", info.Protocol}...)
		fields = append(fields, []any{"full_method", info.FullMethod}...)

		// 2. Protocol-specific metadata extraction (e.g., HTTP headers/paths).
		switch info.Protocol {
		case rest.Protocol:
			if rc, ok := ctx.(*rest.Context); ok {
				fields = append(fields, []any{"header", rc.ReadHeaderParams()}...)
				fields = append(fields, []any{"method", string(rc.Method())}...)
				fields = append(fields, []any{"path", string(rc.Path())}...)
			}
		}

		// 3. Execute the actual business logic handler.
		resp, err = handler(ctx, req)

		// 4. Record post-execution metrics (latency, success, error details).
		fields = append(fields, []any{"cost", time.Since(now).String()}...)
		fields = append(fields, []any{"req", req}...)
		fields = append(fields, []any{"success", err == nil}...)
		fields = append(fields, []any{"err", err}...)

		// 5. Output to log. Errors use Error level; successful requests use Info level.
		if err != nil {
			al.logger.L(ctx).Error("access log", fields...)
		} else {
			al.logger.L(ctx).Info("access log", fields...)
		}
		return resp, err
	}
}

// skipped checks if logging is disabled or if the current method is in the skip list.
func (al *AccessLog) skipped(protocol, method string) bool {
	al.m.RLock()
	defer al.m.RUnlock()

	if !al.cfg.Enabled {
		return true
	}
	// Check protocol-level skip.
	if _, ok := al.cfg.skipMethodsMap[protocol]; ok {
		return true
	}
	// Check method-level skip.
	if _, ok := al.cfg.skipMethodsMap[method]; ok {
		return true
	}
	// Check specific protocol+method skip.
	if _, ok := al.cfg.skipMethodsMap[protocol+"://"+method]; ok {
		return true
	}
	return false
}

// loadAndWatch loads initial config and sets up a listener for dynamic updates.
func (al *AccessLog) loadAndWatch() error {
	if err := al.load(); err != nil {
		return err
	}
	// Watch for configuration changes in real-time.
	config.AddPatternListener("asjard.logger.accessLog.*", al.watch)
	return nil
}

// load parses configuration from the global config system.
func (al *AccessLog) load() error {
	conf := defaultAccessLogConfig
	if err := config.GetWithUnmarshal("asjard.logger.accessLog",
		&conf, config.WithChain([]string{"asjard.logger"})); err != nil {
		return err
	}

	// Convert slice to map for efficient lookups.
	conf.skipMethodsMap = make(map[string]struct{}, len(conf.SkipMethods))
	for _, skipMethod := range conf.SkipMethods {
		conf.skipMethodsMap[skipMethod] = struct{}{}
	}

	al.m.Lock()
	al.cfg = &conf
	// Re-initialize the internal logger with the new configuration.
	al.logger = logger.DefaultLogger(slog.New(logger.NewSlogHandler(&conf.Config))).WithCallerSkip(2)
	al.m.Unlock()
	return nil
}

// watch is the callback triggered by the configuration system when values change.
func (al *AccessLog) watch(_ *config.Event) {
	al.load()
}
