package xgorm

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestLock(t *testing.T) {
	testDB := "lock"
	time.Sleep(50 * time.Millisecond)
	lock, err := NewLock(WithConnName(testDB))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	db, err := DB(context.Background(), WithConnName(testDB))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := db.AutoMigrate(&Lock{}); err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Run("normal", func(t *testing.T) {
		key := "test_lock_nomal"
		now := time.Now()
		threadId := uuid.NewString()
		expiresIn := 5 * time.Second
		// Lock
		if !lock.Lock(context.Background(), key, threadId, expiresIn) {
			t.Error("nomal lock fail")
			t.FailNow()
		}
		var record Lock
		if err := db.Where("lock_key=?", key).First(&record).Error; err != nil {
			t.Error(err)
			t.FailNow()
		}
		if record.Owner != threadId {
			t.Errorf("normal lock value not equal, act: '%s', want: '%s'", record.Owner, threadId)
			t.FailNow()
		}
		// keepAlive
		time.Sleep(time.Second)
		if !lock.KeepAlive(context.Background(), key, threadId, expiresIn) {
			t.Error("lock keepalive fail")
			t.FailNow()
		}
		if err := db.Where("lock_key=?", key).Where("owner=?", threadId).First(&record).Error; err != nil {
			t.Error(err)
			t.FailNow()
		}

		if record.ExpiresAt.Before(now.Add(expiresIn)) {
			t.Error("lock keepalive fail, ttl not refresh")
			t.FailNow()
		}
		// unLock
		if !lock.Unlock(context.Background(), key, threadId) {
			t.Error("unlock fail")
			t.FailNow()
		}
		if err := db.Where("lock_key=?", key).Where("owner=?", threadId).First(&Lock{}).Error; err == nil {
			t.Error("unlock fail, key exist")
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
		key := "test_lock_concurrency"
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
