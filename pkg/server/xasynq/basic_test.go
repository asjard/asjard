package xasynq

import (
	"context"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestContextTask(t *testing.T) {
	task := asynq.NewTask("email:send", []byte("payload"))
	c := Context{Context: context.Background(), task: task}
	require.Equal(t, "email:send", c.Type())
	require.Equal(t, []byte("payload"), c.Payload())
}

func TestDefaultGlobalHandler(t *testing.T) {
	h := defaultGlobalHandler{}
	require.NotNil(t, h.BaseContext())
	require.Nil(t, h.RetryDelayFunc())
	require.NotNil(t, h.IsFailure())
	require.NotNil(t, h.HealthCheckFunc())
	require.NotNil(t, h.ErrorHandler())
	require.Nil(t, h.GroupAggregator())
}
