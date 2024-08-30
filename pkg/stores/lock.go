package stores

import (
	"context"
	"math"
	"time"

	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/utils"
	"github.com/google/uuid"
)

// Locker 分布式锁需要实现的方法
type Locker interface {
	// Lock 加锁
	Lock(ctx context.Context, key, threadId string, expiresIn time.Duration) bool
	// Unlock 解锁
	Unlock(ctx context.Context, key, threadId string)
}

// LockOptions 锁参数
type LockOptions struct {
	// 锁过期时间
	expiresIn utils.JSONDuration
	// 最大重试次数
	maxRetries int

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

// DoWithLock 加锁执行
func DoWithLock(ctx context.Context, key string, locker Locker, do func() error, opts ...LockOption) error {
	options := defaultLockOptions()
	for _, opt := range opts {
		opt(options)
	}
	for i := 0; i < options.maxRetries; i++ {
		if locker.Lock(ctx, key, options.threadId, options.expiresIn.Duration) {
			defer locker.Unlock(ctx, key, options.threadId)
			return do()
		}
		if i == options.maxRetries-1 {
			break
		}
		// TODO 设置不同时间的间隔时间
		time.Sleep(time.Second)
	}
	return status.Errorf(status.GetLockFailCode, "get lock after %d retries not success", options.maxRetries)
}

func defaultLockOptions() *LockOptions {
	return &LockOptions{
		expiresIn:  utils.JSONDuration{Duration: 5 * time.Minute},
		maxRetries: 1,
		threadId:   uuid.NewString(),
	}
}
