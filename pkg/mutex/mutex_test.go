package mutex

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type testLock struct {
	val uint64
}

// Lock 加锁
func (l *testLock) Lock(ctx context.Context, key, threadId string, expiresIn time.Duration) bool {
	return atomic.CompareAndSwapUint64(&l.val, 0, 1)
}

// Unlock 解锁
func (l *testLock) Unlock(ctx context.Context, key, threadId string) bool {
	return atomic.CompareAndSwapUint64(&l.val, 1, 0)
}

// 续期
func (l *testLock) KeepAlive(ctx context.Context, key, threadId string, expiresIn time.Duration) bool {
	return true
}

func TestLock(t *testing.T) {
	t.Run("Lock", func(t *testing.T) {
		var count int64
		do := func() error {
			count += 1
			return nil
		}
		wg := sync.WaitGroup{}
		key := "test_do_with_lock"
		m := Mutex{Locker: &testLock{}}
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := m.TryLock(context.Background(), key, do, WithMaxRetries(-1)); err != nil {
					t.Error(err)
				}
			}()
		}
		wg.Wait()
		if count != 10 {
			t.Error("do with lock fail")
			t.FailNow()
		}
	})
}
