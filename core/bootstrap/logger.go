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
	logger.SetLoggerHandler(l.newLoggerHandler)
}

func (l Logger) newLoggerHandler() slog.Handler {
	var loggerConfig logger.LoggerConfig
	if err := config.GetWithUnmarshal(constant.ConfigLoggerPrefix, &loggerConfig); err != nil {
		logger.Error("get with unmarshal asjard.logger fail", "err", err)
	}
	return logger.GetSlogHandler(&loggerConfig)
}

func (l Logger) Shutdown() {}
