package interceptors

import (
	"context"
	"sync"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/utils"
)

const (
	// ErrLogInterceptorName is the unique identifier for the error logging interceptor.
	ErrLogInterceptorName = "errLog"
)

// ErrLogInterceptor captures failed outgoing requests and logs them with context.
type ErrLogInterceptor struct {
	cfg      ErrLogInterceptorConfig
	cfgMutex sync.RWMutex // Protects the configuration during dynamic updates.
}

// ErrLogInterceptorConfig defines the behavior of the error logger.
type ErrLogInterceptorConfig struct {
	// Enabled determines if the interceptor should perform logging.
	Enabled bool `json:"enabled"`
	// SkipMethods allows defining specific methods to ignore (e.g., health checks).
	SkipMethods utils.JSONStrings `json:"skipMethods"`
	// methodMap is a hash map derived from SkipMethods for O(1) lookups.
	methodMap map[string]bool
}

var (
	// defaultErrLogInterceptorConfig provides the fallback state (disabled by default).
	defaultErrLogInterceptorConfig = ErrLogInterceptorConfig{
		Enabled:   false,
		methodMap: map[string]bool{},
	}
)

func init() {
	// Register the interceptor with the global client manager.
	client.AddInterceptor(ErrLogInterceptorName, NewErrLogInterceptor)
}

// NewErrLogInterceptor initializes the interceptor and sets up the configuration watcher.
func NewErrLogInterceptor() (client.ClientInterceptor, error) {
	errLog := &ErrLogInterceptor{
		cfg: defaultErrLogInterceptorConfig,
	}
	// Bind to the configuration provider and watch for changes.
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientErrLogPrefix,
		&errLog.cfg,
		config.WithWatch(errLog.watch)); err != nil {
		return nil, err
	}
	return errLog, nil
}

// Name returns the interceptor's registration name.
func (e *ErrLogInterceptor) Name() string {
	return ErrLogInterceptorName
}

// Interceptor implements the middleware logic to intercept client calls.
func (e *ErrLogInterceptor) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) error {
		// Execute the downstream call first.
		err := invoker(ctx, method, req, reply, cc)

		// If the call failed and logging is enabled for this method, record the failure.
		if err != nil && !e.skip(method) {
			// logger.L(ctx) ensures the log entry includes trace/request IDs from the context.
			logger.L(ctx).Error("response error",
				"protocol", cc.Protocol(),
				"to", cc.ServiceName(),
				"method", method,
				"req", req,
				"err", err)
		}
		return err
	}
}

// skip checks if the interceptor is disabled or if the current method is in the exclusion list.
func (e *ErrLogInterceptor) skip(method string) bool {
	e.cfgMutex.RLock()
	defer e.cfgMutex.RUnlock()
	if !e.cfg.Enabled {
		return true
	}
	return e.cfg.methodMap[method]
}

// watch is the callback triggered when the configuration provider (ETCD/YAML) updates.
func (e *ErrLogInterceptor) watch(event *config.Event) {
	conf := defaultErrLogInterceptorConfig
	// Re-parse the updated configuration.
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientErrLogPrefix, &conf); err == nil {
		// Rebuild the lookup map for skip logic.
		for _, item := range conf.SkipMethods {
			conf.methodMap[item] = true
		}
	}
	// Atomic update of the local config state.
	e.cfgMutex.Lock()
	e.cfg = conf
	e.cfgMutex.Unlock()
}
