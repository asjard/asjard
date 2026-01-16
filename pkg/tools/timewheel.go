package tools

import (
	"container/list"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/logger"
)

// Task represents a unit of work to be executed after a specific delay.
type Task struct {
	delay  time.Duration
	circle int    // Number of full wheel rotations remaining before execution.
	fn     func() // The actual business logic to execute (e.g., cache deletion).
}

// TimeWheel is a high-performance timer managed via a circular buffer (slots).
// It minimizes the overhead of managing thousands of individual timers by
// using a single goroutine and a ticker.
type TimeWheel struct {
	interval    time.Duration // The duration of each tick (precision).
	maxInterval time.Duration
	ticker      *time.Ticker // Drives the movement of the wheel.
	slots       []*list.List // A slice of linked lists, each representing a time bucket.

	currentPos int // The current index the wheel is processing.
	slotNum    int // Total capacity of the wheel (number of buckets).

	addTaskCh chan *Task    // Thread-safe channel for adding new timers.
	stopCh    chan struct{} // Signal to stop the background goroutine.

	// wg tracks tasks that are currently executing or waiting to execute
	wg sync.WaitGroup
	// running tracks if the wheel is still accepting new tasks
	runing atomic.Bool
	// record how many waiting tasks in slots
	taskCount atomic.Int32
}

// DefaultTW is the shared TimeWheel for the entire application.
// 100ms precision with 60 slots (6 seconds per rotation).
var DefaultTW = NewTimeWheel(100*time.Millisecond, 6*time.Second, 60)

func init() {
	bootstrap.AddBootstrap(DefaultTW)
}

// NewTimeWheel creates a new instance.
//
// Design Philosophy:
//  1. Guaranteed Execution: All tasks will execute at their scheduled time.
//     The Stop() method will block and wait for the physical delay to expire
//     rather than executing tasks prematurely during shutdown.
//  2. Use Case: Ideal for high-frequency, short-term delayed tasks
//     (e.g., Cache Delayed Double Deletion, request retries).
//  3. Limitation: Do not use this for tasks with extremely long wait times
//     (e.g., several hours or days), as it will cause the service shutdown
//     process to hang for an extended period.
//
// Parameters:
//   - interval: The precision of the wheel movement. e.g., 100ms means
//     task execution error is within 100ms.
//   - maxInterval: The maximum allowed delay for a task. Tasks exceeding
//     this value will be rejected by AddTask to prevent OOM or indefinite shutdown blocks.
//   - slotNum: The number of buckets in the wheel. Total Cycle = interval * slotNum.
func NewTimeWheel(interval, maxInterval time.Duration, slotNum int) *TimeWheel {
	// Default precision handling
	if interval == 0 {
		interval = 100 * time.Millisecond
	}

	// If no maxInterval is specified, default to the total duration of one full rotation
	if maxInterval == 0 {
		maxInterval = interval * time.Duration(slotNum)
	}

	tw := &TimeWheel{
		interval:    interval,
		maxInterval: maxInterval,
		slots:       make([]*list.List, slotNum),
		slotNum:     slotNum,
		// Using an unbuffered channel ensures AddTask only returns
		// once the task is accepted by the run goroutine.
		addTaskCh: make(chan *Task),
		stopCh:    make(chan struct{}),
	}

	// Initialize the doubly linked list for each slot
	for i := 0; i < slotNum; i++ {
		tw.slots[i] = list.New()
	}

	return tw
}

// Start launches the background worker that advances the wheel.
func (tw *TimeWheel) Start() error {
	tw.ticker = time.NewTicker(tw.interval)
	tw.runing.Store(true)
	go tw.run()
	return nil
}

// Stop sends a signal to terminate the background worker and stop the ticker.
func (tw *TimeWheel) Stop() {
	// stop accept new task
	tw.runing.Store(false)
	// wait all task be executed
	for tw.taskCount.Load() > 0 {
		time.Sleep(tw.interval)
	}
	close(tw.stopCh)
	// wait all tasks finish
	tw.wg.Wait()
}

// AddTask schedules a function for execution after the specified delay.
func (tw *TimeWheel) AddTask(delay time.Duration, fn func()) error {
	if delay > tw.maxInterval {
		return fmt.Errorf("delay duration too long, current: %s, max: %s", delay.String(), tw.maxInterval.String())
	}
	// Don't accept new tasks if we are stopping
	if !tw.runing.Load() {
		return fmt.Errorf("timewheel is stoping")
	}
	tw.taskCount.Add(1)
	if delay <= 0 {
		tw.safeExcute(fn)
		return nil
	}
	tw.addTaskCh <- &Task{delay: delay, fn: fn}
	return nil
}

func (tw *TimeWheel) run() {
	for {
		select {
		case <-tw.ticker.C:
			tw.tickHandler()
		case task := <-tw.addTaskCh:
			tw.addTask(task)
		case <-tw.stopCh:
			tw.ticker.Stop()
			return
		}
	}
}

// tickHandler moves the cursor and processes tasks in the current slot.
func (tw *TimeWheel) tickHandler() {
	l := tw.slots[tw.currentPos]
	tw.executeTasks(l)

	// Move cursor in a circular fashion.
	if tw.currentPos == tw.slotNum-1 {
		tw.currentPos = 0
	} else {
		tw.currentPos++
	}
}

// executeTasks iterates through the list in a slot.
// If a task's circle count is 0, it executes; otherwise, it decrements the circle.
func (tw *TimeWheel) executeTasks(l *list.List) {
	for e := l.Front(); e != nil; {
		task := e.Value.(*Task)
		if task.circle > 0 {
			task.circle--
			e = e.Next()
			continue
		}

		// Execute in a new goroutine so the wheel's main loop isn't blocked by slow tasks.
		tw.safeExcute(task.fn)

		next := e.Next()
		l.Remove(e)
		e = next
	}
}

func (tw *TimeWheel) safeExcute(fn func()) {
	tw.wg.Add(1)
	go func() {
		defer tw.wg.Done()
		defer tw.taskCount.Add(-1)
		defer func() {
			if r := recover(); r != nil {
				// Log the panic so it can be debugged
				// Assuming you are using your framework's logger
				logger.Error("TimeWheel task panic recovered",
					"err", r,
					"stack", string(debug.Stack()))
			}
		}()
		fn()
	}()
}

func (tw *TimeWheel) addTask(task *Task) {
	ticks := int(task.delay / tw.interval)
	// Circle logic allows for delays much longer than the wheel's total duration.
	task.circle = ticks / tw.slotNum
	pos := (tw.currentPos + ticks) % tw.slotNum

	tw.slots[pos].PushBack(task)
}
