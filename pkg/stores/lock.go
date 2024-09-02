package stores

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/utils"
	"github.com/google/uuid"
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

// RWLocker 读写锁需要实现的方法
type RWLocker interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

// LockOptions 锁参数
type LockOptions struct {
	// 锁过期时间
	expiresIn utils.JSONDuration
	// 最大重试次数
	maxRetries int
	// 最小重试延迟时间
	minRetryDelayDuration time.Duration
	// 最大重试延迟时间
	maxRetryDelayDuration time.Duration

	// 线程ID
	threadId string
}

type LockOption func(options *LockOptions)

// WithExpiresIn 设置锁过期时间
func WithExpiresIn(expiresIn time.Duration) LockOption {
	return func(options *LockOptions) {
		options.expiresIn = utils.JSONDuration{Duration: expiresIn}
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) LockOption {
	return func(options *LockOptions) {
		if maxRetries < 0 {
			options.maxRetries = math.MaxInt
			return
		}
		options.maxRetries = maxRetries
	}
}

// WithMinRetryDelayDuration 最小重试延迟时间
func WithMinRetryDelayDuration(duration time.Duration) LockOption {
	return func(options *LockOptions) {
		options.minRetryDelayDuration = duration
	}
}

// WithMaxRetryDelayDuration 最大重试延迟时间
func WithMaxRetryDelayDuration(duration time.Duration) LockOption {
	return func(options *LockOptions) {
		options.maxRetryDelayDuration = duration
	}
}

// DoWithLock 加锁执行
func DoWithLock(ctx context.Context, key string, locker Locker, do func() error, opts ...LockOption) error {
	options := defaultLockOptions()
	for _, opt := range opts {
		opt(options)
	}
	key = runtime.GetAPP().ResourceKey("lock", key, runtime.WithoutRegion(true))
	for i := 0; i < options.maxRetries; i++ {
		if locker.Lock(ctx, key, options.threadId, options.expiresIn.Duration) {
			defer locker.Unlock(ctx, key, options.threadId)
			// 续期,防止do时间过长导致锁被自动释放了
			exit := make(chan struct{})
			defer close(exit)
			go func() {
				for {
					select {
					case <-exit:
						return
					case <-time.After(options.expiresIn.Duration - (options.expiresIn.Duration / 3)):
						locker.KeepAlive(ctx, key, options.threadId, options.expiresIn.Duration)
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

func defaultLockOptions() *LockOptions {
	return &LockOptions{
		expiresIn:             utils.JSONDuration{Duration: 5 * time.Minute},
		maxRetries:            1,
		threadId:              uuid.NewString(),
		minRetryDelayDuration: 50 * time.Millisecond,
		maxRetryDelayDuration: 250 * time.Millisecond,
	}
}
