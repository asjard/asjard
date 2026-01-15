package client

import (
	"testing"
	"time"

	"github.com/asjard/asjard/core/config"
	_ "github.com/asjard/asjard/pkg/config/mem"
	"github.com/asjard/asjard/utils"
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
		config.Set(key, protocolLB)
		time.Sleep(200 * time.Millisecond)

		conf := GetConfigWithProtocol(protocol)
		if conf.Loadbalance != protocolLB {
			t.Errorf("Expected protocol LB %s, got %s", protocolLB, conf.Loadbalance)
		}
	})

	t.Run("ServiceLevelOverride", func(t *testing.T) {
		// Simulate setting service-level config
		key := "asjard.clients." + protocol + "." + serviceName + ".loadbalance"
		config.Set(key, serviceLB)
		time.Sleep(200 * time.Millisecond)

		conf := serviceConfig(protocol, serviceName)
		if conf.Loadbalance != serviceLB {
			t.Errorf("Expected service LB %s, got %s", serviceLB, conf.Loadbalance)
		}
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
	found := false
	for _, s := range completed.Interceptors {
		if s == "custom-metric" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Custom interceptor lost during merge")
	}
}
