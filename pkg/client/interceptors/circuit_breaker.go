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
	"github.com/asjard/asjard/core/status"
	"google.golang.org/grpc/codes"
)

const (
	// DefaultCommandConfigName 默认配置名称
	DefaultCommandConfigName      = "default"
	CircuitBreakerInterceptorName = "circuitBreaker"
)

// CircuitBreaker 断路器
// 依赖loadbalance注入x-request-dest
type CircuitBreaker struct {
	commandConfig map[string]hystrix.CommandConfig
	cm            sync.RWMutex
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

func init() {
	client.AddInterceptor(CircuitBreakerInterceptorName, NewCircuitBreaker)
}

// NewCircuitBreaker 拦截器初始化
func NewCircuitBreaker() (client.ClientInterceptor, error) {
	commandConfig := make(map[string]hystrix.CommandConfig)
	defaultCommandConfig := defaultConfig
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientCircuitBreakerPrefix, &defaultCommandConfig); err != nil {
		logger.Error("get default circuit breaker fail", "err", err)
		return nil, err
	}
	serviceConfig := make(map[string]any)
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientCircuitBreakerServicePrefix, &serviceConfig); err != nil {
		logger.Error("get service circuit breaker fail", "err", err)
		return nil, err
	}
	for name := range serviceConfig {
		serviceCommandConfig := defaultCommandConfig
		if err := config.GetWithUnmarshal(fmt.Sprintf(constant.ConfigInterceptorClientCircuitBreakerWithServicePrefix, name),
			&serviceCommandConfig); err != nil {
			logger.Error("get service circuit breaker fail", "service", name, "err", err)
			return nil, err
		}
		commandConfig[name] = serviceCommandConfig
	}
	methodsConfig := make(map[string]any)
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientCircuitBreakerMethodPrefix, &methodsConfig); err != nil {
		logger.Error("get method circuit breaker fail", "err", err)
		return nil, err
	}
	for name := range methodsConfig {
		methodCommandConfig := defaultCommandConfig
		if err := config.GetWithUnmarshal(fmt.Sprintf(constant.ConfigInterceptorClientCircuitBreakerWithMethodPrefix, name),
			&methodCommandConfig); err != nil {
			logger.Error("get method circuit breaker fail", "method", name, "err", err)
			return nil, err
		}
		commandConfig[name] = methodCommandConfig
	}
	commandConfig[DefaultCommandConfigName] = defaultCommandConfig
	hystrix.Configure(commandConfig)
	return &CircuitBreaker{
		commandConfig: commandConfig,
	}, nil
}

// Name 拦截器名称
func (ccb *CircuitBreaker) Name() string {
	return CircuitBreakerInterceptorName
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
		if _, ok := err.(hystrix.CircuitError); ok {
			return status.Error(codes.ResourceExhausted, err.Error())
		}
		return err
	}
	return invokeErr
}
