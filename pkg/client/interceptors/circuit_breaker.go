package interceptors

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"google.golang.org/grpc/codes"
)

const (
	// DefaultCommandConfigName 默认配置名称
	DefaultCommandConfigName      = "default"
	CircuitBreakerInterceptorName = "circuitBreaker"

	// 配置前缀
	ConfigPrefix = "asjard.interceptors.client.circuitBreaker"
)

// CircuitBreaker 断路器
// 依赖loadbalance注入x-request-dest
type CircuitBreaker struct {
	commandConfig map[string]hystrix.CommandConfig
	cm            sync.RWMutex
	cache         sync.Map
}

// 熔断配置
type CircuitBreakerConfig struct {
	hystrix.CommandConfig
	Methods []CircuitBreakerMethodConfig
}

// 熔断方法配置
type CircuitBreakerMethodConfig struct {
	Name string `json:"name"`
	hystrix.CommandConfig
}

var (
	defaultConfig = hystrix.CommandConfig{
		Timeout:                hystrix.DefaultTimeout,
		MaxConcurrentRequests:  1000,
		RequestVolumeThreshold: hystrix.DefaultVolumeThreshold,
		SleepWindow:            hystrix.DefaultSleepWindow,
		ErrorPercentThreshold:  hystrix.DefaultErrorPercentThreshold,
	}
)

// 日志
type CircuitBreakerLogger struct{}

func (CircuitBreakerLogger) Printf(format string, items ...interface{}) {
	logger.Error(fmt.Sprintf(format, items...))
}

func init() {
	client.AddInterceptor(CircuitBreakerInterceptorName, NewCircuitBreaker)
	hystrix.SetLogger(&CircuitBreakerLogger{})
}

// NewCircuitBreaker 拦截器初始化
func NewCircuitBreaker() (client.ClientInterceptor, error) {
	circuitBreaker := &CircuitBreaker{}
	if err := circuitBreaker.loadAndWatch(); err != nil {
		return nil, err
	}
	return circuitBreaker, nil
}

// Name 拦截器名称
func (ccb *CircuitBreaker) Name() string {
	return CircuitBreakerInterceptorName
}

// Interceptor 拦截器实现
func (ccb *CircuitBreaker) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) error {
		return ccb.do(ctx, ccb.match(cc.Protocol(), cc.ServiceName(), method), method, req, reply, cc, invoker)
	}
}

func (ccb *CircuitBreaker) do(ctx context.Context, commandConfigName, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) (invokeErr error) {
	if err := hystrix.DoC(ctx, commandConfigName, func(ctx context.Context) error {
		if invokeErr = invoker(ctx, method, req, reply, cc); invokeErr != nil {
			// 只熔断5xx的错误
			es := status.FromError(invokeErr)
			if es.Status%100 != 5 {
				return nil
			}
			return invokeErr
		}
		return nil
	}, nil); err != nil {
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

func (ccb *CircuitBreaker) match(protocol, service, method string) string {
	fullName := protocol + "://" + service + method
	if name, ok := ccb.cache.Load(fullName); ok {
		logger.Debug("circuit breaker matched cache", "fullname", fullName, "command", name)
		return name.(string)
	}

	// 依次按照如下优先级查询
	// protocol://service/method
	// protocol://service
	// protocol:///method
	// protocol
	// //service/method
	// ///method
	// //service
	priorities := []string{
		fullName,
		protocol + "://" + service,
		protocol + "://" + method,
		protocol,
		"//" + service + method,
		"//" + method,
		"//" + service,
	}
	ccb.cm.RLock()
	defer ccb.cm.RUnlock()
	for _, name := range priorities {
		// logger.Debug("circuit breaker match priority", "fullname", fullName, "priority", name)
		if _, ok := ccb.commandConfig[name]; ok {
			logger.Debug("circuit breaker matched command", "fullname", fullName, "command", name)
			ccb.cache.Store(fullName, name)
			return name
		}
	}
	logger.Debug("circuit breaker matched default", "fullname", fullName)
	return DefaultCommandConfigName
}

func (ccb *CircuitBreaker) loadAndWatch() error {
	if err := ccb.load(); err != nil {
		return err
	}
	config.AddPrefixListener(ConfigPrefix, ccb.watch)
	return nil
}

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
		if err := config.GetWithUnmarshal(fmt.Sprintf("%s.methods[%d]", ConfigPrefix, idx), &mc); err != nil {
			return err
		}
		confMap[method.Name] = mc.CommandConfig
	}
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
