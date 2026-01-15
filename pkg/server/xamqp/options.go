package xamqp

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
