package snowflake

import (
	"os"
	"testing"

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

func TestSnowflakeLifecycle(t *testing.T) {
	Node = nil
	require.NoError(t, (SnowFlake{}).Start())
	require.NotNil(t, Node)
	require.NotEqual(t, Node.Generate().Int64(), Node.Generate().Int64())
	(SnowFlake{}).Stop()
}
