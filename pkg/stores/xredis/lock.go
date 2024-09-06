package xredis

import (
	"context"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/pkg/mutex"
	"github.com/redis/go-redis/v9"
)

// Lock redis实现的分布式锁
type Lock struct {
	client *redis.Client
}

var (
	_ mutex.Locker = &Lock{}
	// 解锁lua脚本
	unlockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
    return 0
end`)
	// 续期脚本
	keepAliveScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("EXPIRE", KEYS[1], ARGV[2])
else
	return 0
end`)
)

// NewLock 初始化redis分布式锁
func NewLock(opts ...Option) (mutex.Locker, error) {
	client, err := Client(opts...)
	if err != nil {
		return nil, err
	}
	return &Lock{
		client: client,
	}, nil
}

// Lock 加锁
func (l Lock) Lock(ctx context.Context, key, threadId string, expiresIn time.Duration) bool {
	ok, err := l.client.SetNX(ctx, key, threadId, expiresIn).Result()
	if err != nil {
		logger.Error("redis lock fail", "key", key, "thread_id", threadId, "err", err)
	}
	return ok
}

// Unlock 解锁
func (l Lock) Unlock(ctx context.Context, key, threadId string) bool {
	resp, err := unlockScript.Run(ctx, l.client, []string{key}, threadId).Int()
	if err != nil || resp == 0 {
		logger.Error("redis unlock fail", "key", key, "thread_id", threadId, "resp", resp, "err", err)
		return false
	}
	return true
}

// Keepalive 续期
func (l Lock) KeepAlive(ctx context.Context, key, threadId string, expiresIn time.Duration) bool {
	resp, err := keepAliveScript.Run(ctx, l.client, []string{key}, threadId, expiresIn.Seconds()).Int()
	if err != nil || resp == 0 {
		logger.Error("redis keepalive fail", "key", key, "thread_id", threadId, "resp", resp, "err", err)
		return false
	}
	return true
}
