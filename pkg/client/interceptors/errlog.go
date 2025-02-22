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
	ErrLogInterceptorName = "errLog"
)

type ErrLogInterceptor struct {
	cfg      ErrLogInterceptorConfig
	cfgMutex sync.RWMutex
}
type ErrLogInterceptorConfig struct {
	Enabled     bool              `json:"enabled"`
	Skipmethods utils.JSONStrings `json:"skipMethods"`
	methodMap   map[string]bool
}

var (
	defaultErrLogInterceptorConfig = ErrLogInterceptorConfig{
		Enabled:   false,
		methodMap: map[string]bool{},
	}
)

func init() {
	client.AddInterceptor(ErrLogInterceptorName, NewErrLogInterceptor)
}

func NewErrLogInterceptor() (client.ClientInterceptor, error) {
	errLog := &ErrLogInterceptor{
		cfg: defaultErrLogInterceptorConfig,
	}
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientErrLogPrefix,
		&errLog.cfg,
		config.WithWatch(errLog.watch)); err != nil {
		return nil, err
	}
	return errLog, nil
}

func (e *ErrLogInterceptor) Name() string {
	return ErrLogInterceptorName
}

func (e *ErrLogInterceptor) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) error {
		err := invoker(ctx, method, req, reply, cc)
		if err != nil && !e.skip(method) {
			logger.L().WithContext(ctx).Error("response error",
				"protocol", cc.Protocol(),
				"to", cc.ServiceName(),
				"method", method,
				"req", req,
				"err", err)
		}
		return err
	}
}

func (e *ErrLogInterceptor) skip(method string) bool {
	e.cfgMutex.RLock()
	defer e.cfgMutex.RUnlock()
	if !e.cfg.Enabled {
		return true
	}
	return e.cfg.methodMap[method]
}

func (e *ErrLogInterceptor) watch(event *config.Event) {
	conf := defaultErrLogInterceptorConfig
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientErrLogPrefix, &conf); err == nil {
		e.cfgMutex.Lock()
		for _, item := range conf.Skipmethods {
			conf.methodMap[item] = true
		}
		e.cfg = conf
		e.cfgMutex.Unlock()
	}
}
