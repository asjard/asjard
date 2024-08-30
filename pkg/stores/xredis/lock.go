package xredis

import (
	"context"
	"time"

	"github.com/asjard/asjard/pkg/stores"
	"github.com/redis/go-redis/v9"
)

type Lock struct {
	client *redis.Client
}

var _ stores.Locker = &Lock{}

// NewLock 初始化redis分布式锁
func NewLock(opts ...Option) (stores.Locker, error) {
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
	return false
}

// Unlock 解锁
func (l Lock) Unlock(ctx context.Context, key, threadId string) {}
