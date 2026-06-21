package config

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func testValue(priority int, value any) *Value {
	return &Value{Priority: priority, Value: value, Sourcer: &testSource{}}
}

func TestPrefixIndexPrecedenceAndInvalidation(t *testing.T) {
	configs := &ConfigsWithSyncMap{}
	configs.Set("app.loadbalance", testValue(1, "base"))
	configs.Set("app.grpc.loadbalance", testValue(1, "protocol"))
	configs.Set("app.grpc.users.loadbalance", testValue(1, "service"))

	got := configs.GetAllWithPrefixs("app", "app.grpc", "app.grpc.users")
	require.Equal(t, "service", got["loadbalance"].Value)

	// Updating an existing key must not rebuild or invalidate the key index.
	configs.Set("app.grpc.users.loadbalance", testValue(1, "updated"))
	got = configs.GetAllWithPrefixs("app", "app.grpc", "app.grpc.users")
	require.Equal(t, "updated", got["loadbalance"].Value)

	// Deleting and adding keys must invalidate the lazy index.
	configs.Del("app.grpc.users.loadbalance")
	got = configs.GetAllWithPrefixs("app", "app.grpc", "app.grpc.users")
	require.Equal(t, "protocol", got["loadbalance"].Value)
	configs.Set("app.grpc.users.timeout", testValue(1, "1s"))
	got = configs.GetAllWithPrefixs("app.grpc.users")
	require.Equal(t, "1s", got["timeout"].Value)
}

func TestSourceConfigsConcurrentPriorityUpdates(t *testing.T) {
	stores := []SourceConfiger{
		&SourceConfigs{cfgs: make(map[string][]*Value)},
		&SourceConfigsWithSyncMap{},
	}
	for _, store := range stores {
		store := store
		t.Run(fmt.Sprintf("%T", store), func(t *testing.T) {
			var wg sync.WaitGroup
			for priority := 0; priority < 100; priority++ {
				priority := priority
				wg.Add(1)
				go func() {
					defer wg.Done()
					store.Set("key", testValue(priority, priority))
				}()
			}
			wg.Wait()
			got, ok := store.Get("key")
			require.True(t, ok)
			require.Equal(t, 99, got.Value)

			require.True(t, store.Set("key", testValue(99, "replacement")))
			got, ok = store.Get("key")
			require.True(t, ok)
			require.Equal(t, "replacement", got.Value)

			store.Del("key", "", 99)
			got, ok = store.Get("key")
			require.True(t, ok)
			require.Equal(t, 98, got.Value)
		})
	}
}

func TestPrefixIndexConcurrentInvalidation(t *testing.T) {
	configs := &ConfigsWithSyncMap{}
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		worker := worker
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				key := fmt.Sprintf("app.worker%d.key%d", worker, i%20)
				configs.Set(key, testValue(1, i))
				if i%3 == 0 {
					configs.Del(key)
				}
				_ = configs.GetAllWithPrefixs("app.worker")
			}
		}()
	}
	wg.Wait()
	for worker := 0; worker < 8; worker++ {
		for i := 0; i < 20; i++ {
			configs.Set(fmt.Sprintf("app.worker%d.key%d", worker, i), testValue(1, i))
		}
	}
	require.Len(t, configs.GetAllWithPrefixs("app.worker"), 160)
}

func TestGetStringDefaultOptionsAllocation(t *testing.T) {
	require.NoError(t, Set("allocation_test_key", "value"))
	require.Equal(t, "value", GetString("allocation_test_key", ""))
	allocs := testing.AllocsPerRun(1000, func() {
		_ = GetString("allocation_test_key", "")
	})
	require.LessOrEqual(t, allocs, float64(1))
}

type reentrantSource struct {
	name    string
	manager *ConfigManager
}

func (s *reentrantSource) GetAll() map[string]*Value { return nil }
func (s *reentrantSource) Priority() int             { return 1 }
func (s *reentrantSource) Name() string              { return s.name }
func (s *reentrantSource) Set(string, any) error {
	s.manager.delSourcer(s.name)
	return nil
}
func (s *reentrantSource) Disconnect() { s.manager.delSourcer(s.name) }

func TestManagerDoesNotCallSourcesUnderLock(t *testing.T) {
	manager := &ConfigManager{sourcers: make(map[string]Sourcer)}
	source := &reentrantSource{name: "reentrant", manager: manager}
	manager.sourcers[source.name] = source

	done := make(chan struct{})
	go func() {
		defer close(done)
		require.NoError(t, manager.setValueToSource("key", "", "value"))
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Set called a source while holding the manager lock")
	}

	manager.sourcers[source.name] = source
	done = make(chan struct{})
	go func() {
		defer close(done)
		manager.disconnect()
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Disconnect called a source while holding the manager lock")
	}
}

func benchmarkPrefixStore(size int) *ConfigsWithSyncMap {
	configs := &ConfigsWithSyncMap{}
	for i := 0; i < size; i++ {
		prefix := "unrelated"
		switch i % 20 {
		case 0:
			prefix = "app"
		case 1:
			prefix = "app.grpc"
		case 2:
			prefix = "app.grpc.users"
		}
		configs.Set(fmt.Sprintf("%s.key%05d", prefix, i), testValue(1, i))
	}
	return configs
}

func BenchmarkConfigPrefixLookup(b *testing.B) {
	for _, size := range []int{100, 1_000, 10_000} {
		configs := benchmarkPrefixStore(size)
		b.Run(fmt.Sprintf("indexed/keys=%d/prefixes=1", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = configs.GetAllWithPrefixs("app.grpc.users")
			}
		})
		b.Run(fmt.Sprintf("indexed/keys=%d/prefixes=3", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = configs.GetAllWithPrefixs("app", "app.grpc", "app.grpc.users")
			}
		})
		b.Run(fmt.Sprintf("legacy/keys=%d/prefixes=3", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = legacyPrefixLookup(configs, "app", "app.grpc", "app.grpc.users")
			}
		})
	}
}

func legacyPrefixLookup(configs *ConfigsWithSyncMap, prefixes ...string) map[string]*Value {
	result := make(map[string]*Value)
	for _, prefix := range prefixes {
		configs.cfgs.Range(func(key, value any) bool {
			name := key.(string)
			if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
				result[strings.TrimPrefix(name, prefix+".")] = value.(*Value)
			}
			return true
		})
	}
	return result
}

func BenchmarkConfigGetStringOptions(b *testing.B) {
	require.NoError(b, Set("benchmark_get_string_options", "value"))
	b.Run("default", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = GetString("benchmark_get_string_options", "")
		}
	})
	b.Run("with-option", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = GetString("benchmark_get_string_options", "", WithToUpper())
		}
	})
}

func BenchmarkConfigChainLookup(b *testing.B) {
	manager := &ConfigManager{globalCfgs: &ConfigsWithSyncMap{}}
	manager.globalCfgs.Set("chain.hit", testValue(1, "value"))
	for _, count := range []int{1, 3, 10} {
		keys := make([]string, count)
		for i := range keys {
			keys[i] = "chain.miss." + strconv.Itoa(i)
		}
		keys[count-1] = "chain.hit"
		opts := &Options{keys: keys}
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = manager.getValue("chain.base", opts)
			}
		})
	}
}

func BenchmarkConfigStore(b *testing.B) {
	b.Run("new-key", func(b *testing.B) {
		configs := &ConfigsWithSyncMap{}
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			configs.Set("key."+strconv.Itoa(i), testValue(1, i))
		}
	})
	b.Run("existing-key", func(b *testing.B) {
		configs := &ConfigsWithSyncMap{}
		configs.Set("key", testValue(1, 0))
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			configs.Set("key", testValue(1, i))
		}
	})
}

func BenchmarkConfigMixedReadWrite(b *testing.B) {
	configs := &ConfigsWithSyncMap{}
	for i := 0; i < 100; i++ {
		configs.Set("key."+strconv.Itoa(i), testValue(1, i))
	}
	var sequence atomic.Uint64
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := sequence.Add(1)
			key := "key." + strconv.Itoa(int(n%100))
			if n%20 == 0 {
				configs.Set(key, testValue(1, n))
			} else {
				_, _ = configs.Get(key)
			}
		}
	})
}
