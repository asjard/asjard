package logger

import (
	"context"
	"log/slog"
	"sync/atomic"
)

// sharedDesensitizer is updated on logger config reload so all handlers pick up new rules.
var sharedDesensitizer atomic.Pointer[Desensitizer]

// UpdateDesensitizer replaces the process-wide desensitizer used by DesensitizeHandler.
// Pass a config with Disabled=true (or empty rules) to clear masking.
func UpdateDesensitizer(cfg *DesensitizeConfig) {
	sharedDesensitizer.Store(NewDesensitizer(cfg))
}

// CurrentDesensitizer returns the process-wide desensitizer (may be nil).
func CurrentDesensitizer() *Desensitizer {
	return sharedDesensitizer.Load()
}

// DesensitizeHandler wraps a slog.Handler and masks sensitive attributes
// using the process-wide desensitizer so rules can hot-reload.
type DesensitizeHandler struct {
	next slog.Handler
}

// NewDesensitizeHandler returns a handler that desensitizes attrs before
// delegating to next. It always wraps so enabling/disabling via
// UpdateDesensitizer takes effect without recreating the handler chain.
func NewDesensitizeHandler(next slog.Handler) slog.Handler {
	if next == nil {
		return nil
	}
	return &DesensitizeHandler{next: next}
}

// Enabled reports whether the next handler enables the level.
func (h *DesensitizeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

// Handle masks record attributes then forwards to the next handler.
func (h *DesensitizeHandler) Handle(ctx context.Context, r slog.Record) error {
	d := sharedDesensitizer.Load()
	if d == nil {
		return h.next.Handle(ctx, r)
	}
	nr := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	r.Attrs(func(a slog.Attr) bool {
		nr.AddAttrs(d.MaskAttr(a))
		return true
	})
	return h.next.Handle(ctx, nr)
}

// WithAttrs returns a handler with desensitized attributes attached.
func (h *DesensitizeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	d := sharedDesensitizer.Load()
	if d == nil {
		return &DesensitizeHandler{next: h.next.WithAttrs(attrs)}
	}
	masked := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		masked[i] = d.MaskAttr(a)
	}
	return &DesensitizeHandler{next: h.next.WithAttrs(masked)}
}

// WithGroup returns a handler that starts an attribute group.
func (h *DesensitizeHandler) WithGroup(name string) slog.Handler {
	return &DesensitizeHandler{next: h.next.WithGroup(name)}
}
