package client

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/registry"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
)

// mockSubConn implements balancer.SubConn for testing purposes.
type mockSubConn struct {
	balancer.SubConn
}

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestLocalityRoundRobinPick(t *testing.T) {
	config.Set("asjard.service.instance.region", "region1")
	config.Set("asjard.service.instance.az", "az1")
	config.Set("asjard.service.instance.shareable", true)
	time.Sleep(500 * time.Millisecond)
	// 1. Setup Mock Data
	// Define different instances in various Regions and AZs
	instAZ1 := &registry.Instance{Service: &server.Service{APP: runtime.APP{Region: "region1", AZ: "az1", Instance: runtime.Instance{Shareable: true}}}}
	instAZ2 := &registry.Instance{Service: &server.Service{APP: runtime.APP{Region: "region1", AZ: "az2"}}}
	instOtherRegion := &registry.Instance{Service: &server.Service{APP: runtime.APP{Region: "region2", AZ: "az3"}}}

	sc1 := &mockSubConn{}
	sc2 := &mockSubConn{}
	sc3 := &mockSubConn{}

	// Map mock gRPC sub-connections to their metadata (Address + Attributes)
	scs := map[balancer.SubConn]base.SubConnInfo{
		sc1: {Address: resolver.Address{Attributes: attributes.New(AddressAttrKey{}, instAZ1).
			WithValue(ListenAddressKey{}, "127.0.0.1").
			WithValue(AdvertiseAddressKey{}, "example.com")}},
		sc2: {Address: resolver.Address{Attributes: attributes.New(AddressAttrKey{}, instAZ2).
			WithValue(ListenAddressKey{}, "127.0.0.1").
			WithValue(AdvertiseAddressKey{}, "example.com")}},
		sc3: {Address: resolver.Address{Attributes: attributes.New(AddressAttrKey{}, instOtherRegion).
			WithValue(ListenAddressKey{}, "127.0.0.1")}},
	}

	// 2. Initialize Picker
	// Note: In a real environment, runtime.GetAPP() would provide the local node's location.
	picker := NewLocalityRoundRobinPicker(scs).(*LocalityRoundRobinPicker)
	// Manually override local app locality for predictable testing
	picker.app = runtime.APP{Region: "region1", AZ: "az1"}

	tests := []struct {
		name           string
		md             metadata.MD
		expectedRegion string
		expectedAz     string
		desc           string
	}{
		{
			name: "Exact AZ Match",
			md: metadata.Pairs(
				HeaderRequestRegion, "region1",
				HeaderRequestAz, "az1",
			),
			expectedAz: "az1",
			desc:       "Should prioritize the exact AZ requested",
		},
		{
			name: "Region Match Fallback",
			md: metadata.Pairs(
				HeaderRequestRegion, "region1",
				HeaderRequestAz, "non-existent-az",
			),
			expectedRegion: "region1",
			expectedAz:     "az1",
			desc:           "Should fallback to same Region when requested AZ is missing",
		},
		{
			name:       "Local AZ Default",
			md:         metadata.MD{},
			expectedAz: "az1",
			desc:       "Should default to local AZ when no metadata is provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), tt.md)
			info := balancer.PickInfo{Ctx: ctx}

			result, err := picker.Pick(info)
			if err != nil {
				t.Fatalf("Pick failed: %v", err)
			}

			// Extract instance from the picked result
			pickedInst := result.SubConn.Address.Attributes.Value(AddressAttrKey{}).(*registry.Instance)

			if tt.expectedAz != "" && pickedInst.Service.AZ != tt.expectedAz {
				t.Errorf("%s: expected AZ %s, got %s", tt.name, tt.expectedAz, pickedInst.Service.AZ)
			}
			if tt.expectedRegion != "" && pickedInst.Service.Region != tt.expectedRegion {
				t.Errorf("%s: expected Region %s, got %s", tt.name, tt.expectedRegion, pickedInst.Service.Region)
			}
		})
	}
}

func TestLocalityRoundRobinEmpty(t *testing.T) {
	// Test behavior when no sub-connections are available
	picker := NewLocalityRoundRobinPicker(make(map[balancer.SubConn]base.SubConnInfo))
	_, err := picker.Pick(balancer.PickInfo{Ctx: context.Background()})
	if err != balancer.ErrNoSubConnAvailable {
		t.Errorf("Expected ErrNoSubConnAvailable, got %v", err)
	}
}
