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

// Logger is the main structure for logging operations.
// It supports context-aware logging and caller information tracking.
type Logger struct {
	ctx        context.Context
	callerSkip int          // Number of stack frames to skip to reach the actual caller
	slogger    *slog.Logger // The underlying structured logger
}

// Config defines the parameters for log file management and formatting.
type Config struct {
	FileName   string `json:"filepath"`   // Path to the log file
	MaxSize    int    `json:"maxSize"`    // Maximum size in megabytes before rotation
	MaxAge     int    `json:"maxAge"`     // Maximum days to retain old log files
	MaxBackups int    `json:"maxBackups"` // Maximum number of old log files to keep
	Compress   bool   `json:"compress"`   // Whether to compress rotated files
	Level      string `json:"level"`      // Logging threshold (DEBUG, INFO, etc.)
	Format     string `json:"format"`     // Output format (Text or Json)
}

// NewLoggerHandler is a factory type for creating custom slog handlers.
type NewLoggerHandler func() slog.Handler

// DefaultConfig provides sensible defaults (stdout, JSON format, INFO level).
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
	// defaultLogger is stored atomically to allow safe runtime updates to the logging configuration.
	defaultLogger atomic.Pointer[Logger]
)

func init() {
	// Initialize the default logger on package load.
	defaultLogger.Store(DefaultLogger(slog.New(NewSlogHandler(&DefaultConfig))))
}

// L returns a logger instance bound to the provided context.
// This is the primary entry point for context-aware logging.
func L(ctx context.Context) *Logger {
	return defaultLogger.Load().clone().withContext(ctx)
}

// DefaultLogger creates a basic Logger wrapper around a slog instance.
func DefaultLogger(slogger *slog.Logger) *Logger {
	return &Logger{
		ctx:        context.TODO(),
		callerSkip: 3, // Default skip level to find the user's code calling the logger
		slogger:    slogger,
	}
}

// SetLoggerHandler allows users to override the global logger with a custom handler.
func SetLoggerHandler(newFunc NewLoggerHandler) {
	defaultLogger.Store(DefaultLogger(slog.New(newFunc())))
}

// NewSlogHandler initializes a slog handler with rotation support via lumberjack.
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

// Standard logging methods: Info, Debug, Warn, Error
func (l Logger) Info(msg string, kvs ...any)  { l.log(slog.LevelInfo, msg, kvs...) }
func (l Logger) Debug(msg string, kvs ...any) { l.log(slog.LevelDebug, msg, kvs...) }
func (l Logger) Warn(msg string, kvs ...any)  { l.log(slog.LevelWarn, msg, kvs...) }
func (l Logger) Error(msg string, kvs ...any) { l.log(slog.LevelError, msg, kvs...) }

// withContext attaches a context to the logger for trace ID extraction.
func (l *Logger) withContext(ctx context.Context) *Logger {
	l.ctx = ctx
	return l
}

// WithCallerSkip adjusts the stack frame depth (useful when wrapping this logger).
func (l *Logger) WithCallerSkip(skip int) *Logger {
	l.callerSkip = skip
	return l
}

// log is the internal core method that gathers all metadata and writes the log.
func (l Logger) log(level slog.Level, msg string, args ...any) {
	// 1. Identify the source code location (file and line)
	_, f, ln, ok := runtime.Caller(l.callerSkip)
	if !ok {
		f = "???"
		ln = 0
	} else {
		// Trim the path to show only the last 3 directories for readability
		if fl := strings.Split(f, string(filepath.Separator)); len(fl) >= 3 {
			f = filepath.Join(fl[len(fl)-3:]...)
		}
	}

	// 2. Inject Framework Metadata (Environment, Region, App Name)
	args = append(args, []any{
		"app", constant.APP.Load(),
		"region", constant.Region.Load(),
		"az", constant.AZ.Load(),
		"env", constant.Env.Load(),
		"service", constant.ServiceName.Load(),
		"source", f + ":" + strconv.Itoa(ln),
	}...)

	// 3. Inject Distributed Tracing Information (TraceID and SpanID)
	traceCtx := trace.SpanContextFromContext(l.ctx)
	if traceCtx.TraceID().IsValid() {
		args = append(args, []any{"trace", traceCtx.TraceID().String()}...)
	}
	if traceCtx.SpanID().IsValid() {
		args = append(args, []any{"span", traceCtx.SpanID().String()}...)
	}

	// 4. Final log output
	l.slogger.Log(l.ctx, level, msg, args...)
}

// L is a shorthand for withContext on an existing Logger instance.
func (l *Logger) L(ctx context.Context) *Logger {
	return l.clone().withContext(ctx)
}

func (l *Logger) clone() *Logger {
	c := *l
	return &c
}

// Package-level shortcut functions for the global default logger
func Info(msg string, kvs ...any)  { defaultLogger.Load().Info(msg, kvs...) }
func Debug(msg string, kvs ...any) { defaultLogger.Load().Debug(msg, kvs...) }
func Warn(msg string, kvs ...any)  { defaultLogger.Load().Warn(msg, kvs...) }
func Error(msg string, kvs ...any) { defaultLogger.Load().Error(msg, kvs...) }
