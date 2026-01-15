package xasynq

import (
	"github.com/asjard/asjard/pkg/stores/xredis"
	"github.com/redis/go-redis/v9"
)

// RedisConn is a wrapper around the go-redis client.
// It is designed to satisfy the asynq.RedisConnOpt interface, allowing
// the Asynq server to utilize a pre-existing Redis client instance.
type RedisConn struct {
	client *redis.Client
}

// MakeRedisClient satisfies the requirement for Asynq to retrieve
// the underlying redis.Client. It returns an interface{} as per
// the library's internal connection factory signature.
func (r RedisConn) MakeRedisClient() interface{} {
	return r.client
}

// NewRedisConn initializes a Redis connection by looking up a client
// defined in the Asjard configuration by its name.
// This allows the Asynq server to share the same Redis connection pool
// as the rest of the application.
func NewRedisConn(clientName string) (*RedisConn, error) {
	// Request a client from the central xredis store using functional options.
	client, err := xredis.NewClient(xredis.WithClientName(clientName))
	if err != nil {
		return nil, err
	}

	return &RedisConn{client: client}, nil
}
