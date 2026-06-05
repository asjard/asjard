package xamqp

import (
	"github.com/rabbitmq/amqp091-go"
)

// PublishOptions holds the configuration for sending a message to RabbitMQ.
type PublishOptions struct {
	// Mandatory: If true, the server returns an unroutable message with a Return method.
	// If false, the server silently drops unroutable messages.
	Mandatory bool

	// Immediate: If true, the server returns an undeliverable message if it cannot
	// be consumed immediately. (Note: Not supported by modern RabbitMQ versions).
	Immediate bool

	// Exchange: The name of the exchange to publish the message to.
	Exchange string

	// Key: The routing key used by the exchange to determine which queues receive the message.
	Key string

	// Application or exchange specific fields,
	// the headers exchange will inspect this field.
	Headers amqp091.Table

	// Properties
	DeliveryMode  uint8  // Transient (0 or 1) or Persistent (2)
	Priority      uint8  // 0 to 9
	CorrelationId string // correlation identifier
	ReplyTo       string // address to to reply to (ex: RPC)
	// Expiration represents the message TTL in milliseconds. A value of "0"
	// indicates that the message will immediately expire if the message arrives
	// at its destination and the message is not directly handled by a consumer
	// that currently has the capacatity to do so. If you wish the message to
	// not expire on its own, set this value to any ttl value, empty string or
	// use the corresponding constants NeverExpire and ImmediatelyExpire. This
	// does not influence queue configured TTL values.
	Expiration string
	MessageId  string // message identifier
	Type       string // message type name
	UserId     string // creating user id - ex: "guest"
	AppId      string // creating application id
}

// PublishOption defines a function signature used to modify PublishOptions.
type PublishOption func(opts *PublishOptions)

// WithPublishMandatory ensures the message is returned to the sender if it
// cannot be routed to any queue.
func WithPublishMandatory() PublishOption {
	return func(opts *PublishOptions) {
		opts.Mandatory = true
	}
}

// WithPublishImmediate tells the server to return the message if it cannot
// be delivered to a consumer immediately.
func WithPublishImmediate() PublishOption {
	return func(opts *PublishOptions) {
		opts.Immediate = true
	}
}

// WithPublishExchange sets the destination exchange for the message.
func WithPublishExchange(exchange string) PublishOption {
	return func(opts *PublishOptions) {
		opts.Exchange = exchange
	}
}

// WithPublishKey sets the routing key for the message.
func WithPublishKey(key string) PublishOption {
	return func(opts *PublishOptions) {
		opts.Key = key
	}
}

func WithPublishExpiration(expiration string) PublishOption {
	return func(opts *PublishOptions) {
		opts.Expiration = expiration
	}
}

func WithPublishDeliveryMode(deliveryMode uint8) PublishOption {
	return func(opts *PublishOptions) {
		opts.DeliveryMode = deliveryMode
	}
}
