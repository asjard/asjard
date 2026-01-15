package xamqp

import (
	"context"

	"github.com/streadway/amqp"
)

// Context wraps the standard Go context and the RabbitMQ delivery object.
// It provides a unified interface for handlers to access message data
// and manage request-scoped values or deadlines.
type Context struct {
	context.Context
	// task represents the raw delivery from the AMQP broker,
	// containing headers, routing info, and the payload.
	task amqp.Delivery
}

// Body returns the raw byte payload of the AMQP message.
// This is a helper method to easily access the message content
// without interacting directly with the underlying amqp.Delivery object.
func (c Context) Body() []byte {
	return c.task.Body
}
