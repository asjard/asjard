package xrabbitmq

import (
	"context"

	"github.com/streadway/amqp"
)

type Context struct {
	context.Context
	task amqp.Delivery
}

func (c Context) Body() []byte {
	return c.task.Body
}
