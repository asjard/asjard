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
	if d := logger.NewDesensitizer(&cfg); d != nil {
		t.Fatal("expected nil desensitizer when disabled")
	}
	if d := logger.NewDesensitizer(nil); d != nil {
		t.Fatal("expected nil desensitizer for nil config")
	}
}

func TestDesensitizerEmptyRules(t *testing.T) {
	d := logger.NewDesensitizer(&logger.DesensitizeConfig{
		ReplacementConfig: logger.ReplacementConfig{Replacement: "***"},
	})
	if d != nil {
		t.Fatal("expected nil when no keys/patterns")
	}
}

func TestDesensitizerDefaultBuiltinKeys(t *testing.T) {
	// DefaultConfig: global keepPrefix=3 keepSuffix=4;
	// password/pwd/pwd_md5 use KeepPrefix/Suffix=-1 to force full mask (no inherit).
	cfg := defaultDesensitize()
	d := logger.NewDesensitizer(&cfg)
	if d == nil {
		t.Fatal("expected desensitizer")
	}

	if got := d.MaskValue("password", "secret123"); got != "***" {
		t.Fatalf("password forced full mask: got %v", got)
	}
	if got := d.MaskValue("pwd", "abcdef"); got != "***" {
		t.Fatalf("pwd forced full mask: got %v", got)
	}
	// authorization inherits global 3/4
	if got := d.MaskValue("Authorization", "Bearer abc"); got != "Bea*** abc" {
		t.Fatalf("authorization inherit global keep: got %v", got)
	}
	if got := d.MaskValue("mobile", "13800138000"); got != "138***8000" {
		t.Fatalf("mobile inherit global keep: got %v", got)
	}
	if got := d.MaskValue("note", "call 13800138000 please"); got != "call 13800138000 please" {
		t.Fatalf("non-sensitive key should stay plain: got %v", got)
	}
}

func TestDesensitizerKeepPrefixSuffix(t *testing.T) {
	d := logger.NewDesensitizer(&logger.DesensitizeConfig{
		ReplacementConfig: logger.ReplacementConfig{
			Replacement: "***",
			KeepPrefix:  3,
			KeepSuffix:  4,
		},
		Keys: []logger.DesensitizeKey{
			{Name: "mobile"},
			{Name: "token", ReplacementConfig: logger.ReplacementConfig{KeepPrefix: -1, KeepSuffix: -1}},
		},
	})

	if got := d.MaskValue("mobile", "13800138000"); got != "138***8000" {
		t.Fatalf("global keep: got %v", got)
	}
	// length <= keepPrefix+keepSuffix => replacement only
	if got := d.MaskValue("mobile", "1234567"); got != "***" {
		t.Fatalf("short value: got %v", got)
	}
	if got := d.MaskValue("mobile", "密文测试数据一二"); got != "密文测***数据一二" {
		t.Fatalf("unicode keep: got %v", got)
	}
	if got := d.MaskValue("token", "abcdefghij"); got != "***" {
		t.Fatalf("KeepPrefix=-1 forces full mask: got %v", got)
	}
}

func TestDesensitizerPerKeyAndPatternOverride(t *testing.T) {
	d := logger.NewDesensitizer(&logger.DesensitizeConfig{
		ReplacementConfig: logger.ReplacementConfig{
			Replacement: "***",
			KeepPrefix:  1,
			KeepSuffix:  1,
		},
		Keys: []logger.DesensitizeKey{
			{Name: "pin"}, // 0/0 => inherit global 1/1
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
	// pattern-only entry does not register name key; non-matching value stays plain
	if got := d.MaskValue("code", "plain-text"); got != "plain-text" {
		t.Fatalf("pattern-only entry should not mask by name: got %v", got)
	}
}

func TestDesensitizerInvalidPatternSkipped(t *testing.T) {
	d := logger.NewDesensitizer(&logger.DesensitizeConfig{
		ReplacementConfig: logger.ReplacementConfig{Replacement: "***"},
		Keys: []logger.DesensitizeKey{
			{Name: "bad", Pattern: `(`},
			{Name: "password", ReplacementConfig: logger.ReplacementConfig{KeepPrefix: -1, KeepSuffix: -1}},
		},
	})
	if d == nil {
		t.Fatal("expected desensitizer from valid key")
	}
	if got := d.MaskValue("msg", "CODE-1"); got != "CODE-1" {
		t.Fatalf("invalid pattern should be ignored: got %v", got)
	}
	if got := d.MaskValue("password", "x"); got != "***" {
		t.Fatalf("valid key still works: got %v", got)
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
		"password": "p@ssword1",
		"profile": map[string]string{
			"mobile": "13900001111",
		},
	}).(map[string]any)
	if m["password"] != "***" {
		t.Fatalf("nested password full mask: %v", m["password"])
	}
	profile := m["profile"].(map[string]string)
	if profile["mobile"] != "139***1111" {
		t.Fatalf("nested mobile keep 3/4: %v", profile["mobile"])
	}

	headers := d.MaskValue("header", map[string][]string{
		"Authorization": {"BearerTokenValue"},
		"X-Request-Id":  {"req-1"},
	}).(map[string][]string)
	if len(headers["Authorization"]) != 1 || headers["Authorization"][0] != "Bea***alue" {
		t.Fatalf("map[]string authorization: %#v", headers["Authorization"])
	}
	if len(headers["X-Request-Id"]) != 1 || headers["X-Request-Id"][0] != "req-1" {
		t.Fatalf("map[]string x-request-id: %#v", headers["X-Request-Id"])
	}

	type loginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Mobile   string `json:"mobile"`
	}
	s := d.MaskValue("req", loginReq{
		Username: "bob",
		Password: "x",
		Mobile:   "13800138000",
	}).(map[string]any)
	if s["password"] != "***" || s["username"] != "bob" || s["mobile"] != "138***8000" {
		t.Fatalf("struct mask: %#v", s)
	}

	h := d.MaskValue("header", http.Header{
		"Authorization": []string{"BearerTokenValue"},
		"X-Request-Id":  []string{"req-1"},
	}).(http.Header)
	if h.Get("Authorization") != "Bea***alue" {
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
		slog.String("token", "tok_abcdefgh"),
		slog.String("password", "secret"),
	))
	group := attr.Value.Group()
	found := map[string]string{}
	for _, a := range group {
		found[a.Key] = a.Value.String()
	}
	if found["name"] != "alice" {
		t.Fatalf("name: %#v", found)
	}
	if found["token"] != "tok***efgh" {
		t.Fatalf("token keep 3/4: %#v", found)
	}
	if found["password"] != "***" {
		t.Fatalf("password full mask: %#v", found)
	}
}

func TestDesensitizeHandler(t *testing.T) {
	base := &attrCaptureHandler{}
	cfg := defaultDesensitize()
	logger.UpdateDesensitizer(&cfg)
	h := logger.NewDesensitizeHandler(base)
	l := logger.DefaultLogger(slog.New(h))

	l.Info("login", "password", "secret", "mobile", "13800138000", "msg", "hello")

	if base.get("password").String() != "***" {
		t.Fatalf("handler password: %v", base.get("password"))
	}
	if base.get("mobile").String() != "138***8000" {
		t.Fatalf("handler mobile: %v", base.get("mobile"))
	}
	if base.get("msg").String() != "hello" {
		t.Fatalf("handler msg: %v", base.get("msg"))
	}
}

func TestDesensitizeHandlerHotReload(t *testing.T) {
	base := &attrCaptureHandler{}
	h := logger.NewDesensitizeHandler(base)
	l := logger.DefaultLogger(slog.New(h))

	on := defaultDesensitize()
	logger.UpdateDesensitizer(&on)
	l.Info("before", "mobile", "13800138000")
	if base.get("mobile").String() != "138***8000" {
		t.Fatalf("enabled: got %v", base.get("mobile"))
	}

	off := defaultDesensitize()
	off.Disabled = true
	logger.UpdateDesensitizer(&off)
	l.Info("after", "mobile", "13800138000")
	if base.get("mobile").String() != "13800138000" {
		t.Fatalf("disabled hot-reload: got %v", base.get("mobile"))
	}

	// override mobile keep via Keys (builtin mobile would inherit global)
	reloaded := logger.DesensitizeConfig{
		ReplacementConfig: logger.ReplacementConfig{
			Replacement: "***",
			KeepPrefix:  1,
			KeepSuffix:  1,
		},
		Keys: []logger.DesensitizeKey{{
			Name: "mobile",
			ReplacementConfig: logger.ReplacementConfig{
				KeepPrefix: 2,
				KeepSuffix: 2,
			},
		}},
	}
	logger.UpdateDesensitizer(&reloaded)
	l.Info("keep", "mobile", "13800138000")
	if base.get("mobile").String() != "13***00" {
		t.Fatalf("rules hot-reload: got %v", base.get("mobile"))
	}
}

func TestDesensitizeHandlerWithAttrs(t *testing.T) {
	base := &attrCaptureHandler{}
	cfg := defaultDesensitize()
	logger.UpdateDesensitizer(&cfg)
	h := logger.NewDesensitizeHandler(base).WithAttrs([]slog.Attr{
		slog.String("password", "secret"),
	})
	// captureHandler ignores WithAttrs; ensure API does not panic and Enabled works
	if !h.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("expected enabled")
	}
	_ = h.WithGroup("g")
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

func TestDesensitizerNilReceiver(t *testing.T) {
	var d *logger.Desensitizer
	if got := d.MaskValue("password", "x"); got != "x" {
		t.Fatalf("nil MaskValue: got %v", got)
	}
	attr := d.MaskAttr(slog.String("password", "x"))
	if attr.Value.String() != "x" {
		t.Fatalf("nil MaskAttr: got %v", attr.Value.String())
	}
}
