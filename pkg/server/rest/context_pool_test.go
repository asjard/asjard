package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestContextPoolReuseAfterClose(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	c := NewContext(req, rec)
	c.SetUserValue("prev", "value")
	c.Close()

	c2 := NewContext(req, rec)
	require.Nil(t, c2.GetUserValue("prev"))
	require.NotPanics(t, func() {
		c2.SetUserValue("next", "ok")
	})
	require.Equal(t, "ok", c2.GetUserValue("next"))
	c2.Close()
}

func TestSetUserValueUnlocksWhenMapWasNil(t *testing.T) {
	c := &Context{}
	require.NotPanics(t, func() {
		c.SetUserValue("k", "v")
	})
	require.Equal(t, "v", c.GetUserValue("k"))

	done := make(chan struct{})
	go func() {
		c.SetUserValue("k2", "v2")
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("SetUserValue blocked, mutex was likely left locked")
	}
}

func TestContextValueFallsThroughToRequestContext(t *testing.T) {
	type ctxKey struct{}
	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(
		context.WithValue(context.Background(), ctxKey{}, "from-request"),
	)
	c := NewContext(req, httptest.NewRecorder())
	t.Cleanup(c.Close)

	require.Equal(t, "from-request", c.Value(ctxKey{}))
	c.SetUserValue(ctxKey{}, "from-user")
	require.Equal(t, "from-user", c.Value(ctxKey{}))
}

func TestContextAfterClose(t *testing.T) {
	c := NewContext(httptest.NewRequest(http.MethodGet, "/test", nil), httptest.NewRecorder())
	c.Close()
	require.NotPanics(t, func() {
		_ = c.Context()
		_ = c.Value("missing")
	})
}
