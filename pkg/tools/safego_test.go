package tools

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSafeGoRunsAndRecovers(t *testing.T) {
	done := make(chan struct{})
	SafeGo(func() { close(done) })
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("SafeGo did not execute")
	}

	recovered := make(chan struct{})
	SafeGo(func() {
		defer close(recovered)
		panic("boom")
	})
	select {
	case <-recovered:
	case <-time.After(time.Second):
		t.Fatal("SafeGo did not recover panic")
	}
	require.NotPanics(t, func() {})
}
