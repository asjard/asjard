package config

// EventType defines the nature of the change that occurred to a configuration entry.
type EventType int

const (
	// EventTypeCreate indicates a new configuration key has been added to a source.
	EventTypeCreate EventType = iota
	// EventTypeUpdate indicates an existing configuration value has been modified.
	EventTypeUpdate
	// EventTypeDelete indicates a configuration key has been removed from a source.
	EventTypeDelete
)

// Event encapsulates the details of a configuration change.
// It is used as the payload for internal communication between the ConfigManager and Listeners.
type Event struct {
	// Type specifies whether the configuration was created, updated, or deleted.
	Type EventType

	// Key is the unique identifier (property name) for the configuration setting.
	Key string

	// Value contains the actual data, metadata, and origin information for the event.
	// For Delete events, this may represent the state of the value prior to removal.
	Value *Value
}

// String provides a human-readable representation of the EventType.
// This is useful for logging and debugging configuration state transitions.
func (e EventType) String() string {
	switch e {
	case EventTypeCreate:
		return "Create"
	case EventTypeUpdate:
		return "Update"
	case EventTypeDelete:
		return "Delete"
	default:
		return "Unknown"
	}
}
