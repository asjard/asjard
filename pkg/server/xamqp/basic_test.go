package xamqp

import (
	"context"
	"testing"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
)

func TestPublishOptions(t *testing.T) {
	opts := &PublishOptions{}
	for _, option := range []PublishOption{
		WithPublishMandatory(), WithPublishImmediate(), WithPublishExchange("events"), WithPublishKey("created"),
		WithPublishExpiration("5000"), WithPublishDeliveryMode(2),
	} {
		option(opts)
	}
	require.True(t, opts.Mandatory)
	require.True(t, opts.Immediate)
	require.Equal(t, "events", opts.Exchange)
	require.Equal(t, "created", opts.Key)
	require.Equal(t, "5000", opts.Expiration)
	require.Equal(t, uint8(2), opts.DeliveryMode)
}

func TestContextBody(t *testing.T) {
	c := Context{Context: context.Background(), task: amqp091.Delivery{Body: []byte("payload")}}
	require.Equal(t, []byte("payload"), c.Body())
	require.NoError(t, c.Err())
}
