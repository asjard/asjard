package xredis

import (
	"context"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/pkg/mutex"
	"github.com/redis/go-redis/v9"
)

// Lock represents a distributed lock implementation using Redis as the backend.
type Lock struct {
	client *redis.Client
}

var (
	// Ensure Lock satisfies the mutex.Locker interface.
	_ mutex.Locker = &Lock{}

	// unlockScript is a Lua script that ensures atomicity when releasing a lock.
	// Logic: Only delete the key if the value matches the threadId (ARGV[1]).
	// This prevents a process from accidentally releasing a lock held by someone else.
	unlockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
    return 0
end`)

	// keepAliveScript is a Lua script that ensures atomicity when extending a lock.
	// Logic: Only update the expiration if the current owner matches the threadId.
	keepAliveScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("EXPIRE", KEYS[1], ARGV[2])
else
	return 0
end`)
)

// NewLock initializes a new Redis distributed lock instance.
// It retrieves the Redis client using the provided functional options.
func NewLock(opts ...Option) (mutex.Locker, error) {
	client, err := Client(opts...)
	if err != nil {
		return nil, err
	}
	return &Lock{
		client: client,
	}, nil
}

// Lock attempts to acquire the lock using the Redis SETNX (Set if Not eXists) command.
// If the key doesn't exist, it sets the key to the threadId with an expiration time.
func (l Lock) Lock(ctx context.Context, key, threadId string, expiresIn time.Duration) bool {
	// SetNX provides an atomic "check-and-set" operation.
	ok, err := l.client.SetNX(ctx, key, threadId, expiresIn).Result()
	if err != nil {
		logger.Error("redis lock fail", "key", key, "thread_id", threadId, "err", err)
	}
	return ok
}

// Unlock releases the lock. It uses a Lua script to ensure that the operation is atomic
// and that only the thread that originally acquired the lock can release it.
func (l Lock) Unlock(ctx context.Context, key, threadId string) bool {
	resp, err := unlockScript.Run(ctx, l.client, []string{key}, threadId).Int()
	if err != nil || resp == 0 {
		logger.Error("redis unlock fail", "key", key, "thread_id", threadId, "resp", resp, "err", err)
		return false
	}
	return true
}

// KeepAlive extends the expiration time of the lock.
// This is critical for preventing a lock from expiring during long-running tasks.
// The Lua script ensures that only the current owner can renew the lease.
func (l Lock) KeepAlive(ctx context.Context, key, threadId string, expiresIn time.Duration) bool {
	// expiresIn.Seconds() is passed to the EXPIRE command in the Lua script.
	resp, err := keepAliveScript.Run(ctx, l.client, []string{key}, threadId, expiresIn.Seconds()).Int()
	if err != nil || resp == 0 {
		logger.Error("redis keepalive fail", "key", key, "thread_id", threadId, "resp", resp, "err", err)
		return false
	}
	return true
}
