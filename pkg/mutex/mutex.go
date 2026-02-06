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

// Locker defines the interface that underlying distributed storage
// (Redis, ETCD, etc.) must implement.
type Locker interface {
	// Lock attempts to acquire the lock. Returns true if successful.
	Lock(ctx context.Context, key, threadId string, expiresIn time.Duration) bool
	// Unlock releases the lock held by the specific threadId.
	Unlock(ctx context.Context, key, threadId string) bool
	// KeepAlive extends the lock's expiration time to prevent accidental release.
	KeepAlive(ctx context.Context, key, threadId string, expiresIn time.Duration) bool
}

// noCopy is a sentinel used to prevent the Mutex struct from being copied by value.
// Copying a mutex can lead to logical errors where two different instances
// think they are controlling the same lock state.
type noCopy struct{}

// Lock/Unlock on noCopy allows `go vet` to trigger a warning if the struct is copied.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

// Mutex is the high-level distributed lock controller.
type Mutex struct {
	noCopy noCopy
	Locker Locker // The implementation (e.g., RedisLocker or EtcdLocker)
}

// Lock manually acquires a lock for a specific duration.
func (m *Mutex) Lock(ctx context.Context, key, threadId string, expiresIn time.Duration) bool {
	key = m.resourceKey(key)
	return m.Locker.Lock(ctx, key, threadId, expiresIn)
}

// Unlock manually releases a lock.
func (m *Mutex) Unlock(ctx context.Context, key, threadId string) bool {
	key = m.resourceKey(key)
	return m.Locker.Unlock(ctx, key, threadId)
}

// TryLock executes the 'do' function if the lock is successfully acquired.
// It handles retries and includes a "Watchdog" goroutine to automatically
// extend the lock's life while 'do' is still running.
func (m *Mutex) TryLock(ctx context.Context, key string, do func() error, opts ...LockOption) error {
	if m.Locker == nil {
		return status.Error(codes.Internal, "locker is required")
	}

	options := m.defaultLockOptions()
	for _, opt := range opts {
		opt(options)
	}

	key = m.resourceKey(key)
	for i := 0; i < options.maxRetries; i++ {
		if m.Locker.Lock(ctx, key, options.threadId, options.expiresIn.Duration) {
			// Ensure the lock is released when 'do' completes or the function exits.
			defer m.Locker.Unlock(ctx, key, options.threadId)

			// Watchdog: Start a goroutine to keep the lock alive.
			exit := make(chan struct{})
			defer close(exit)
			go func() {
				for {
					select {
					case <-exit:
						return
					// Renew the lock when 2/3 of its TTL has elapsed.
					case <-time.After(options.expiresIn.Duration - (options.expiresIn.Duration / 3)):
						m.Locker.KeepAlive(ctx, key, options.threadId, options.expiresIn.Duration)
					}
				}
			}()

			// Execute the protected business logic.
			return do()
		}

		if i == options.maxRetries-1 {
			break
		}

		// Backoff with jitter to prevent "thundering herd" effect on the locker storage.
		time.Sleep(time.Duration(rand.Int63n(int64(options.maxRetryDelayDuration-options.minRetryDelayDuration))) + options.minRetryDelayDuration)
	}

	return status.Errorf(status.GetLockFailCode, "failed to acquire lock after %d retries", options.maxRetries)
}

// resourceKey generates a standardized, namespaced key (e.g., app:lock:my_key).
func (m *Mutex) resourceKey(key string) string {
	return runtime.GetAPP().ResourceKey("lock", key,
		runtime.WithoutAz(true),
		runtime.WithDelimiter(":"))
}

// defaultLockOptions provides sensible defaults for lock behavior.
func (m *Mutex) defaultLockOptions() *LockOptions {
	return &LockOptions{
		expiresIn:             utils.JSONDuration{Duration: 5 * time.Minute},
		maxRetries:            1,
		threadId:              uuid.NewString(),
		minRetryDelayDuration: 50 * time.Millisecond,
		maxRetryDelayDuration: 250 * time.Millisecond,
	}
}
