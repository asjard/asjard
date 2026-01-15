package snowflake

import (
	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/runtime"
	"github.com/bwmarrin/snowflake"
)

// Node is the global snowflake node instance used to generate IDs.
// It is thread-safe and should be accessed via Node.Generate().
var Node *snowflake.Node

// SnowFlake implements the bootstrap.Bootstrap interface.
// This allows the framework to automatically initialize the ID generator during startup.
type SnowFlake struct{}

func init() {
	// Register the Snowflake component into the bootstrap chain.
	bootstrap.AddBootstrap(&SnowFlake{})
}

// Start is called during the application's bootstrap phase.
// It initializes the Snowflake node using the instance's SystemCode as the unique Node ID.
func (SnowFlake) Start() error {
	// The SystemCode (derived from the runtime environment) serves as the 'Worker ID'.
	// This ensures that different service instances generate unique IDs even at the same millisecond.
	// TODO: Consider adding logic to handle cases where SystemCode exceeds the 10-bit Snowflake limit (0-1023).
	node, err := snowflake.NewNode(int64(runtime.GetAPP().Instance.SystemCode))
	if err != nil {
		return err
	}

	// Assign the initialized node to the package-level variable for global access.
	Node = node
	return nil
}

// Stop satisfies the bootstrap interface. No specific cleanup is required for the Snowflake node.
func (SnowFlake) Stop() {}
