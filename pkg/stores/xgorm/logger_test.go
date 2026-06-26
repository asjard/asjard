package xgorm

import (
	"context"
	"log/slog"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type captureHandler struct {
	mu     sync.Mutex
	record slog.Record
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.record = r
	return nil
}

func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler {
	return h
}

func (h *captureHandler) WithGroup(string) slog.Handler {
	return h
}

func (h *captureHandler) attr(name string) (slog.Value, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	var value slog.Value
	var ok bool
	h.record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == name {
			value = attr.Value
			ok = true
			return false
		}
		return true
	})
	return value, ok
}

func (h *captureHandler) source() runtime.Frame {
	h.mu.Lock()
	defer h.mu.Unlock()
	frames := runtime.CallersFrames([]uintptr{h.record.PC})
	frame, _ := frames.Next()
	return frame
}

func TestTraceReportsCallerOutsideXgormLogger(t *testing.T) {
	handler := &captureHandler{}
	l := &xgormLogger{
		logLevel: gormLogger.Info,
		name:     "test",
		slogger:  logger.DefaultLogger(slog.New(handler)),
	}

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)

	l.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "SELECT 1", 1
	}, nil)

	line, ok := handler.attr("line")
	require.True(t, ok)
	require.True(t, strings.HasPrefix(line.String(), file+":"), line.String())
	require.NotContains(t, line.String(), "logger.go")

	source := handler.source()
	require.Equal(t, file, source.File)
}

func TestGormTraceReportsQueryCaller(t *testing.T) {
	handler := &captureHandler{}
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/caller.db"), &gorm.Config{
		Logger: &xgormLogger{
			logLevel: gormLogger.Info,
			name:     "test",
			slogger:  logger.DefaultLogger(slog.New(handler)),
		},
	})
	require.NoError(t, err)

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)

	require.NoError(t, db.Exec("SELECT 1").Error)

	line, ok := handler.attr("line")
	require.True(t, ok)
	require.True(t, strings.HasPrefix(line.String(), file+":"), line.String())
	require.NotContains(t, line.String(), "logger.go")
	require.NotContains(t, line.String(), "gorm.io")

	source := handler.source()
	require.Equal(t, file, source.File)
	require.NotContains(t, source.File, "logger.go")
	require.NotContains(t, source.File, "gorm.io")
}
