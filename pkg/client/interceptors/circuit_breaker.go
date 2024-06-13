package interceptors

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
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

func init() {
	client.AddInterceptor(NewCircuitBreaker)
}

// NewCircuitBreaker 拦截器初始化
func NewCircuitBreaker() client.ClientInterceptor {
	commandConfig := make(map[string]hystrix.CommandConfig)
	defaultCommandConfig := hystrix.CommandConfig{
		Timeout:                config.GetInt("interceptors.client.circuitBreaker.timeout", hystrix.DefaultTimeout),
		MaxConcurrentRequests:  config.GetInt("interceptors.client.circuitBreaker.maxConcurrentRequests", hystrix.DefaultMaxConcurrent),
		RequestVolumeThreshold: config.GetInt("interceptors.client.circuitBreaker.requestVolumeThreshold", hystrix.DefaultVolumeThreshold),
		SleepWindow:            config.GetInt("interceptors.client.circuitBreaker.sleepWindow", hystrix.DefaultSleepWindow),
		ErrorPercentThreshold:  config.GetInt("interceptors.client.circuitBreaker.errorPercentThreshold", hystrix.DefaultErrorPercentThreshold),
	}
	serviceConfig := make(map[string]any)
	config.GetWithUnmarshal("interceptors.client.circuitBreaker.services", &serviceConfig)
	for name := range serviceConfig {
		commandConfig[name] = hystrix.CommandConfig{
			Timeout:                config.GetInt(fmt.Sprintf("interceptors.client.circuitBreaker.services.%s.timeout", name), defaultCommandConfig.Timeout),
			MaxConcurrentRequests:  config.GetInt(fmt.Sprintf("interceptors.client.circuitBreaker.services.%s.maxConcurrentRequests", name), defaultCommandConfig.MaxConcurrentRequests),
			RequestVolumeThreshold: config.GetInt(fmt.Sprintf("interceptors.client.circuitBreaker.services.%s.requestVolumeThreshold", name), defaultCommandConfig.RequestVolumeThreshold),
			SleepWindow:            config.GetInt(fmt.Sprintf("interceptors.client.circuitBreaker.services.%s.sleepWindow", name), defaultCommandConfig.SleepWindow),
			ErrorPercentThreshold:  config.GetInt(fmt.Sprintf("interceptors.client.circuitBreaker.services.%s.errorPercentThreshold", name), defaultCommandConfig.ErrorPercentThreshold),
		}
	}
	methodsConfig := make(map[string]any)
	config.GetWithUnmarshal("interceptors.client.circuitBreaker.methods", &methodsConfig)
	for name := range methodsConfig {
		commandConfig[name] = hystrix.CommandConfig{
			Timeout:                config.GetInt(fmt.Sprintf("interceptors.client.circuitBreaker.methods.%s.timeout", name), defaultCommandConfig.Timeout),
			MaxConcurrentRequests:  config.GetInt(fmt.Sprintf("interceptors.client.circuitBreaker.methods.%s.maxConcurrentRequests", name), defaultCommandConfig.MaxConcurrentRequests),
			RequestVolumeThreshold: config.GetInt(fmt.Sprintf("interceptors.client.circuitBreaker.methods.%s.requestVolumeThreshold", name), defaultCommandConfig.RequestVolumeThreshold),
			SleepWindow:            config.GetInt(fmt.Sprintf("interceptors.client.circuitBreaker.methods.%s.sleepWindow", name), defaultCommandConfig.SleepWindow),
			ErrorPercentThreshold:  config.GetInt(fmt.Sprintf("interceptors.client.circuitBreaker.methods.%s.errorPercentThreshold", name), defaultCommandConfig.ErrorPercentThreshold),
		}
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
			return status.Error(codes.Internal, err.Error())
		}
		return err
	}
	return nil
}

// func (ccb *CircuitBreaker) fallback(ctx context.Context, err error) error {
// 	return err
// }
