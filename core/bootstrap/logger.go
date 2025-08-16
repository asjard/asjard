package bootstrap

import (
	"log/slog"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
)

// Logger watch log configuration change after server start.
type Logger struct{}

func init() {
	AddBootstrap(&Logger{})
}

// Start watch log cofiguration change.
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
