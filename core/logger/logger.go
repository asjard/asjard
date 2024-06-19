package logger

import (
	"context"
	"log/slog"
	"runtime"
	"strconv"

	"gopkg.in/natefinch/lumberjack.v2"
	// aruntime "github.com/asjard/asjard/core/runtime"
)

// Logger 日志
type Logger interface {
	// info级别日志
	Info(format string, kv ...any)
	// debug级别日志
	Debug(format string, kv ...any)
	// warn级别日志
	Warn(format string, kv ...any)
	// error级别日志
	Error(format string, kv ...any)
}

// L 日志组件
var L Logger

// NewLoggerHandler 初始化logger handler的方法
type NewLoggerHandler func() slog.Handler

func init() {
	SetLoggerHandler(defaultLoggerHandler)
}

// SetLoggerHandler 设置logger handler
func SetLoggerHandler(newFunc NewLoggerHandler) {
	L = NewLogger(newFunc)
}

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
	Format:     Json.String(),
}

// NewLogger .
func NewLogger(newFunc NewLoggerHandler) *defaultLogger {
	return &defaultLogger{
		slogger: slog.New(newFunc()),
	}
}

func defaultLoggerHandler() slog.Handler {
	return GetSlogHandler(defaultLoggerConfig)
}

// GetSlogHandler .
func GetSlogHandler(cfg *LoggerConfig) slog.Handler {
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

// Info .
func Info(format string, kv ...any) {
	L.Info(format, kv...)
}

// Debug .
func Debug(format string, kv ...any) {
	L.Debug(format, kv...)
}

// Warn .
func Warn(format string, kv ...any) {
	L.Warn(format, kv...)
}

// Error .
func Error(format string, kv ...any) {
	L.Error(format, kv...)
}
