package mutex

import (
	"math"
	"time"

	"github.com/asjard/asjard/utils"
)

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
