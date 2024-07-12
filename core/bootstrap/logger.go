package bootstrap

import (
	"log/slog"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
)

// Logger .
type Logger struct{}

func init() {
	AddBootstrap(&Logger{})
}

// Bootstrap 监听日志变化
func (l Logger) Bootstrap() error {
	l.update()
	config.AddPatternListener(constant.ConfigLoggerPrefix+".*", func(*config.Event) {
		l.update()
	})
	return nil
}

func (l Logger) update() {
	conf := logger.DefaultConfig
	config.GetWithUnmarshal(constant.ConfigLoggerPrefix, &conf)
	logger.SetLoggerHandler(func() slog.Handler {
		return logger.NewSlogHandler(conf)
	})
}

func (l Logger) Shutdown() {}
