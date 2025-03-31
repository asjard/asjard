package xasynq

import (
	"github.com/asjard/asjard/pkg/stores/xredis"
	"github.com/redis/go-redis/v9"
)

type RedisConn struct {
	client *redis.Client
}

func (r RedisConn) MakeRedisClient() interface{} {
	return r.client
}

// NewRedisConn 通过redis名称获取redis conn
func NewRedisConn(clientName string) (*RedisConn, error) {
	client, err := xredis.Client(xredis.WithClientName(clientName))
	if err != nil {
		return nil, err
	}
	return &RedisConn{client: client}, nil
}
