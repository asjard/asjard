package interceptors

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// DefaultCommandConfigName 默认配置名称
	DefaultCommandConfigName = "default"
)

// CircuitBreaker 断路器
// 依赖loadbalance注入x-request-dest
type CircuitBreaker struct {
	commandConfig map[string]hystrix.CommandConfig
	cm            sync.RWMutex
}

var (
	defaultConfig = hystrix.CommandConfig{
		// Timeout:                hystrix.DefaultTimeout,
		Timeout:                3000,
		MaxConcurrentRequests:  1000,
		RequestVolumeThreshold: hystrix.DefaultVolumeThreshold,
		SleepWindow:            hystrix.DefaultSleepWindow,
		ErrorPercentThreshold:  hystrix.DefaultErrorPercentThreshold,
	}
)

func init() {
	client.AddInterceptor(NewCircuitBreaker)
}

// NewCircuitBreaker 拦截器初始化
func NewCircuitBreaker() client.ClientInterceptor {
	commandConfig := make(map[string]hystrix.CommandConfig)
	defaultCommandConfig := defaultConfig
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientCircuitBreakerPrefix, &defaultCommandConfig); err != nil {
		logger.Error("get default circuit breaker fail", "err", err)
	}
	serviceConfig := make(map[string]any)
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientCircuitBreakerServicePrefix, &serviceConfig); err != nil {
		logger.Error("get service circuit breaker fail", "err", err)
	}
	for name := range serviceConfig {
		serviceCommandConfig := defaultCommandConfig
		if err := config.GetWithUnmarshal(fmt.Sprintf(constant.ConfigInterceptorClientCircuitBreakerWithServicePrefix, name),
			&serviceCommandConfig); err != nil {
			logger.Error("get service circuit breaker fail", "service", name, "err", err)
		}
		commandConfig[name] = serviceCommandConfig
	}
	methodsConfig := make(map[string]any)
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientCircuitBreakerMethodPrefix, &methodsConfig); err != nil {
		logger.Error("get method circuit breaker fail", "err", err)
	}
	for name := range methodsConfig {
		methodCommandConfig := defaultCommandConfig
		if err := config.GetWithUnmarshal(fmt.Sprintf(constant.ConfigInterceptorClientCircuitBreakerWithMethodPrefix, name),
			&methodCommandConfig); err != nil {
			logger.Error("get method circuit breaker fail", "method", name, "err", err)
		}
		commandConfig[name] = methodCommandConfig
	}
	commandConfig[DefaultCommandConfigName] = defaultCommandConfig
	hystrix.Configure(commandConfig)
	return &CircuitBreaker{
		commandConfig: commandConfig,
	}
}

// Name 拦截器名称
func (ccb *CircuitBreaker) Name() string {
	return "circuitBreaker"
}

// Interceptor 拦截器实现
func (ccb *CircuitBreaker) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) error {
		methodName := strings.ReplaceAll(strings.ReplaceAll(strings.Trim(method, "/"), "/", "."), ".", "_")
		ccb.cm.RLock()
		defer ccb.cm.RUnlock()
		commandConfigName := DefaultCommandConfigName
		// method
		if _, ok := ccb.commandConfig[methodName]; ok {
			commandConfigName = methodName
		}
		// service
		if _, ok := ccb.commandConfig[cc.ServiceName()]; ok {
			commandConfigName = cc.ServiceName()
		}
		return ccb.do(ctx, commandConfigName, method, req, reply, cc, invoker)
	}
}

func (ccb *CircuitBreaker) do(ctx context.Context, commandConfigName, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) error {
	if err := hystrix.DoC(ctx, commandConfigName, func(ctx context.Context) error {
		return invoker(ctx, method, req, reply, cc)
	}, nil); err != nil {
		if _, ok := err.(hystrix.CircuitError); ok {
			return status.Error(codes.ResourceExhausted, err.Error())
		}
		return err
	}
	return nil
}

// func (ccb *CircuitBreaker) fallback(ctx context.Context, err error) error {
// 	return err
// }
