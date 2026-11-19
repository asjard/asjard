package logger_test

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"testing"

	"github.com/asjard/asjard/v2/core/logger"
)

type attrCaptureHandler struct {
	mu    sync.Mutex
	attrs map[string]slog.Value
}

func (h *attrCaptureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *attrCaptureHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.attrs = make(map[string]slog.Value)
	r.Attrs(func(a slog.Attr) bool {
		h.attrs[a.Key] = a.Value
		return true
	})
	return nil
}

func (h *attrCaptureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *attrCaptureHandler) WithGroup(string) slog.Handler      { return h }

func (h *attrCaptureHandler) get(key string) slog.Value {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.attrs[key]
}

func defaultDesensitize() logger.DesensitizeConfig {
	return logger.DefaultConfig.Desensitize
}

func TestDesensitizerDisabled(t *testing.T) {
	cfg := defaultDesensitize()
	cfg.Disabled = true
	d := logger.NewDesensitizer(&cfg)
	if d != nil {
		t.Fatal("expected nil desensitizer when disabled")
	}
}

func TestDesensitizerDefaultKeys(t *testing.T) {
	cfg := defaultDesensitize()
	d := logger.NewDesensitizer(&cfg)
	if d == nil {
		t.Fatal("expected desensitizer")
	}

	if got := d.MaskValue("password", "secret123"); got != "***" {
		t.Fatalf("password key: got %v", got)
	}
	if got := d.MaskValue("Authorization", "Bearer abc"); got != "***" {
		t.Fatalf("authorization key case-insensitive: got %v", got)
	}
	if got := d.MaskValue("note", "call 13800138000 please"); got != "call 13800138000 please" {
		t.Fatalf("unexpected pattern masking: got %v", got)
	}
}

func TestDesensitizerGlobalKeepPrefixSuffix(t *testing.T) {
	cfg := defaultDesensitize()
	cfg.KeepPrefix = 3
	cfg.KeepSuffix = 4
	d := logger.NewDesensitizer(&cfg)

	if got := d.MaskValue("mobile", "13800138000"); got != "138***8000" {
		t.Fatalf("global keep: got %v", got)
	}
	// length <= keepPrefix+keepSuffix => replacement only
	if got := d.MaskValue("password", "secret"); got != "***" {
		t.Fatalf("short value: got %v", got)
	}
	if got := d.MaskValue("password", "密文测试数据一二"); got != "密文测***数据一二" {
		t.Fatalf("unicode keep: got %v", got)
	}
}

func TestDesensitizerPerKeyOverride(t *testing.T) {
	d := logger.NewDesensitizer(&logger.DesensitizeConfig{
		ReplacementConfig: logger.ReplacementConfig{
			Replacement: "***",
			KeepPrefix:  1,
			KeepSuffix:  1,
		},
		Keys: []logger.DesensitizeKey{
			{Name: "pin"}, // keepPrefix/keepSuffix=0 => inherit global when global > 0
			{
				Name: "mobile",
				ReplacementConfig: logger.ReplacementConfig{
					KeepPrefix: 3,
					KeepSuffix: 4,
				},
			},
			{
				Name: "secret",
				ReplacementConfig: logger.ReplacementConfig{
					Replacement: "[HIDDEN]",
					KeepPrefix:  2,
					KeepSuffix:  2,
				},
			},
			{
				Name:    "code",
				Pattern: `CODE-\d+`,
				ReplacementConfig: logger.ReplacementConfig{
					KeepPrefix: 2,
					KeepSuffix: 2,
				},
			},
		},
	})

	if got := d.MaskValue("pin", "123456"); got != "1***6" {
		t.Fatalf("key inherit global: got %v", got)
	}
	if got := d.MaskValue("mobile", "13800138000"); got != "138***8000" {
		t.Fatalf("key keep override: got %v", got)
	}
	if got := d.MaskValue("secret", "abcdefgh"); got != "ab[HIDDEN]gh" {
		t.Fatalf("key replacement override: got %v", got)
	}
	if got := d.MaskValue("msg", "use CODE-123456 now"); got != "use CO***56 now" {
		t.Fatalf("pattern override: got %v", got)
	}
}

func TestDesensitizerUserKeysOverrideBuiltin(t *testing.T) {
	cfg := defaultDesensitize()
	cfg.Keys = []logger.DesensitizeKey{{
		Name: "password",
		ReplacementConfig: logger.ReplacementConfig{
			Replacement: "###",
			KeepPrefix:  1,
			KeepSuffix:  1,
		},
	}}
	d := logger.NewDesensitizer(&cfg)
	if got := d.MaskValue("password", "secret"); got != "s###t" {
		t.Fatalf("user key should override builtin: got %v", got)
	}
}

func TestDesensitizerNestedMapStructHeader(t *testing.T) {
	cfg := defaultDesensitize()
	d := logger.NewDesensitizer(&cfg)

	m := d.MaskValue("req", map[string]any{
		"username": "alice",
		"password": "p@ss",
		"profile": map[string]string{
			"mobile": "13900001111",
		},
	}).(map[string]any)
	if m["password"] != "***" {
		t.Fatalf("nested password: %v", m["password"])
	}
	profile := m["profile"].(map[string]string)
	if profile["mobile"] != "***" {
		t.Fatalf("nested mobile key: %v", profile["mobile"])
	}

	headers := d.MaskValue("header", map[string][]string{
		"Authorization": {"Bearer token"},
		"X-Request-Id":  {"req-1"},
	}).(map[string][]string)
	if len(headers["Authorization"]) == 0 || headers["Authorization"][0] != "***" {
		t.Fatalf("map[]string authorization: %#v", headers["Authorization"])
	}
	if len(headers["X-Request-Id"]) != 1 || headers["X-Request-Id"][0] != "req-1" {
		t.Fatalf("map[]string x-request-id: %#v", headers["X-Request-Id"])
	}

	type loginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	s := d.MaskValue("req", loginReq{Username: "bob", Password: "x"}).(map[string]any)
	if s["password"] != "***" || s["username"] != "bob" {
		t.Fatalf("struct mask: %#v", s)
	}

	h := d.MaskValue("header", http.Header{
		"Authorization": []string{"Bearer token"},
		"X-Request-Id":  []string{"req-1"},
	}).(http.Header)
	if h.Get("Authorization") != "***" {
		t.Fatalf("header authorization: %v", h.Get("Authorization"))
	}
	if h.Get("X-Request-Id") != "req-1" {
		t.Fatalf("header x-request-id: %v", h.Get("X-Request-Id"))
	}
}

func TestDesensitizerSlogGroup(t *testing.T) {
	cfg := defaultDesensitize()
	d := logger.NewDesensitizer(&cfg)
	attr := d.MaskAttr(slog.Group("user",
		slog.String("name", "alice"),
		slog.String("token", "abc"),
	))
	group := attr.Value.Group()
	found := map[string]string{}
	for _, a := range group {
		found[a.Key] = a.Value.String()
	}
	if found["token"] != "***" || found["name"] != "alice" {
		t.Fatalf("group mask: %#v", found)
	}
}

func TestDesensitizeHandler(t *testing.T) {
	base := &attrCaptureHandler{}
	cfg := defaultDesensitize()
	logger.UpdateDesensitizer(&cfg)
	h := logger.NewDesensitizeHandler(base)
	l := logger.DefaultLogger(slog.New(h))

	l.Info("login", "password", "secret", "msg", "phone 13700001111")

	if base.get("password").String() != "***" {
		t.Fatalf("handler password: %v", base.get("password"))
	}
	if base.get("msg").String() != "phone 13700001111" {
		t.Fatalf("handler msg should not be pattern-masked: %v", base.get("msg"))
	}
}

func TestDesensitizeHandlerHotReload(t *testing.T) {
	base := &attrCaptureHandler{}
	h := logger.NewDesensitizeHandler(base)
	l := logger.DefaultLogger(slog.New(h))

	on := defaultDesensitize()
	logger.UpdateDesensitizer(&on)
	l.Info("before", "password", "secret123")
	if base.get("password").String() != "***" {
		t.Fatalf("enabled: got %v", base.get("password"))
	}

	off := defaultDesensitize()
	off.Disabled = true
	logger.UpdateDesensitizer(&off)
	l.Info("after", "password", "secret123")
	if base.get("password").String() != "secret123" {
		t.Fatalf("disabled hot-reload: got %v", base.get("password"))
	}

	on.KeepPrefix = 2
	on.KeepSuffix = 2
	logger.UpdateDesensitizer(&on)
	l.Info("keep", "password", "secret123")
	if base.get("password").String() != "se***23" {
		t.Fatalf("rules hot-reload: got %v", base.get("password"))
	}
}

func TestNewSlogHandlerDesensitizeDefaultOn(t *testing.T) {
	cfg := logger.DefaultConfig
	h := logger.NewSlogHandler(&cfg)
	if _, ok := h.(*logger.DesensitizeHandler); !ok {
		t.Fatal("expected desensitize wrapper")
	}
	if logger.CurrentDesensitizer() == nil {
		t.Fatal("expected shared desensitizer to be set")
	}
}

func TestNewSlogHandlerDesensitizeOff(t *testing.T) {
	cfg := logger.DefaultConfig
	cfg.Desensitize.Disabled = true
	h := logger.NewSlogHandler(&cfg)
	if _, ok := h.(*logger.DesensitizeHandler); !ok {
		t.Fatal("expected desensitize wrapper to remain for hot-reload")
	}
	if logger.CurrentDesensitizer() != nil {
		t.Fatal("expected shared desensitizer cleared when disabled")
	}
}
