package interceptors

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/utils"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc/codes"
)

const (
	// DefaultCommandConfigName is the fallback Gobreaker command key.
	DefaultCommandConfigName = "default"
	// CircuitBreakerInterceptorName is the unique identifier for this interceptor.
	CircuitBreakerInterceptorName = "circuitBreaker"

	// ConfigPrefix is the configuration path in the framework settings.
	ConfigPrefix = "asjard.interceptors.client.circuitBreaker"
)

// GobreakerConfig 自定义结构，映射原生 gobreaker 的 Settings
type GobreakerConfig struct {
	Timeout               utils.JSONDuration `json:"timeout"`                 // 请求执行的超时时间(毫秒)，gobreaker不自带，我们手工控制
	Interval              utils.JSONDuration `json:"interval"`                // 统计周期(毫秒)，处于Closed状态时多久清空一次计数器
	SleepWindow           utils.JSONDuration `json:"sleep_window"`            // 熔断器开启后，多久进入Half-Open状态(毫秒)
	MaxConcurrentRequests uint32             `json:"max_concurrent_requests"` // 半开状态下允许放行的最大请求数
	ConsecutiveFailures   uint32             `json:"consecutive_failures"`    // 触发熔断的连续失败阈值
}

// CircuitBreaker manages fault tolerance for outgoing client calls using sony/gobreaker.
type CircuitBreaker struct {
	//存储具体实例映射，不再是单纯的配置项
	breakers map[string]*gobreaker.TwoStepCircuitBreaker
	configs  map[string]GobreakerConfig
	cm       sync.RWMutex
	cache    sync.Map
}

// CircuitBreakerConfig represents the global and method-specific configuration.
type CircuitBreakerConfig struct {
	GobreakerConfig
	Methods []CircuitBreakerMethodConfig
}

// CircuitBreakerMethodConfig defines Gobreaker settings for a specific method/service.
type CircuitBreakerMethodConfig struct {
	Name string `json:"name"` // Key used for matching (e.g., "grpc://UserService/GetUser")
	GobreakerConfig
}

var (
	// baseline settings for the circuit breaker.
	defaultConfig = GobreakerConfig{
		Timeout:               utils.JSONDuration{Duration: 30 * time.Second}, // 默认 30s 超时
		Interval:              utils.JSONDuration{Duration: 10 * time.Second}, // 默认 10s 清空一次计数
		SleepWindow:           utils.JSONDuration{Duration: 5 * time.Second},  // 默认熔断后 5s 进入半开
		MaxConcurrentRequests: 1,                                              // 半开状态默认只放行 1 个请求探路
		ConsecutiveFailures:   5,                                              // 默认连续失败 5 次熔断
	}

	builderPool = sync.Pool{
		New: func() any {
			var b strings.Builder
			b.Grow(128)
			return &b
		},
	}
	prioritiesPool = sync.Pool{
		New: func() any {
			return make([]string, 0, 7)
		},
	}
)

func init() {
	client.AddInterceptor(CircuitBreakerInterceptorName, NewCircuitBreaker)
}

// NewCircuitBreaker initializes the interceptor and starts watching for config changes.
func NewCircuitBreaker() (client.ClientInterceptor, error) {
	circuitBreaker := &CircuitBreaker{
		breakers: make(map[string]*gobreaker.TwoStepCircuitBreaker),
		configs:  make(map[string]GobreakerConfig),
	}
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
		commandName := ccb.match(cc.Protocol(), cc.ServiceName(), method)
		return ccb.do(ctx, commandName, method, req, reply, cc, invoker)
	}
}

// do executes the request within a Gobreaker context.
func (ccb *CircuitBreaker) do(ctx context.Context, commandName, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) error {
	ccb.cm.RLock()
	breaker, hasBreaker := ccb.breakers[commandName]
	cfg, hasCfg := ccb.configs[commandName]
	ccb.cm.RUnlock()

	if !hasBreaker {
		// 如果没有对应的熔断器，直接裸跑调用链
		return invoker(ctx, method, req, reply, cc)
	}

	// 熔断前置拦截判断（Allow）
	// TwoStepCircuitBreaker 的 Allow() 比 Execute() 更加适合 RPC 拦截器模式，因为它不需要包裹整个闭包
	success, err := breaker.Allow()
	if err != nil {
		logger.L(ctx).Error("circuit breaker open", "command_name", commandName, "err", err)
		// gobreaker 会在熔断时返回 gobreaker.ErrCircuitOpen
		return status.Error(codes.Unavailable, "circuit breaker is open")
	}

	subCtx, cancel := context.WithCancel(ctx)

	if hasCfg && cfg.Timeout.Duration > 0 {
		if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
			// 如果上游有时限，计算上游还剩多少时间完蛋
			remaining := time.Until(deadline)

			// 如果熔断器配置的超时时间，比上游剩下的时间还要短，说明上游宽裕，我们应该用更短的来保护系统
			if cfg.Timeout.Duration < remaining {
				subCtx, cancel = context.WithTimeout(ctx, cfg.Timeout.Duration)
			}
			// 反之，如果上游剩下的时间（比如还剩 200ms）比熔断器配置（1500ms）还要短，
			// 那就没必要瞎派生了，直接沿用原有的 subCtx/cancel（即随上游 200ms 后一起爆炸）

		} else {
			// 上游没有时限，无脑使用熔断器自身的超时配置
			subCtx, cancel = context.WithTimeout(ctx, cfg.Timeout.Duration)
		}
	} else {
		subCtx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	// 执行真正的 RPC 调用
	invokeErr := invoker(subCtx, method, req, reply, cc)

	// 根据业务状态码 精准决定是否上报失败
	if invokeErr != nil {
		es := status.FromError(invokeErr)
		// 校验状态码：如果是网络超时、5xx 服务端崩溃，或者 context 层面超时，判定为失败
		if es.Status%100 == 5 || errors.Is(invokeErr, context.DeadlineExceeded) {
			success(false) // 触发熔断计数
		} else {
			success(true) // 4xx 等业务客户端错误，依然视作当前通道健康，不计入失败率
		}

		// 如果是因为我们主动设置的超时引起的，包装错误码
		if errors.Is(invokeErr, context.DeadlineExceeded) && subCtx.Err() != nil && ctx.Err() == nil {
			logger.L(ctx).Error("client call timeout", "method", method, "timeout", cfg.Timeout.Duration)
			return status.Error(codes.DeadlineExceeded, "invoke timeout")
		}
		return invokeErr
	}

	// 调用成功
	success(true)
	return nil
}

// match identifies which Hystrix command configuration should be applied to the request.
func (ccb *CircuitBreaker) match(protocol, service, method string) string {
	fullName := ccb.buildKey(protocol, "//", service, "/", method)
	if name, ok := ccb.cache.Load(fullName); ok {
		return name.(string)
	}

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
		if _, ok := ccb.breakers[name]; ok {
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

// load fetches current configuration and registers it with the Gobreaker engine.
func (ccb *CircuitBreaker) load() error {
	conf := CircuitBreakerConfig{
		GobreakerConfig: defaultConfig,
	}
	if err := config.GetWithUnmarshal(ConfigPrefix, &conf); err != nil {
		return err
	}

	rawConfigs := make(map[string]GobreakerConfig)
	rawConfigs[DefaultCommandConfigName] = conf.GobreakerConfig

	for idx, method := range conf.Methods {
		mc := CircuitBreakerMethodConfig{
			Name:            method.Name,
			GobreakerConfig: conf.GobreakerConfig,
		}
		if err := config.GetWithUnmarshal(fmt.Sprintf("%s.methods[%d]", ConfigPrefix, idx), &mc); err != nil {
			return err
		}
		rawConfigs[method.Name] = mc.GobreakerConfig
	}

	// 动态构建、增量更新状态机实例，防止每次热加载都把线上正在处于熔断状态的数据给洗掉
	ccb.cm.Lock()
	newBreakers := make(map[string]*gobreaker.TwoStepCircuitBreaker)
	for name, itemCfg := range rawConfigs {
		// 如果之前的实例存在，直接沿用，保留其熔断计数器状态！
		if oldBreaker, exist := ccb.breakers[name]; exist {
			newBreakers[name] = oldBreaker
		} else {
			// 如果是新加的方法路由，创建全新的状态机
			targetCfg := itemCfg // 逃逸局部变量
			sb := gobreaker.Settings{
				Name:        name,
				MaxRequests: targetCfg.MaxConcurrentRequests,
				Interval:    targetCfg.Interval.Duration,
				Timeout:     targetCfg.SleepWindow.Duration,
				ReadyToTrip: func(counts gobreaker.Counts) bool {
					// 判定连续失败次数是否触线
					return counts.ConsecutiveFailures >= targetCfg.ConsecutiveFailures
				},
				OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
					logger.Warn("circuit breaker changed state", "command", name, "from", from.String(), "to", to.String())
				},
			}
			newBreakers[name] = gobreaker.NewTwoStepCircuitBreaker(sb)
		}
	}

	ccb.breakers = newBreakers
	ccb.configs = rawConfigs
	ccb.cm.Unlock()

	// 清空匹配缓存，让下一次请求重新执行优先级匹配
	ccb.cache.Clear()
	return nil
}

func (ccb *CircuitBreaker) watch(_ *config.Event) {
	if err := ccb.load(); err != nil {
		logger.Error("load circuit breaker config fail", "err", err)
	}
}
