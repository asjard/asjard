package mem

import (
	"fmt"
	"sync"
	"testing"

	"github.com/asjard/asjard/core/config"
	"github.com/stretchr/testify/require"
)

func TestMemLifecycleAndSnapshot(t *testing.T) {
	var events []*config.Event
	source, err := New(&config.SourceOptions{Callback: func(event *config.Event) {
		events = append(events, event)
	}})
	require.NoError(t, err)
	m := source.(*Mem)
	require.Equal(t, Name, m.Name())
	require.Equal(t, Priority, m.Priority())

	require.NoError(t, m.Set("key", "first"))
	require.NoError(t, m.Set("key", "second"))
	require.Len(t, events, 2)
	require.Equal(t, config.EventTypeUpdate, events[1].Type)
	require.Equal(t, "key", events[1].Key)
	require.Equal(t, "second", events[1].Value.Value)
	require.Same(t, m, events[1].Value.Sourcer)

	snapshot := m.GetAll()
	require.Equal(t, "second", snapshot["key"].Value)
	delete(snapshot, "key")
	require.Contains(t, m.GetAll(), "key")
	m.Disconnect()
}

func TestMemConcurrentAccess(t *testing.T) {
	m := &Mem{configs: make(map[string]any), options: &config.SourceOptions{Callback: func(*config.Event) {}}}
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			require.NoError(t, m.Set(key, i))
			_, _ = m.get(key)
			_ = m.GetAll()
		}(i)
	}
	wg.Wait()
	require.Len(t, m.GetAll(), 50)
}
