package xredis

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

//gocyclo:ignore
func TestLock(t *testing.T) {
	lock, err := NewLock()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	client, err := Client()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Run("normal", func(t *testing.T) {
		key := "test_lock_nomal"
		threadId := uuid.NewString()
		expiresIn := 5 * time.Second
		// Lock
		if !lock.Lock(context.Background(), key, threadId, expiresIn) {
			t.Error("nomal lock fail")
			t.FailNow()
		}
		if v := client.Get(context.Background(), key).Val(); v != threadId {
			t.Errorf("normal lock value not equal, act: '%s', want: '%s'", v, threadId)
			t.FailNow()
		}
		// keepAlive
		time.Sleep(time.Second)
		if !lock.KeepAlive(context.Background(), key, threadId, expiresIn) {
			t.Error("lock keepalive fail")
			t.FailNow()
		}
		if v := client.TTL(context.Background(), key).Val(); v < expiresIn-time.Second {
			t.Error("lock keepalive fail, ttl not refresh")
			t.FailNow()
		}
		// unLock
		if !lock.Unlock(context.Background(), key, threadId) {
			t.Error("unlock fail")
			t.FailNow()
		}
		if v := client.Exists(context.Background(), key).Val(); v != 0 {
			t.Error("unlock fail, key exist", v)
			t.FailNow()
		}
	})
	t.Run("race", func(t *testing.T) {
		key := "test_lock_race"
		expiresIn := 5 * time.Second
		threadId := uuid.NewString()
		anotherThreadId := uuid.NewString()
		if !lock.Lock(context.Background(), key, threadId, expiresIn) {
			t.Error("lock fail")
			t.FailNow()
		}
		if lock.Lock(context.Background(), key, anotherThreadId, expiresIn) {
			t.Error("another thread lock success")
			t.FailNow()
		}
		if !lock.Unlock(context.Background(), key, threadId) {
			t.Error("unlock fail")
			t.FailNow()
		}
		if !lock.Lock(context.Background(), key, anotherThreadId, expiresIn) {
			t.Error("another thread lock fail")
			t.FailNow()
		}
	})
	t.Run("concurrency", func(t *testing.T) {
		key := "test_lock_thread"
		expiresIn := 5 * time.Second
		var count int64 = 0
		wg := sync.WaitGroup{}
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				threadId := uuid.NewString()
				for {
					if lock.Lock(context.Background(), key, threadId, expiresIn) {
						// 不是并发安全的
						count += 1
						lock.Unlock(context.Background(), key, threadId)
						break
					}
					time.Sleep(100 * time.Millisecond)
				}
			}(i)
		}
		wg.Wait()
		if count != 10 {
			t.Error("concurrency test fail")
			t.FailNow()
		}
	})
}
