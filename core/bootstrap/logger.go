package bootstrap

import (
	"github.com/asjard/asjard/core/config"
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
	config.AddPatternListener("logger.*", func(*config.Event) {
		l.update()
	})
	return nil
}

func (l Logger) update() {
	logger.SetLogger(logger.NewDefaultLogger(&logger.LoggerConfig{
		FileName:   config.GetString("logger.filePath", "/dev/stdout"),
		MaxSize:    config.GetInt("logger.maxSize", 100),
		MaxAge:     config.GetInt("logger.maxAge", 0),
		MaxBackups: config.GetInt("logger.maxBackups", 10),
		Compress:   config.GetBool("logger.compress", true),
		Level:      config.GetString("logger.level", logger.DEBUG.String()),
		Format:     config.GetString("logger.format", logger.Json.String()),
	}))
}

func (l Logger) Shutdown() {}
