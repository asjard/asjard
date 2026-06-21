package xasynq

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/config"
	_ "github.com/asjard/asjard/pkg/config/mem"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestAsynqRedisIntegration(t *testing.T) {
	require.NoError(t, config.Set("asjard.stores.redis.clients.default.address", "127.0.0.1:6379"))
	require.NoError(t, config.Set("asjard.stores.redis.clients.default.options.dialTimeout", "3s"))
	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", "127.0.0.1:6379", 500*time.Millisecond)
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	}, 30*time.Second, 500*time.Millisecond)
	require.NoError(t, bootstrap.Bootstrap())
	conn, err := NewRedisConn("default")
	require.NoError(t, err)
	require.NotNil(t, conn.MakeRedisClient())
}
