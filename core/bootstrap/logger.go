package bootstrap

import (
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"

	// 引入zap日志
	// _ "github.com/asjard/asjard/pkg/logger"
	"github.com/spf13/cast"
)

// Logger .
type Logger struct{}

func init() {
	AddBootstrap(&Logger{})
}

// Bootstrap .
func (l Logger) Bootstrap() error {
	level := config.GetString("logger.level", logger.DEBUG.String(),
		config.WithToUpper(), config.WithWatch(func(event *config.Event) {
			logger.SetLevel(cast.ToString(event.Value.Value))
		}))
	logger.SetLevel(level)
	return nil
}

// Shutdown .
func (l Logger) Shutdown() {}
