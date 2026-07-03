package logger_test

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/asjard/asjard/core/logger"
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

func (h *captureHandler) hasCaller(suffix string) bool {
	return strings.Contains(strings.Join(h.callers(), "\n"), suffix)
}

func (h *captureHandler) source() runtime.Frame {
	h.mu.Lock()
	defer h.mu.Unlock()
	frames := runtime.CallersFrames([]uintptr{h.record.PC})
	frame, _ := frames.Next()
	return frame
}

func (h *captureHandler) callers() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	var names []string
	frames := runtime.CallersFrames([]uintptr{h.record.PC})
	for {
		frame, more := frames.Next()
		names = append(names, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
		if !more {
			return names
		}
	}
}

func nextLine(t *testing.T) (string, int) {
	t.Helper()
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return file, line + 1
}

func TestLoggerReportsDirectCaller(t *testing.T) {
	handler := &captureHandler{}
	l := logger.DefaultLogger(slog.New(handler))

	wantFile, wantLine := nextLine(t)
	l.Info("direct")

	source := handler.source()
	if source.File != wantFile || source.Line != wantLine || !strings.HasSuffix(source.Function, ".TestLoggerReportsDirectCaller") {
		t.Fatalf("source = %s:%d %s, want %s:%d TestLoggerReportsDirectCaller",
			source.File, source.Line, source.Function, wantFile, wantLine)
	}
}

func TestLoggerReportsContextCaller(t *testing.T) {
	handler := &captureHandler{}
	l := logger.DefaultLogger(slog.New(handler))

	l.L(context.Background()).Info("context")

	if !handler.hasCaller(".TestLoggerReportsContextCaller") {
		t.Fatal("caller does not include test call site")
	}
}

func TestPackageLoggerReportsCaller(t *testing.T) {
	handler := &captureHandler{}
	logger.SetLoggerHandler(func() slog.Handler {
		return handler
	})

	logger.Info("package")

	if !handler.hasCaller(".TestPackageLoggerReportsCaller") {
		t.Fatalf("caller does not include package shortcut caller: %v", handler.callers())
	}
}

func TestLoggerCanSkipExternalWrapper(t *testing.T) {
	handler := &captureHandler{}
	l := logger.DefaultLogger(slog.New(handler))

	wrappedInfo(l)

	if !handler.hasCaller(".TestLoggerCanSkipExternalWrapper") {
		t.Fatal("caller does not include wrapper caller")
	}
}

func wrappedInfo(l *logger.Logger) {
	l.WithCallerSkip(1).Info("wrapped")
}

func TestLoggerCanUseSourcePC(t *testing.T) {
	handler := &captureHandler{}
	l := logger.DefaultLogger(slog.New(handler))

	pc, wantFile, wantLine, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	l.WithSourcePC(pc).Info("source pc")

	source := handler.source()
	if source.File != wantFile || source.Line != wantLine {
		t.Fatalf("source = %s:%d, want %s:%d", source.File, source.Line, wantFile, wantLine)
	}
}
