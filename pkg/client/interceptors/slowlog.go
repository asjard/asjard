package interceptors

import (
	"context"
	"sync"
	"time"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/utils"
)

const (
	// SlowLogInterceptorName is the unique identifier for the slow log interceptor.
	SlowLogInterceptorName = "slowLog"
)

// SlowLogInterceptor monitors the execution time of outgoing requests
// and logs those that exceed a defined threshold.
type SlowLogInterceptor struct {
	cfg      SlowLogInterceptorConfig
	cfgMutex sync.RWMutex // Protects config during dynamic updates.
}

// SlowLogInterceptorConfig defines the threshold and exclusion rules for slow logging.
type SlowLogInterceptorConfig struct {
	// SlowThreshold defines the duration after which a call is considered "slow".
	SlowThreshold utils.JSONDuration `json:"slowThreshold"`
	// SkipMethods allows excluding specific high-latency expected methods (e.g., long-polling).
	SkipMethods utils.JSONStrings `json:"skipMethods"`
	// methodMap is a derived map for efficient O(1) lookup during request processing.
	methodMap map[string]bool
}

var (
	// defaultSlowLogInterceptorConfig sets a baseline threshold (3 seconds).
	defaultSlowLogInterceptorConfig = SlowLogInterceptorConfig{
		SlowThreshold: utils.JSONDuration{Duration: 3 * time.Second},
		methodMap:     map[string]bool{},
	}
)

func init() {
	// Register the interceptor with the global client interceptor manager.
	client.AddInterceptor(SlowLogInterceptorName, NewSlowLogInterceptor)
}

// NewSlowLogInterceptor initializes the interceptor and binds it to the configuration system.
func NewSlowLogInterceptor() (client.ClientInterceptor, error) {
	slowLogInterceptor := &SlowLogInterceptor{
		cfg: defaultSlowLogInterceptorConfig,
	}
	// Fetch configuration and register a watcher for real-time threshold adjustments.
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientSlowLogPrefix,
		&slowLogInterceptor.cfg,
		config.WithWatch(slowLogInterceptor.watch)); err != nil {
		return nil, err
	}
	return slowLogInterceptor, nil
}

// Name returns the interceptor's registration name.
func (*SlowLogInterceptor) Name() string {
	return SlowLogInterceptorName
}

// Interceptor provides the timing and logging logic for client calls.
func (s *SlowLogInterceptor) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) error {
		start := time.Now()

		// Use defer to ensure timing is calculated even if the request fails or panics.
		defer func() {
			duration := time.Since(start)
			// check if the call duration meets the "slow" criteria.
			if s.isSlow(duration, method) {
				// Record the slow call with full request context for debugging.
				logger.L(ctx).Warn("slowcall",
					"protocol", cc.Protocol(),
					"to", cc.ServiceName(),
					"method", method,
					"req", req,
					"duration", duration.String())
			}
		}()

		// Proceed with the actual service call.
		return invoker(ctx, method, req, reply, cc)
	}
}

// isSlow checks if the duration exceeds the threshold and ensures the method isn't skipped.
func (s *SlowLogInterceptor) isSlow(duration time.Duration, method string) bool {
	s.cfgMutex.RLock()
	defer s.cfgMutex.RUnlock()

	// Criteria: Threshold must be set (>0), duration must exceed it, and method must not be in skip list.
	return s.cfg.SlowThreshold.Duration > 0 &&
		duration > s.cfg.SlowThreshold.Duration &&
		!s.cfg.methodMap[method]
}

// watch handles dynamic configuration reloads from YAML or configuration centers (ETCD/Consul).
func (s *SlowLogInterceptor) watch(event *config.Event) {
	conf := defaultSlowLogInterceptorConfig
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientSlowLogPrefix, &conf); err == nil {
		s.cfgMutex.Lock()
		// Rebuild the skip map for the new configuration.
		for _, item := range conf.SkipMethods {
			conf.methodMap[item] = true
		}
		s.cfg = conf
		s.cfgMutex.Unlock()
	}
}
