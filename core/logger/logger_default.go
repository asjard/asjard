package logger

import (
	"context"
	"log/slog"
	"runtime"
	"strconv"

	"gopkg.in/natefinch/lumberjack.v2"
)

// defaultLogger 默认日志
type defaultLogger struct {
	slogger *slog.Logger
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	FileName   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool
	Level      string
	Format     string
}

var defaultLoggerConfig = &LoggerConfig{
	FileName:   "/dev/stdout",
	MaxSize:    100,
	MaxAge:     0,
	MaxBackups: 10,
	Compress:   true,
	Level:      DEBUG.String(),
	Format:     Text.String(),
}

// NewDefaultLogger .
func NewDefaultLogger(cfg *LoggerConfig) *defaultLogger {
	return &defaultLogger{
		slogger: slog.New(getSlogHandler(cfg)),
	}
}

func getSlogHandler(cfg *LoggerConfig) slog.Handler {
	writer := &lumberjack.Logger{
		Filename:   cfg.FileName,
		MaxSize:    cfg.MaxSize,
		MaxAge:     cfg.MaxAge,
		MaxBackups: cfg.MaxBackups,
	}
	handlerOptions := &slog.HandlerOptions{
		Level: getSlogLevel(cfg.Level),
	}
	switch cfg.Format {
	case Text.String():
		return slog.NewTextHandler(writer, handlerOptions)
	default:
		return slog.NewJSONHandler(writer, handlerOptions)
	}
}

func getSlogLevel(level string) slog.Level {
	switch level {
	case DEBUG.String():
		return slog.LevelDebug
	case INFO.String():
		return slog.LevelInfo
	case WARN.String():
		return slog.LevelWarn
	case ERROR.String():
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}

func (dl defaultLogger) log(level slog.Level, msg string, args ...any) {
	_, f, l, ok := runtime.Caller(3)
	if !ok {
		f = "???"
		l = 0
	}
	args = append(args, []any{"source", f + ":" + strconv.Itoa(l)}...)
	dl.slogger.Log(context.Background(), level, msg, args...)
}

// Info .
func (dl defaultLogger) Info(msg string, v ...any) {
	dl.log(slog.LevelInfo, msg, v...)
}

// Debug .
func (dl defaultLogger) Debug(msg string, v ...any) {
	dl.log(slog.LevelDebug, msg, v...)
}

// Warn .
func (dl defaultLogger) Warn(msg string, v ...any) {
	dl.log(slog.LevelWarn, msg, v...)
}

// Error .
func (dl defaultLogger) Error(msg string, v ...any) {
	dl.log(slog.LevelError, msg, v...)
}
