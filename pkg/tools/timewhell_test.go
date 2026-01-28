package tools

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestTimeWheel_DelayedExecution(t *testing.T) {
	// Create a wheel with 100ms precision
	tw := NewTimeWheel(100*time.Millisecond, 2*time.Second, 20)
	tw.Start()

	var taskExecuted atomic.Bool
	delay := 500 * time.Millisecond
	start := time.Now()

	err := tw.AddTask(delay, func() {
		taskExecuted.Store(true)
	})
	if err != nil {
		t.Fatalf("failed to add task: %v", err)
	}

	// Stop will wait for the 500ms to pass
	tw.Stop()

	duration := time.Since(start)
	if !taskExecuted.Load() {
		t.Error("task was not executed after Stop()")
	}

	// Verify the delay was respected (within a reasonable buffer for CI/scheduler jitter)
	if duration < delay {
		t.Errorf("task executed too early: got %v, want at least %v", duration, delay)
	}
}

func TestTimeWheel_GracefulStopWaiting(t *testing.T) {
	tw := NewTimeWheel(50*time.Millisecond, 1*time.Second, 10)
	tw.Start()

	count := atomic.Int32{}
	taskCount := 5
	delay := 300 * time.Millisecond

	for i := 0; i < taskCount; i++ {
		_ = tw.AddTask(delay, func() {
			count.Add(1)
		})
	}

	start := time.Now()
	// Stop should block for ~300ms
	tw.Stop()
	elapsed := time.Since(start)

	if count.Load() != int32(taskCount) {
		t.Errorf("expected %d tasks, but only %d finished", taskCount, count.Load())
	}

	if elapsed < delay {
		t.Errorf("Stop() returned too early: %v, expected at least %v", elapsed, delay)
	}
}

func TestTimeWheel_MaxInterval(t *testing.T) {
	max := 500 * time.Millisecond
	tw := NewTimeWheel(100*time.Millisecond, max, 10)

	err := tw.AddTask(1*time.Second, func() {})
	if err == nil {
		t.Error("expected error for delay exceeding maxInterval, got nil")
	}
}

func TestTimeWheel_RejectAfterStop(t *testing.T) {
	tw := NewTimeWheel(100*time.Millisecond, 1*time.Second, 10)
	tw.Start()
	tw.Stop()

	err := tw.AddTask(100*time.Millisecond, func() {})
	if err == nil {
		t.Error("expected error adding task to stopped timewheel, got nil")
	}
}
