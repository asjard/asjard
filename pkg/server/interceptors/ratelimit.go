package interceptors

import (
	"context"
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/core/status"
	"golang.org/x/time/rate"
)

const (
	// RateLimiterInterceptorName is the unique identifier for this interceptor.
	RateLimiterInterceptorName = "ratelimiter"
	// AllMethods is the wildcard key used for global rate limiting.
	AllMethods = "*"
)

// RateLimiter provides local traffic control to protect the service from overload.
// It is designed for high performance and does not require external dependencies like Redis.
// For precise distributed quotas, the 'quota' interceptor should be used instead.
type RateLimiter struct {
	// limiters stores instances mapped by protocol, method, or wildcard.
	limiters map[string]*rate.Limiter
	lm       sync.RWMutex

	conf *RateLimiterConfig
}

// RateLimiterConfig defines the thresholds and scope for rate limiting.
type RateLimiterConfig struct {
	Enabled bool `json:"enabled"`
	// Limit is the number of tokens generated per second (RPS).
	Limit int `json:"limit"`
	// Burst is the maximum number of tokens the bucket can hold at once.
	Burst int `json:"burst"`
	// Methods allows for overriding global limits for specific endpoints.
	Methods []*MethodRateLimiterConfig `json:"methods"`
}

// MethodRateLimiterConfig specifies limits for a specific gRPC or REST endpoint.
type MethodRateLimiterConfig struct {
	Name string `json:"name"` // e.g., "grpc://pkg.Service/Method" or "*"
	*RateLimiterConfig
}

var (
	// defaultRateLimiteConfig provides "infinite" capacity by default.
	defaultRateLimiteConfig = RateLimiterConfig{
		Limit: -1,
		Burst: -1,
	}
)

func init() {
	// Register the rate limiter for all server instances.
	server.AddInterceptor(RateLimiterInterceptorName, NewRateLimiterInterceptor)
}

// NewRateLimiterInterceptor initializes the component and starts the configuration watcher.
func NewRateLimiterInterceptor() (server.ServerInterceptor, error) {
	ratelimiter := &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
	if err := ratelimiter.loadAndWatch(); err != nil {
		return nil, err
	}
	return ratelimiter, nil
}

// Interceptor returns the middleware that enforces rate limits.
func (rl *RateLimiter) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		if !rl.conf.Enabled {
			return handler(ctx, req)
		}

		// 1. Identify the most specific limiter available for this request.
		limiter := rl.getLimiter(info.Protocol, info.FullMethod)

		logger.L(ctx).Debug("start ratelimiter interceptor",
			"protocol", info.Protocol,
			"limit", limiter.Limit(),
			"burst", limiter.Burst())

		// 2. Check if the bucket has available tokens.
		if limiter.Allow() {
			return handler(ctx, req)
		}

		// 3. Drop the request and return 429 (REST) or ResourceExhausted (gRPC).
		return nil, status.TooManyRequest()
	}
}

// Name returns the interceptor's unique name.
func (*RateLimiter) Name() string {
	return RateLimiterInterceptorName
}

// loadAndWatch handles initial loading and dynamic hot-reloads of configuration.
func (rl *RateLimiter) loadAndWatch() error {
	if err := rl.load(); err != nil {
		return err
	}
	config.AddPrefixListener("asjard.interceptors.server.rateLimiter", rl.watch)
	return nil
}

// load parses the configuration and updates internal limiter instances.
func (rl *RateLimiter) load() error {
	conf := defaultRateLimiteConfig
	if err := config.GetWithUnmarshal("asjard.interceptors.server.rateLimiter", &conf); err != nil {
		return err
	}

	// Update or create limiters based on new config.
	rl.setLimiters(&conf)
	// Remove limiters that are no longer present in the configuration.
	rl.cleanLimiters(&conf)

	rl.conf = &conf
	return nil
}

func (rl *RateLimiter) watch(event *config.Event) {
	if err := rl.load(); err != nil {
		logger.Error("ratelimiter load config fail", "err", err)
	}
}

// getLimiter performs a hierarchical lookup:
// 1. protocol://method (most specific)
// 2. method
// 3. protocol
// 4. * (global default)
func (rl *RateLimiter) getLimiter(protocol, method string) *rate.Limiter {
	rl.lm.RLock()
	defer rl.lm.RUnlock()

	if limiter, ok := rl.limiters[protocol+"://"+method]; ok {
		return limiter
	}
	if limiter, ok := rl.limiters[method]; ok {
		return limiter
	}
	if limiter, ok := rl.limiters[protocol]; ok {
		return limiter
	}
	return rl.limiters[AllMethods]
}

// setLimiters initializes the global and method-specific token buckets.
func (rl *RateLimiter) setLimiters(conf *RateLimiterConfig) {
	rl.setLimiter(AllMethods, rate.Limit(conf.Limit), conf.Burst)
	for _, method := range conf.Methods {
		rl.setLimiter(method.Name, rate.Limit(method.Limit), method.Burst)
	}
}

// setLimiter creates a new bucket or updates an existing one without losing current state.
func (rl *RateLimiter) setLimiter(method string, limit rate.Limit, burst int) {
	if limit < 0 {
		limit = rate.Inf
	}
	if burst < 0 {
		burst = int(limit)
	}

	rl.lm.Lock()
	defer rl.lm.Unlock()

	if limiter, ok := rl.limiters[method]; ok {
		limiter.SetLimit(limit)
		limiter.SetBurst(burst)
	} else {
		rl.limiters[method] = rate.NewLimiter(limit, burst)
	}
}

// cleanLimiters ensures that the memory map doesn't leak limiters for removed configs.
func (rl *RateLimiter) cleanLimiters(conf *RateLimiterConfig) {
	var deleteMethods []string
	rl.lm.RLock()
	for existMethod := range rl.limiters {
		if existMethod == AllMethods {
			continue
		}
		exist := false
		for _, method := range conf.Methods {
			if existMethod == method.Name {
				exist = true
				break
			}
		}
		if !exist {
			deleteMethods = append(deleteMethods, existMethod)
		}
	}
	rl.lm.RUnlock()

	rl.lm.Lock()
	for _, method := range deleteMethods {
		delete(rl.limiters, method)
	}
	rl.lm.Unlock()
}
