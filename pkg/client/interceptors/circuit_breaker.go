package interceptors

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"google.golang.org/grpc/codes"
)

const (
	// DefaultCommandConfigName is the fallback Hystrix command key.
	DefaultCommandConfigName = "default"
	// CircuitBreakerInterceptorName is the unique identifier for this interceptor.
	CircuitBreakerInterceptorName = "circuitBreaker"

	// ConfigPrefix is the configuration path in the framework settings.
	ConfigPrefix = "asjard.interceptors.client.circuitBreaker"
)

// CircuitBreaker manages fault tolerance for outgoing client calls.
// It uses a command map to store various Hystrix configurations and a cache
// for fast method-to-command resolution.
type CircuitBreaker struct {
	commandConfig map[string]hystrix.CommandConfig
	cm            sync.RWMutex
	cache         sync.Map // Caches matching results for high-performance lookups
}

// CircuitBreakerConfig represents the global and method-specific configuration.
type CircuitBreakerConfig struct {
	hystrix.CommandConfig
	Methods []CircuitBreakerMethodConfig
}

// CircuitBreakerMethodConfig defines Hystrix settings for a specific method/service.
type CircuitBreakerMethodConfig struct {
	Name string `json:"name"` // Key used for matching (e.g., "grpc://UserService/GetUser")
	hystrix.CommandConfig
}

var (
	// baseline settings for the circuit breaker.
	defaultConfig = hystrix.CommandConfig{
		Timeout:                hystrix.DefaultTimeout,
		MaxConcurrentRequests:  10_0000,
		RequestVolumeThreshold: hystrix.DefaultVolumeThreshold,
		SleepWindow:            hystrix.DefaultSleepWindow,
		ErrorPercentThreshold:  hystrix.DefaultErrorPercentThreshold,
	}
	// builderPool reduces allocations during string concatenation for keys.
	builderPool = sync.Pool{
		New: func() any {
			var b strings.Builder
			b.Grow(128)
			return &b
		},
	}
	// prioritiesPool reuse slices for priority matching logic.
	prioritiesPool = sync.Pool{
		New: func() any {
			return make([]string, 0, 7)
		},
	}
)

// CircuitBreakerLogger redirects Hystrix internal logs to the framework logger.
type CircuitBreakerLogger struct{}

func (CircuitBreakerLogger) Printf(format string, items ...interface{}) {
	logger.Error(fmt.Sprintf(format, items...))
}

func init() {
	// Register the interceptor specifically for the ALL protocol.
	client.AddInterceptor(CircuitBreakerInterceptorName, NewCircuitBreaker)
	hystrix.SetLogger(&CircuitBreakerLogger{})
}

// NewCircuitBreaker initializes the interceptor and starts watching for config changes.
func NewCircuitBreaker() (client.ClientInterceptor, error) {
	circuitBreaker := &CircuitBreaker{}
	if err := circuitBreaker.loadAndWatch(); err != nil {
		return nil, err
	}
	return circuitBreaker, nil
}

// Name returns the interceptor's registration name.
func (ccb *CircuitBreaker) Name() string {
	return CircuitBreakerInterceptorName
}

// Interceptor returns the actual middleware function for client requests.
func (ccb *CircuitBreaker) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) error {
		// Match the call to a specific Hystrix command based on protocol/service/method.
		return ccb.do(ctx, ccb.match(cc.Protocol(), cc.ServiceName(), method), method, req, reply, cc, invoker)
	}
}

// do executes the request within a Hystrix context.
func (ccb *CircuitBreaker) do(ctx context.Context, commandConfigName, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) (invokeErr error) {
	if err := hystrix.DoC(ctx, commandConfigName, func(ctx context.Context) error {
		if invokeErr = invoker(ctx, method, req, reply, cc); invokeErr != nil {
			// Only count 5xx (Server Errors) towards the circuit breaker failure rate.
			// Client errors (4xx) usually don't indicate an unhealthy downstream service.
			es := status.FromError(invokeErr)
			if es.Status%100 != 5 {
				return nil
			}
			return invokeErr
		}
		return nil
	}, nil); err != nil {
		// Handle Hystrix-specific errors like "Circuit Open" or "Max Concurrency Reached".
		if cerr, ok := err.(hystrix.CircuitError); ok {
			if errors.Is(cerr, hystrix.ErrMaxConcurrency) {
				return status.Error(codes.ResourceExhausted, cerr.Error())
			}
			return status.Error(codes.DataLoss, err.Error())
		}
		return err
	}
	return invokeErr
}

// match identifies which Hystrix command configuration should be applied to the request.
// It follows a specific priority order from most specific to most general.
func (ccb *CircuitBreaker) match(protocol, service, method string) string {
	fullName := ccb.buildKey(protocol, "//", service, "/", method)
	if name, ok := ccb.cache.Load(fullName); ok {
		return name.(string)
	}

	// Priority Matching Order:
	// 1. protocol://service/method
	// 2. protocol://service
	// 3. protocol:///method
	// 4. protocol
	// 5. //service/method
	// 6. ///method
	// 7. //service
	priorities := prioritiesPool.Get().([]string)
	priorities = priorities[:0]
	priorities = append(priorities,
		fullName,
		ccb.buildKey(protocol, "//", service),
		ccb.buildKey(protocol, "//", method),
		ccb.buildKey(protocol),
		ccb.buildKey("//", service, "/", method),
		ccb.buildKey("///", method),
		ccb.buildKey("//", service),
	)
	defer prioritiesPool.Put(priorities)

	ccb.cm.RLock()
	defer ccb.cm.RUnlock()
	for _, name := range priorities {
		if _, ok := ccb.commandConfig[name]; ok {
			ccb.cache.Store(fullName, name)
			return name
		}
	}
	return DefaultCommandConfigName
}

// buildKey efficiently joins string parts using a pool.
func (ccb *CircuitBreaker) buildKey(parts ...string) string {
	b := builderPool.Get().(*strings.Builder)
	b.Reset()
	for _, p := range parts {
		b.WriteString(p)
	}
	s := b.String()
	builderPool.Put(b)
	return s
}

// loadAndWatch initializes the config and attaches a prefix listener for dynamic updates.
func (ccb *CircuitBreaker) loadAndWatch() error {
	if err := ccb.load(); err != nil {
		return err
	}
	config.AddPrefixListener(ConfigPrefix, ccb.watch)
	return nil
}

// load fetches current configuration and registers it with the Hystrix engine.
func (ccb *CircuitBreaker) load() error {
	conf := CircuitBreakerConfig{
		CommandConfig: defaultConfig,
	}
	if err := config.GetWithUnmarshal(ConfigPrefix, &conf); err != nil {
		return err
	}

	confMap := make(map[string]hystrix.CommandConfig)
	confMap[DefaultCommandConfigName] = conf.CommandConfig
	for idx, method := range conf.Methods {
		mc := CircuitBreakerMethodConfig{
			Name:          method.Name,
			CommandConfig: conf.CommandConfig,
		}
		// Merge global settings with method-specific overrides.
		if err := config.GetWithUnmarshal(fmt.Sprintf("%s.methods[%d]", ConfigPrefix, idx), &mc); err != nil {
			return err
		}
		confMap[method.Name] = mc.CommandConfig
	}

	// Reset Hystrix states and local caches upon config reload.
	hystrix.Flush()
	ccb.cache.Clear()

	ccb.cm.Lock()
	ccb.commandConfig = confMap
	ccb.cm.Unlock()

	hystrix.Configure(confMap)
	return nil
}

func (ccb *CircuitBreaker) watch(_ *config.Event) {
	if err := ccb.load(); err != nil {
		logger.Error("load circuit breaker config fail", "err", err)
	}
}
