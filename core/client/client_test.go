package client

import (
	"testing"
)

// mockClient implements ClientInterface for testing URI generation and options.
type mockClient struct {
	lastTarget  string
	lastOptions *ClientOptions
}

// NewConn records the target and options passed by the orchestrator.
func (m *mockClient) NewConn(target string, options *ClientOptions) (ClientConnInterface, error) {
	m.lastTarget = target
	m.lastOptions = options
	return nil, nil // Return nil for connection as we only test orchestration
}

func TestClient_Conn(t *testing.T) {
	// 1. Setup Mock Registration
	protocol := "mock-proto"
	serviceName := "test-service"
	mock := &mockClient{}

	// Register the mock client factory
	AddClient(protocol, func(opts *ClientOptions) ClientInterface {
		return mock
	})

	// Manually inject the initialized mock into the active clients map
	// Normally this is handled by Init(), but we isolate the Conn() logic here.
	clients[protocol] = mock

	// 2. Define Test Cases
	tests := []struct {
		name       string
		ops        []ConnOption
		wantTarget string
	}{
		{
			name:       "Basic Connection",
			ops:        []ConnOption{},
			wantTarget: "asjard://mock-proto/test-service?instanceID=",
		},
		{
			name: "Connection With Instance ID",
			ops: []ConnOption{
				WithInstanceID("inst-123"),
			},
			wantTarget: "asjard://mock-proto/test-service?instanceID=inst-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(protocol, serviceName)

			// We expect an error because our mock returns nil for ClientConnInterface,
			// but we can still inspect the mock's recorded state.
			_, _ = c.Conn(tt.ops...)

			if mock.lastTarget != tt.wantTarget {
				t.Errorf("Target URI mismatch\ngot:  %s\nwant: %s", mock.lastTarget, tt.wantTarget)
			}

			if mock.lastOptions == nil {
				t.Error("ClientOptions were not passed to NewConn")
			}
		})
	}
}

func TestConnOptions_QueryString(t *testing.T) {
	opts := ConnOptions{
		InstanceID: "12345",
	}
	got := opts.queryString()
	want := "instanceID=12345"
	if got != want {
		t.Errorf("queryString() = %s, want %s", got, want)
	}
}
