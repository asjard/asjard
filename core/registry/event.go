package registry

// EventType defines the nature of the change occurring within the registry.
type EventType int

const (
	// EventTypeCreate indicates a brand new service instance has been discovered.
	EventTypeCreate EventType = 0

	// EventTypeUpdate indicates an existing service instance has changed its
	// metadata or status (e.g., changing from 'starting' to 'healthy').
	EventTypeUpdate EventType = 1

	// EventTypeDelete indicates a service instance has been removed or has
	// failed health checks and should no longer receive traffic.
	EventTypeDelete EventType = 2
)

// Event encapsulates the details of a registry change.
// This is the object passed to watchers and listeners throughout the framework.
type Event struct {
	// Type specifies whether the instance was created, updated, or deleted.
	Type EventType

	// Instance contains the full details of the service involved in the event,
	// including its ID, name, addresses, and discovery source.
	Instance *Instance
}
