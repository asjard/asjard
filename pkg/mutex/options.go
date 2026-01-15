package mutex

import (
	"math"
	"time"

	"github.com/asjard/asjard/utils"
)

// LockOptions defines the configuration parameters for acquiring a distributed lock.
type LockOptions struct {
	// expiresIn specifies how long the lock remains valid before automatically expiring.
	// This is used as a safety mechanism to prevent deadlocks if a process crashes.
	expiresIn utils.JSONDuration

	// maxRetries defines the total number of attempts to acquire the lock.
	maxRetries int

	// minRetryDelayDuration is the lower bound for the random backoff between retries.
	minRetryDelayDuration time.Duration

	// maxRetryDelayDuration is the upper bound for the random backoff between retries.
	maxRetryDelayDuration time.Duration

	// threadId uniquely identifies the owner of the lock.
	// Only the holder of this ID can extend or release the lock.
	threadId string
}

// LockOption is a function type used to modify LockOptions.
type LockOption func(options *LockOptions)

// WithExpiresIn sets a custom expiration duration for the lock.
// Example: mutex.WithExpiresIn(10 * time.Second)
func WithExpiresIn(expiresIn time.Duration) LockOption {
	return func(options *LockOptions) {
		options.expiresIn = utils.JSONDuration{Duration: expiresIn}
	}
}

// WithMaxRetries sets the maximum number of times to attempt acquiring the lock.
// If maxRetries is less than 0, it will retry indefinitely (math.MaxInt).
func WithMaxRetries(maxRetries int) LockOption {
	return func(options *LockOptions) {
		if maxRetries < 0 {
			options.maxRetries = math.MaxInt
			return
		}
		options.maxRetries = maxRetries
	}
}

// WithMinRetryDelayDuration sets the minimum wait time before retrying a failed lock attempt.
func WithMinRetryDelayDuration(duration time.Duration) LockOption {
	return func(options *LockOptions) {
		options.minRetryDelayDuration = duration
	}
}

// WithMaxRetryDelayDuration sets the maximum wait time before retrying a failed lock attempt.
// This helps implement a jittered backoff strategy.
func WithMaxRetryDelayDuration(duration time.Duration) LockOption {
	return func(options *LockOptions) {
		options.maxRetryDelayDuration = duration
	}
}
