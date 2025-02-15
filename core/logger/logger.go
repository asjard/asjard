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
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 默认日志
type Logger struct {
	slogger *slog.Logger
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

// L 日志组件
var (
	L atomic.Pointer[Logger]
)

func init() {
	L.Store(&Logger{
		slogger: slog.New(NewSlogHandler(&DefaultConfig)),
	})
}

// NewLoggerHandler 初始化logger handler的方法
type NewLoggerHandler func() slog.Handler

// SetLoggerHandler 设置logger handler
func SetLoggerHandler(newFunc NewLoggerHandler) {
	L.Store(&Logger{
		slogger: slog.New(newFunc()),
	})
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

func (dl Logger) Info(msg string, kvs ...any) {
	dl.log(slog.LevelInfo, msg, kvs...)
}

func (dl Logger) Debug(msg string, kvs ...any) {
	dl.log(slog.LevelDebug, msg, kvs...)
}

func (dl Logger) Warn(msg string, kvs ...any) {
	dl.log(slog.LevelWarn, msg, kvs...)
}

func (dl Logger) Error(msg string, kvs ...any) {
	dl.log(slog.LevelError, msg, kvs...)
}

func (dl Logger) log(level slog.Level, msg string, args ...any) {
	_, f, l, ok := runtime.Caller(3)
	if !ok {
		f = "???"
		l = 0
	} else {
		if fl := strings.Split(f, string(filepath.Separator)); len(fl) >= 3 {
			f = filepath.Join(fl[len(fl)-3:]...)
		}
	}
	dl.slogger.Log(context.Background(),
		level,
		msg,
		append(args,
			[]any{
				"app", constant.APP.Load(),
				"service", constant.ServiceName.Load(),
				"source", f + ":" + strconv.Itoa(l),
			}...)...)
}

func Info(msg string, kvs ...any) {
	L.Load().Info(msg, kvs...)
}

func Debug(msg string, kvs ...any) {
	L.Load().Debug(msg, kvs...)
}

func Warn(msg string, kvs ...any) {
	L.Load().Warn(msg, kvs...)
}

func Error(msg string, kvs ...any) {
	L.Load().Error(msg, kvs...)
}
