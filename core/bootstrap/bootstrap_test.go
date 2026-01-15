package bootstrap

import (
	"errors"
	"testing"
)

// mockInitiator is a helper structure to track execution flow.
type mockInitiator struct {
	name     string
	startErr error
	startLog []string
	stopLog  []string
}

// Start records the execution order into the startLog slice.
func (m *mockInitiator) Start() error {
	if m.startErr != nil {
		return m.startErr
	}
	m.startLog = append(m.startLog, m.name+" started")
	return nil
}

// Stop records the execution order into the stopLog slice.
func (m *mockInitiator) Stop() {
	m.stopLog = append(m.stopLog, m.name+" stopped")
}

// reset clears the global handlers to ensure test isolation.
func reset() {
	bootstrapHandlers = nil
	bootstrapedMap = make(map[Initiator]struct{})
	initiatorHandlers = nil
	initiatorMap = make(map[Initiator]struct{})
}

// TestLifecycle verifies the full process of registration, forward execution, and reverse shutdown.
func TestLifecycle(t *testing.T) {
	reset()
	startLog := make([]string, 0)
	stopLog := make([]string, 0)

	i1 := &mockInitiator{name: "init1", startLog: startLog, stopLog: stopLog}
	i2 := &mockInitiator{name: "init2", startLog: startLog, stopLog: stopLog}
	b1 := &mockInitiator{name: "boot1", startLog: startLog, stopLog: stopLog}
	b2 := &mockInitiator{name: "boot2", startLog: startLog, stopLog: stopLog}

	// 1. Test Registration and Idempotency
	AddInitiators(i1, i2)
	AddBootstraps(b1, b2)
	// Duplicate registration should be ignored by the internal map
	AddBootstrap(b1)

	if len(initiatorHandlers) != 2 || len(bootstrapHandlers) != 2 {
		t.Errorf("Registration failed: expected 2/2, got %d/%d", len(initiatorHandlers), len(bootstrapHandlers))
	}

	// 2. Test Initiator Sequence (Sequential)
	if err := Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if startLog[0] != "init1 started" || startLog[1] != "init2 started" {
		t.Errorf("Initiator execution order mismatch: %v", startLog)
	}

	// 3. Test Bootstrap Sequence (Sequential)
	if err := Bootstrap(); err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}
	if startLog[2] != "boot1 started" || startLog[3] != "boot2 started" {
		t.Errorf("Bootstrap execution order mismatch: %v", startLog)
	}

	// 4. Test Shutdown Sequence (LIFO - Last In First Out)
	// Expected Order: boot2 -> boot1 -> init2 -> init1
	Shutdown()
	expectedStop := []string{"boot2 stopped", "boot1 stopped", "init2 stopped", "init1 stopped"}
	for i, v := range expectedStop {
		if stopLog[i] != v {
			t.Errorf("Shutdown LIFO mismatch at index %d: expected %s, got %s", i, v, stopLog[i])
		}
	}
}

// TestInitError ensures that the startup process halts immediately if a component fails.
func TestInitError(t *testing.T) {
	reset()
	mockErr := errors.New("critical failure")
	errMock := &mockInitiator{
		name:     "faulty_node",
		startErr: mockErr,
	}

	AddInitiator(errMock)
	err := Init()
	if err == nil {
		t.Error("Expected Init() to return an error, but got nil")
	}
	if !errors.Is(err, mockErr) {
		t.Errorf("Expected error %v, got %v", mockErr, err)
	}
}
