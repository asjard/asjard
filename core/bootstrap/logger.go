package bootstrap

import (
	"log/slog"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
)

// Logger 服务启动后日志配置变化监听
type Logger struct{}

func init() {
	AddBootstrap(&Logger{})
}

// Start 监听日志变化
func (l Logger) Start() error {
	l.update()
	config.AddPrefixListener("asjard.logger", func(*config.Event) {
		l.update()
	})
	return nil
}

func (l Logger) update() {
	conf := logger.DefaultConfig
	config.GetWithUnmarshal("asjard.logger", &conf)
	logger.SetLoggerHandler(func() slog.Handler {
		return logger.NewSlogHandler(&conf)
	})
}

func (l Logger) Stop() {}
