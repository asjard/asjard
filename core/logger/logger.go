/*
Package logger 根据配置定义日志级别，防爆，输出位置等维护日志
*/
package logger

import (
	"context"
	"log/slog"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/asjard/asjard/core/constant"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 默认日志
type Logger struct {
	ctx        context.Context
	callerSkip int
	slogger    *slog.Logger
}

// Config 日志配置
type Config struct {
	FileName   string `json:"filepath"`
	MaxSize    int    `json:"maxSize"`
	MaxAge     int    `json:"maxAge"`
	MaxBackups int    `json:"maxBackups"`
	Compress   bool   `json:"compress"`
	Level      string `json:"level"`
	Format     string `json:"format"`
}

// NewLoggerHandler 初始化logger handler的方法
type NewLoggerHandler func() slog.Handler

// DefaultConfig 日志默认配置
var DefaultConfig = Config{
	FileName:   "/dev/stdout",
	MaxSize:    100,
	MaxAge:     0,
	MaxBackups: 10,
	Compress:   true,
	Level:      INFO.String(),
	Format:     Json.String(),
}

var (
	defaultLogger atomic.Pointer[Logger]
)

func init() {
	defaultLogger.Store(DefaultLogger(slog.New(NewSlogHandler(&DefaultConfig))))
}

func L(ctx context.Context) *Logger {
	return defaultLogger.Load().clone().withContext(ctx)
}

func DefaultLogger(slogger *slog.Logger) *Logger {
	return &Logger{
		ctx:        context.TODO(),
		callerSkip: 3,
		slogger:    slogger,
	}
}

// SetLoggerHandler 设置logger handler
func SetLoggerHandler(newFunc NewLoggerHandler) {
	defaultLogger.Store(DefaultLogger(slog.New(newFunc())))
}

// NewSlogHandler slog初始化
func NewSlogHandler(cfg *Config) slog.Handler {
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

func (l Logger) Info(msg string, kvs ...any) {
	l.log(slog.LevelInfo, msg, kvs...)
}

func (l Logger) Debug(msg string, kvs ...any) {
	l.log(slog.LevelDebug, msg, kvs...)
}

func (l Logger) Warn(msg string, kvs ...any) {
	l.log(slog.LevelWarn, msg, kvs...)
}

func (l Logger) Error(msg string, kvs ...any) {
	l.log(slog.LevelError, msg, kvs...)
}

func (l *Logger) withContext(ctx context.Context) *Logger {
	l.ctx = ctx
	return l
}

func (l *Logger) WithCallerSkip(skip int) *Logger {
	l.callerSkip = skip
	return l
}

func (l Logger) log(level slog.Level, msg string, args ...any) {
	_, f, ln, ok := runtime.Caller(l.callerSkip)
	if !ok {
		f = "???"
		ln = 0
	} else {
		if fl := strings.Split(f, string(filepath.Separator)); len(fl) >= 3 {
			f = filepath.Join(fl[len(fl)-3:]...)
		}
	}
	args = append(args, []any{
		"app", constant.APP.Load(),
		"region", constant.Region.Load(),
		"az", constant.AZ.Load(),
		"env", constant.Env.Load(),
		"service", constant.ServiceName.Load(),
		"source", f + ":" + strconv.Itoa(ln),
	}...)
	traceCtx := trace.SpanContextFromContext(l.ctx)
	if traceCtx.TraceID().IsValid() {
		args = append(args, []any{
			"trace", traceCtx.TraceID().String(),
		}...)
	}
	if traceCtx.SpanID().IsValid() {
		args = append(args, []any{
			"span", traceCtx.SpanID().String(),
		}...)
	}
	l.slogger.Log(l.ctx,
		level,
		msg,
		args...)
}

func (l *Logger) L(ctx context.Context) *Logger {
	return l.clone().withContext(ctx)
}

func (l *Logger) clone() *Logger {
	c := *l
	return &c
}

func Info(msg string, kvs ...any) {
	defaultLogger.Load().Info(msg, kvs...)
}

func Debug(msg string, kvs ...any) {
	defaultLogger.Load().Debug(msg, kvs...)
}

func Warn(msg string, kvs ...any) {
	defaultLogger.Load().Warn(msg, kvs...)
}

func Error(msg string, kvs ...any) {
	defaultLogger.Load().Error(msg, kvs...)
}
