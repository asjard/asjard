/*
Package mutex 分布式锁
*/
package mutex

import (
	"context"
	"math/rand"
	"time"

	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/utils"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
)

// Locker 互斥锁需要实现的方法
type Locker interface {
	// Lock 加锁
	Lock(ctx context.Context, key, threadId string, expiresIn time.Duration) bool
	// Unlock 解锁
	Unlock(ctx context.Context, key, threadId string) bool
	// 续期
	KeepAlive(ctx context.Context, key, threadId string, expiresIn time.Duration) bool
}

// noCopy may be added to structs which must not be copied
// after the first use.
//
// See https://golang.org/issues/8005#issuecomment-190753527
// for details.
//
// Note that it must not be embedded, due to the Lock and Unlock methods.
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

// Mutex 互斥锁
type Mutex struct {
	noCopy noCopy
	Locker Locker
}

// Lock 加锁
func (m *Mutex) Lock(ctx context.Context, key, threadId string, expiresIn time.Duration) bool {
	key = m.resourceKey(key)
	return m.Locker.Lock(ctx, key, threadId, expiresIn)
}

// Unlock 解锁
func (m *Mutex) Unlock(ctx context.Context, key, threadId string) bool {
	key = m.resourceKey(key)
	return m.Locker.Unlock(ctx, key, threadId)
}

// TryLock
func (m *Mutex) TryLock(ctx context.Context, key string, do func() error, opts ...LockOption) error {
	if m.Locker == nil {
		return status.Error(codes.Internal, "locker is must")
	}
	options := m.defaultLockOptions()
	for _, opt := range opts {
		opt(options)
	}
	key = m.resourceKey(key)
	for i := 0; i < options.maxRetries; i++ {
		if m.Locker.Lock(ctx, key, options.threadId, options.expiresIn.Duration) {
			defer m.Locker.Unlock(ctx, key, options.threadId)
			// 续期,防止do时间过长导致锁被自动释放了
			exit := make(chan struct{})
			defer close(exit)
			go func() {
				for {
					select {
					case <-exit:
						return
					case <-time.After(options.expiresIn.Duration - (options.expiresIn.Duration / 3)):
						m.Locker.KeepAlive(ctx, key, options.threadId, options.expiresIn.Duration)
					}
				}
			}()
			return do()
		}
		if i == options.maxRetries-1 {
			break
		}
		time.Sleep(time.Duration(rand.Int63n(int64(options.maxRetryDelayDuration-options.minRetryDelayDuration))) + options.minRetryDelayDuration)
	}
	return status.Errorf(status.GetLockFailCode, "get lock after %d retries not success", options.maxRetries)
}

func (m *Mutex) resourceKey(key string) string {
	return runtime.GetAPP().ResourceKey("lock", key,
		runtime.WithoutAz(true),
		runtime.WithoutVersion(true),
		runtime.WithDelimiter(":"))
}

func (m *Mutex) defaultLockOptions() *LockOptions {
	return &LockOptions{
		expiresIn:             utils.JSONDuration{Duration: 5 * time.Minute},
		maxRetries:            1,
		threadId:              uuid.NewString(),
		minRetryDelayDuration: 50 * time.Millisecond,
		maxRetryDelayDuration: 250 * time.Millisecond,
	}
}
