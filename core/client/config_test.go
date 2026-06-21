package client

import (
	"slices"
	"testing"
	"time"

	"github.com/asjard/asjard/core/config"
	_ "github.com/asjard/asjard/pkg/config/mem"
	"github.com/asjard/asjard/utils"
	"github.com/stretchr/testify/require"
)

// TestConfigHierarchy verifies that service-level configuration correctly overrides
// protocol-level and default configurations.
func TestConfigHierarchy(t *testing.T) {
	// 1. Setup Mock Config Data
	protocol := "grpc"
	serviceName := "user-service"

	// Mock protocol-level load balance
	protocolLB := "roundRobin"
	// Mock service-level load balance (should override protocolLB)
	serviceLB := "random"

	// Mocking the config system behavior
	// Note: In a real test, you would use config.Mock or a similar mechanism
	// provided by your asjard/core/config package.

	t.Run("ProtocolLevelOverride", func(t *testing.T) {
		// Simulate setting protocol-level config
		key := "asjard.clients." + protocol + ".loadbalance"
		require.NoError(t, config.Set(key, protocolLB))
		require.Eventually(t, func() bool {
			return GetConfigWithProtocol(protocol).Loadbalance == protocolLB
		}, 3*time.Second, 20*time.Millisecond)
	})

	t.Run("ServiceLevelOverride", func(t *testing.T) {
		serviceProtocol := "grpc-service-test"
		protocolKey := "asjard.clients." + serviceProtocol + ".loadbalance"
		require.NoError(t, config.Set(protocolKey, protocolLB))
		require.Eventually(t, func() bool {
			return GetConfigWithProtocol(serviceProtocol).Loadbalance == protocolLB
		}, 3*time.Second, 20*time.Millisecond)

		// Simulate setting service-level config
		key := "asjard.clients." + serviceProtocol + "." + serviceName + ".loadbalance"
		require.NoError(t, config.Set(key, serviceLB))
		require.Eventually(t, func() bool {
			return serverConfig(serviceProtocol, serviceName).Loadbalance == serviceLB
		}, 3*time.Second, 20*time.Millisecond)
	})
}

// TestConfigComplete verifies that built-in interceptors and custom interceptors
// are merged correctly without losing the defaults.
func TestConfigComplete(t *testing.T) {
	conf := Config{
		BuiltInInterceptors: utils.JSONStrings{"auth", "log"},
		Interceptors:        utils.JSONStrings{"custom-metric"},
	}

	completed := conf.complete()

	// Check if both lists are merged
	expectedCount := 3
	if len(completed.Interceptors) != expectedCount {
		t.Errorf("Expected %d interceptors after merge, got %d", expectedCount, len(completed.Interceptors))
	}

	// Verify specific elements exist
	if !slices.Contains(completed.Interceptors, "custom-metric") {
		t.Error("Custom interceptor lost during merge")
	}
}
