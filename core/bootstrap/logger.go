package bootstrap

import (
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/spf13/cast"
)

// Logger .
type Logger struct{}

func init() {
	AddBootstrap(&Logger{})
}

// Start .
func (l Logger) Start() error {
	level := config.GetString("logger.level", logger.DEBUG.String(),
		config.WithToUpper(), config.WithWatch(func(event *config.Event) {
			logger.SetLevel(cast.ToString(event.Value.Value))
		}))
	logger.SetLevel(level)
	return nil
}

// Stop .
func (l Logger) Stop() {}
