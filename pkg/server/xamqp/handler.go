package xamqp

// Handler is the interface that all RabbitMQ consumer services must implement.
// It acts as a contract between the business logic and the AMQP server engine,
// ensuring the server can retrieve the necessary metadata to bind to queues
// and route incoming messages to the correct functions.
type Handler interface {
	// AmqpServiceDesc returns the service specification, which includes
	// information about the exchanges, queues, and the specific methods
	// that will handle incoming message deliveries.
	AmqpServiceDesc() *ServiceDesc
}
