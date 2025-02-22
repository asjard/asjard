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
	SlowLogInterceptorName = "slowLog"
)

type SlowLogInterceptor struct {
	cfg      SlowLogInterceptorConfig
	cfgMutex sync.RWMutex
}

type SlowLogInterceptorConfig struct {
	SlowThreshold utils.JSONDuration `json:"slowThreshold"`
	SkipMethods   utils.JSONStrings  `json:"skipMethods"`
	methodMap     map[string]bool
}

var (
	defaultSlowLogInterceptorConfig = SlowLogInterceptorConfig{
		SlowThreshold: utils.JSONDuration{Duration: 3 * time.Second},
		methodMap:     map[string]bool{},
	}
)

func init() {
	client.AddInterceptor(SlowLogInterceptorName, NewSlowLogInterceptor)
}

func NewSlowLogInterceptor() (client.ClientInterceptor, error) {
	slowLogInterceptor := &SlowLogInterceptor{
		cfg: defaultSlowLogInterceptorConfig,
	}
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientSlowLogPrefix,
		&slowLogInterceptor.cfg,
		config.WithWatch(slowLogInterceptor.watch)); err != nil {
		return nil, err
	}
	return slowLogInterceptor, nil
}

func (*SlowLogInterceptor) Name() string {
	return SlowLogInterceptorName
}

func (s *SlowLogInterceptor) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) error {
		start := time.Now()
		defer func() {
			duration := time.Since(start)
			if s.isSlow(duration, method) {
				logger.L().WithContext(ctx).Warn("slowcall",
					"protocol", cc.Protocol(),
					"to", cc.ServiceName(),
					"method", method,
					"req", req,
					"duration", duration.String())
			}
		}()
		return invoker(ctx, method, req, reply, cc)
	}
}

func (s *SlowLogInterceptor) isSlow(duration time.Duration, method string) bool {
	s.cfgMutex.RLock()
	defer s.cfgMutex.RUnlock()
	return s.cfg.SlowThreshold.Duration > 0 && duration > s.cfg.SlowThreshold.Duration && !s.cfg.methodMap[method]
}

func (s *SlowLogInterceptor) watch(event *config.Event) {
	conf := defaultSlowLogInterceptorConfig
	if err := config.GetWithUnmarshal(constant.ConfigInterceptorClientSlowLogPrefix, &conf); err == nil {
		s.cfgMutex.Lock()
		for _, item := range conf.SkipMethods {
			conf.methodMap[item] = true
		}
		s.cfg = conf
		s.cfgMutex.Unlock()
	}
}
