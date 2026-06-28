package xgorm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
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
	mu      sync.Mutex
	records []slog.Record
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = append(h.records, r.Clone())
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
	h.lastRecord().Attrs(func(attr slog.Attr) bool {
		if attr.Key == name {
			value = attr.Value
			ok = true
			return false
		}
		return true
	})
	return value, ok
}

func (h *captureHandler) recordCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.records)
}

func (h *captureHandler) lastRecord() slog.Record {
	if len(h.records) == 0 {
		return slog.Record{}
	}
	return h.records[len(h.records)-1]
}

func (h *captureHandler) lastRecordSnapshot() slog.Record {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.lastRecord()
}

func (h *captureHandler) source() runtime.Frame {
	h.mu.Lock()
	defer h.mu.Unlock()
	frames := runtime.CallersFrames([]uintptr{h.lastRecord().PC})
	frame, _ := frames.Next()
	return frame
}

func newTestLogger(handler *captureHandler) *xgormLogger {
	return &xgormLogger{
		logLevel: gormLogger.Info,
		name:     "test",
		slogger:  logger.DefaultLogger(slog.New(handler)),
	}
}

func TestTraceReportsCallerOutsideXgormLogger(t *testing.T) {
	handler := &captureHandler{}
	l := newTestLogger(handler)

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)

	l.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "SELECT 1", 1
	}, nil)

	source := handler.source()
	require.Equal(t, file, source.File)
	require.Contains(t, source.Function, ".TestTraceReportsCallerOutsideXgormLogger")
}

func TestGormTraceReportsQueryCaller(t *testing.T) {
	handler := &captureHandler{}
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/caller.db"), &gorm.Config{
		Logger: newTestLogger(handler),
	})
	require.NoError(t, err)

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)

	require.NoError(t, db.Exec("SELECT 1").Error)

	source := handler.source()
	require.Equal(t, file, source.File)
	require.NotContains(t, source.File, "logger.go")
	require.NotContains(t, source.File, "gorm.io")
}

func TestTraceLogsExpectedFields(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*xgormLogger)
		begin       time.Time
		err         error
		wantLevel   slog.Level
		wantMessage string
		wantRecords int
	}{
		{
			name: "error",
			setup: func(l *xgormLogger) {
				l.logLevel = gormLogger.Error
			},
			begin:       time.Now(),
			err:         errors.New("insert failed"),
			wantLevel:   slog.LevelError,
			wantMessage: "insert failed",
			wantRecords: 1,
		},
		{
			name: "record not found ignored",
			setup: func(l *xgormLogger) {
				l.logLevel = gormLogger.Error
				l.ignoreRecordNotFoundError = true
			},
			begin:       time.Now(),
			err:         gormLogger.ErrRecordNotFound,
			wantRecords: 0,
		},
		{
			name: "record not found logged when enabled",
			setup: func(l *xgormLogger) {
				l.logLevel = gormLogger.Error
				l.ignoreRecordNotFoundError = false
			},
			begin:       time.Now(),
			err:         gormLogger.ErrRecordNotFound,
			wantLevel:   slog.LevelError,
			wantMessage: gormLogger.ErrRecordNotFound.Error(),
			wantRecords: 1,
		},
		{
			name: "slow sql",
			setup: func(l *xgormLogger) {
				l.logLevel = gormLogger.Warn
				l.slowThreshold = time.Millisecond
			},
			begin:       time.Now().Add(-2 * time.Millisecond),
			wantLevel:   slog.LevelWarn,
			wantMessage: "SLOW SQL >= 1ms",
			wantRecords: 1,
		},
		{
			name: "info sql",
			setup: func(l *xgormLogger) {
				l.logLevel = gormLogger.Info
			},
			begin:       time.Now(),
			wantLevel:   slog.LevelDebug,
			wantMessage: "SELECT 1",
			wantRecords: 1,
		},
		{
			name: "silent",
			setup: func(l *xgormLogger) {
				l.logLevel = gormLogger.Silent
			},
			begin:       time.Now(),
			wantRecords: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &captureHandler{}
			l := newTestLogger(handler)
			l.ignoreRecordNotFoundError = true
			tt.setup(l)

			l.Trace(context.Background(), tt.begin, func() (string, int64) {
				return "SELECT 1", 2
			}, tt.err)

			require.Equal(t, tt.wantRecords, handler.recordCount())
			if tt.wantRecords == 0 {
				return
			}

			record := handler.lastRecordSnapshot()
			require.Equal(t, tt.wantLevel, record.Level)
			require.Equal(t, tt.wantMessage, record.Message)

			for name, want := range map[string]string{
				"cost": "",
				"db":   "test",
			} {
				value, ok := handler.attr(name)
				require.True(t, ok, "missing %s attr", name)
				if want != "" {
					require.Equal(t, want, value.String())
				}
			}
			if tt.wantMessage != "SELECT 1" {
				sql, ok := handler.attr("sql")
				require.True(t, ok)
				require.Equal(t, "SELECT 1", sql.String())
			}

			row, ok := handler.attr("row")
			require.True(t, ok)
			require.Equal(t, int64(2), row.Int64())
		})
	}
}

func TestInfoWarnErrorLogDatabaseName(t *testing.T) {
	tests := []struct {
		name      string
		log       func(*xgormLogger)
		wantLevel slog.Level
	}{
		{
			name: "info",
			log: func(l *xgormLogger) {
				l.Info(context.Background(), "hello %s", "info")
			},
			wantLevel: slog.LevelInfo,
		},
		{
			name: "warn",
			log: func(l *xgormLogger) {
				l.Warn(context.Background(), "hello %s", "warn")
			},
			wantLevel: slog.LevelWarn,
		},
		{
			name: "error",
			log: func(l *xgormLogger) {
				l.Error(context.Background(), "hello %s", "error")
			},
			wantLevel: slog.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &captureHandler{}
			l := newTestLogger(handler)

			tt.log(l)

			record := handler.lastRecordSnapshot()
			require.Equal(t, tt.wantLevel, record.Level)
			require.Equal(t, fmt.Sprintf("hello %s", tt.name), record.Message)

			db, ok := handler.attr("db")
			require.True(t, ok)
			require.Equal(t, "test", db.String())
		})
	}
}

func TestLogModeClonesLoggerConfiguration(t *testing.T) {
	handler := &captureHandler{}
	l := newTestLogger(handler)
	l.ignoreRecordNotFoundError = false
	l.slowThreshold = time.Second

	cloned, ok := l.LogMode(gormLogger.Error).(*xgormLogger)
	require.True(t, ok)
	require.NotSame(t, l, cloned)
	require.Equal(t, gormLogger.Error, cloned.logLevel)
	require.Equal(t, l.ignoreRecordNotFoundError, cloned.ignoreRecordNotFoundError)
	require.Equal(t, l.slowThreshold, cloned.slowThreshold)
	require.Equal(t, l.name, cloned.name)
	require.Same(t, l.slogger, cloned.slogger)
	require.Equal(t, gormLogger.Info, l.logLevel)
}
