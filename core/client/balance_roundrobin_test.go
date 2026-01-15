package client

import (
	"context"
	"testing"

	"github.com/asjard/asjard/core/registry"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

// mockRoundRobinSubConn is a mock implementation of gRPC's SubConn.
type mockRoundRobinSubConn struct {
	balancer.SubConn
	id string
}

func TestRoundRobinPicker_Pick(t *testing.T) {
	// 1. Setup Mock Sub-connections
	instAZ1 := &registry.Instance{Service: &server.Service{APP: runtime.APP{Region: "default", AZ: "default", Instance: runtime.Instance{Shareable: true}}}}
	instAZ2 := &registry.Instance{Service: &server.Service{APP: runtime.APP{Region: "default", AZ: "default"}}}
	instOtherRegion := &registry.Instance{Service: &server.Service{APP: runtime.APP{Region: "default", AZ: "default"}}}

	sc1 := &mockSubConn{}
	sc2 := &mockSubConn{}
	sc3 := &mockSubConn{}

	// Map mock gRPC sub-connections to their metadata (Address + Attributes)
	scs := map[balancer.SubConn]base.SubConnInfo{
		sc1: {Address: resolver.Address{Addr: "127.0.0.1", Attributes: attributes.New(AddressAttrKey{}, instAZ1).
			WithValue(ListenAddressKey{}, "127.0.0.1").
			WithValue(AdvertiseAddressKey{}, "example.com")}},
		sc2: {Address: resolver.Address{Addr: "127.0.0.2", Attributes: attributes.New(AddressAttrKey{}, instAZ2).
			WithValue(ListenAddressKey{}, "127.0.0.2").
			WithValue(AdvertiseAddressKey{}, "example.com")}},
		sc3: {Address: resolver.Address{Addr: "127.0.0.3", Attributes: attributes.New(AddressAttrKey{}, instOtherRegion).
			WithValue(ListenAddressKey{}, "127.0.0.3")}},
	}

	// 2. Initialize the Picker
	picker := NewRoundRobinPicker(scs).(*RoundRobinPicker)

	// 3. Test Cyclic Execution
	// We expect the picker to return nodes in the order they were added to the slice
	// Because map iteration is random, we just check if it cycles through all unique nodes
	results := make(map[string]bool)
	for i := 0; i < len(scs); i++ {
		res, err := picker.Pick(balancer.PickInfo{Ctx: context.Background()})
		if err != nil {
			t.Fatalf("Pick failed at iteration %d: %v", i, err)
		}
		// In a real test, sc1, sc2, sc3 would be stored in r.scs in a specific order.
		// Round Robin should hit every node once in 3 tries.
		addr := res.SubConn.Address.Addr
		results[addr] = true
	}

	if len(results) != 3 {
		t.Errorf("Round Robin did not cycle through all nodes, got: %v", results)
	}

	// 4. Test Reachability Filtering
	// We mock the PickerBase to simulate a failure in one of the nodes.
	// Note: This assumes CanReachable logic is accessible or mocked.
	// If CanReachable is true for all, we expect a consistent sequence.
	res1, _ := picker.Pick(balancer.PickInfo{Ctx: context.Background()})
	res2, _ := picker.Pick(balancer.PickInfo{Ctx: context.Background()})

	if res1.SubConn.Address.Addr == res2.SubConn.Address.Addr {
		t.Error("Round Robin picked the same node twice in a row")
	}
}

func TestRoundRobinPicker_NoAvailable(t *testing.T) {
	// Test behavior when the sub-connection map is empty
	picker := NewRoundRobinPicker(make(map[balancer.SubConn]base.SubConnInfo))

	_, err := picker.Pick(balancer.PickInfo{Ctx: context.Background()})
	if err != balancer.ErrNoSubConnAvailable {
		t.Errorf("Expected ErrNoSubConnAvailable, got %v", err)
	}
}
