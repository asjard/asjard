package config

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListenerSupportsMultipleCallbacks(t *testing.T) {
	listener := newListener()
	var directCalls atomic.Int32
	for range 2 {
		listener.watch("config.key", &watchOptions{callback: func(*Event) {
			directCalls.Add(1)
		}})
	}
	listener.notify(&Event{Key: "config.key"})
	require.Equal(t, int32(2), directCalls.Load())

	var patternCalls atomic.Int32
	for range 2 {
		listener.watch("", &watchOptions{pattern: `^cache\..*`, callback: func(*Event) {
			patternCalls.Add(1)
		}})
	}
	listener.notify(&Event{Key: "cache.enabled"})
	require.Equal(t, int32(2), patternCalls.Load())

	listener.remove(`^cache\..*`)
	listener.notify(&Event{Key: "cache.client"})
	require.Equal(t, int32(2), patternCalls.Load())
}

func TestListenerConcurrentRegistration(t *testing.T) {
	listener := newListener()
	const callbackCount = 100
	var calls atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < callbackCount; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			listener.watch("", &watchOptions{
				pattern: fmt.Sprintf(`^shared\.%d$`, i%2),
				callback: func(*Event) {
					calls.Add(1)
				},
			})
		}()
	}
	wg.Wait()
	listener.notify(&Event{Key: "shared.0"})
	require.Equal(t, int32(callbackCount/2), calls.Load())
}
