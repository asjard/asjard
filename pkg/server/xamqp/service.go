package xamqp

import "github.com/asjard/asjard/core/server"

// HandlerFunc defines the signature for processing an AMQP delivery.
// It receives a custom AMQP Context, the service implementation (srv),
// and an interceptor for cross-cutting concerns like tracing or logging.
type HandlerFunc func(ctx *Context, srv any, interceptor server.UnaryServerInterceptor) (any, error)

// ServiceDesc contains the top-level metadata for an AMQP service.
type ServiceDesc struct {
	// ServiceName is the unique identifier for the service (e.g., "account.v1.listener").
	ServiceName string
	// HandlerType is a pointer to the interface the service must implement, used for reflection.
	HandlerType any
	// Methods is a list of specific queue consumers defined within this service.
	Methods []MethodDesc
}

// MethodDesc contains the detailed configuration for an individual AMQP consumer.
// It maps directly to the parameters used in RabbitMQ's QueueDeclare, ExchangeDeclare, and Consume methods.
type MethodDesc struct {
	// Queue is the name of the RabbitMQ queue to consume from.
	Queue string
	// Exchange is the name of the exchange to bind the queue to.
	Exchange string
	// Kind is the type of exchange (e.g., "direct", "topic", "fanout", "headers").
	Kind string
	// Route is the routing key for the binding between the exchange and queue.
	Route string
	// Consumer is a unique identifier for this specific consumer instance.
	Consumer string

	// AutoAck determines if messages are acknowledged automatically by the server
	// upon delivery (true) or manually by the code after processing (false).
	AutoAck bool
	// Durable ensures the queue/exchange survives a RabbitMQ server restart.
	Durable bool
	// AutoDelete removes the queue/exchange when the last consumer/queue unsubscribes.
	AutoDelete bool
	// Exclusive restricts the queue to the current connection only.
	Exclusive bool
	// NoLocal is an AMQP feature to prevent receiving messages published by the same connection.
	NoLocal bool
	// NoWait tells the server not to wait for a response from the broker for the declaration.
	NoWait bool
	// Internal exchanges cannot be published to directly by users, only by other exchanges.
	Internal bool

	// Table provides additional arguments for advanced features like TTL,
	// Max-Length, or Dead Letter Exchanges.
	Table map[string]any
	// Handler is the actual function that will be executed for each message received.
	Handler HandlerFunc
}
