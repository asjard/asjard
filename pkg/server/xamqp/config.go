package xamqp

import "github.com/asjard/asjard/core/server"

// Config defines the configuration parameters for the RabbitMQ server/consumer.
// It inherits base server settings such as listener addresses and enabled status.
type Config struct {
	server.Config

	// ClientName is the unique identifier for the connection,
	// visible in the RabbitMQ management dashboard.
	ClientName string `json:"clientName"`

	// PrefetchCount defines the maximum number of unacknowledged messages
	// the server will deliver to this consumer.
	// Setting this helps with load balancing and preventing a single
	// consumer from being overwhelmed.
	PrefetchCount int `json:"prefetchCount"`

	// PrefetchSize defines the maximum number of octets (bytes)
	// the server will deliver as unacknowledged messages.
	// 0 means no specific limit on the total byte size.
	PrefetchSize int `json:"prefetchSize"`

	// Global determines the scope of the prefetch limits.
	// If true, the QoS settings are applied to the entire connection/channel.
	// If false, the settings apply specifically to each new consumer on the channel.
	Global bool `json:"global"`
}

// defaultConfig provides the baseline settings for the RabbitMQ consumer.
func defaultConfig() Config {
	return Config{
		// Defaulting to 1 ensures that the consumer only processes one
		// message at a time, providing a safe "fair dispatch" model.
		PrefetchCount: 1,
	}
}
