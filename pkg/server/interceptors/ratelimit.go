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
	// RateLimiterInterceptorName 限速器名称
	RateLimiterInterceptorName = "ratelimiter"
	// AllMethods 所有方法
	AllMethods = "*"
)

// RateLimiter 服务端限速拦截器
// 无需实现redis版本的限速器
// 限速的目的是为了保护服务,以免服务过载
// 如需精确限制访问速度，请参考quota拦截器
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	lm       sync.RWMutex
}

// RateLimiterConfig 限速器配置
type RateLimiterConfig struct {
	// 每秒生成多少个Token
	Limit int `json:"limit"`
	// 桶大小
	Burst   int                        `json:"burst"`
	Methods []*MethodRateLimiterConfig `json:"methods"`
}

type MethodRateLimiterConfig struct {
	Name string `json:"name"`
	*RateLimiterConfig
}

var (
	defaultRateLimiteConfig = RateLimiterConfig{
		Limit: -1,
		Burst: -1,
	}
)

func init() {
	server.AddInterceptor(RateLimiterInterceptorName, NewRateLimiterInterceptor)
}

// NewRateLimiterInterceptor 限速器初始化
func NewRateLimiterInterceptor() (server.ServerInterceptor, error) {
	ratelimiter := &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
	if err := ratelimiter.loadAndWatch(); err != nil {
		return nil, err
	}
	return ratelimiter, nil
}

// Interceptor 拦截器实现
func (rl *RateLimiter) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		limiter := rl.getLimiter(info.Protocol, info.FullMethod)
		logger.Debug("start ratelimiter interceptor", "protocol", info.Protocol, "limit", limiter.Limit(), "burst", limiter.Burst())
		if limiter.Allow() {
			return handler(ctx, req)
		}
		return nil, status.TooManyRequest()
	}
}

// Name 拦截器名称
func (*RateLimiter) Name() string {
	return RateLimiterInterceptorName
}

func (rl *RateLimiter) loadAndWatch() error {
	if err := rl.load(); err != nil {
		return err
	}
	config.AddPrefixListener("asjard.interceptors.server.rateLimiter", rl.watch)
	return nil
}

func (rl *RateLimiter) load() error {
	conf := defaultRateLimiteConfig
	if err := config.GetWithUnmarshal("asjard.interceptors.server.rateLimiter", &conf); err != nil {
		return err
	}
	rl.setLimiters(&conf)
	rl.cleanLimiters(&conf)
	return nil
}

func (rl *RateLimiter) watch(event *config.Event) {
	if err := rl.load(); err != nil {
		logger.Error("ratelimiter load config fail", "err", err)
	}
}

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

func (rl *RateLimiter) setLimiters(conf *RateLimiterConfig) {
	rl.setLimiter(AllMethods, rate.Limit(conf.Limit), conf.Burst)
	for _, method := range conf.Methods {
		rl.setLimiter(method.Name, rate.Limit(method.Limit), method.Burst)
	}
}

func (rl *RateLimiter) setLimiter(method string, limit rate.Limit, burst int) {
	if limit < 0 {
		limit = rate.Inf
	}
	if burst < 0 {
		burst = int(limit)
	}
	rl.lm.Lock()
	if limiter, ok := rl.limiters[method]; ok {
		limiter.SetLimit(limit)
		limiter.SetBurst(burst)
	} else {
		rl.limiters[method] = rate.NewLimiter(limit, burst)
	}
	rl.lm.Unlock()
}

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
