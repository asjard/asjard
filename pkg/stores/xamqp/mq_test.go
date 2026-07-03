package xamqp

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/asjard/asjard/core/config"
	_ "github.com/asjard/asjard/pkg/config/mem"
	"github.com/rabbitmq/amqp091-go"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestClientOptions(t *testing.T) {
	opts := defaultClientOptions()
	require.Equal(t, DefaultClientName, opts.clientName)
	WithClientName("named")(opts)
	require.Equal(t, "named", opts.clientName)
	WithClientName("")(opts)
	require.Equal(t, "named", opts.clientName)
}

func TestEmptyClientConnection(t *testing.T) {
	conn := &ClientConn{}
	_, err := conn.Channel()
	require.Error(t, err)
	_, err = conn.NotifyClose(make(chan *amqp091.Error))
	require.Error(t, err)
}

func TestRabbitMQPublishConsumeIntegration(t *testing.T) {
	manager := &ClientManager{}
	var conn *amqp.Connection
	require.Eventually(t, func() bool {
		var err error
		conn, err = manager.newClient("integration", &ClientConnConfig{Url: "amqp://guest:guest@127.0.0.1:5672/"})
		return err == nil
	}, 30*time.Second, 500*time.Millisecond)
	t.Cleanup(func() { _ = conn.Close() })

	channel, err := conn.Channel()
	require.NoError(t, err)
	t.Cleanup(func() { _ = channel.Close() })
	queueName := fmt.Sprintf("asjard-test-%d", time.Now().UnixNano())
	queue, err := channel.QueueDeclare(queueName, false, true, true, false, nil)
	require.NoError(t, err)
	require.NoError(t, channel.PublishWithContext(context.Background(), "", queue.Name, false, false, amqp.Publishing{Body: []byte("payload")}))
	require.Eventually(t, func() bool {
		message, ok, err := channel.Get(queue.Name, true)
		return err == nil && ok && string(message.Body) == "payload"
	}, 10*time.Second, 100*time.Millisecond)
}
