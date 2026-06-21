package config

import (
	"regexp"
	"sync"

	"github.com/asjard/asjard/core/logger"
)

// Listener manages configuration change subscribers.
// It maintains a registry of callbacks and matches incoming events against them.
type Listener struct {
	// callbacks stores direct key-to-callback groups.
	callbacks sync.Map
	// matchCallbacks stores regex-pattern-to-callback groups.
	matchCallbacks sync.Map

	// Internal slice for tracking watch relationships (reserved for future use).
	watchs []*watch
}

type callbackGroup struct {
	mu        sync.RWMutex
	callbacks []CallbackFunc
}

func (g *callbackGroup) add(callback CallbackFunc) {
	g.mu.Lock()
	g.callbacks = append(g.callbacks, callback)
	g.mu.Unlock()
}

func (g *callbackGroup) snapshot() []CallbackFunc {
	g.mu.RLock()
	callbacks := append([]CallbackFunc(nil), g.callbacks...)
	g.mu.RUnlock()
	return callbacks
}

// watch represents the relationship between a listening function and its callback.
type watch struct {
	f ListenFunc
	c CallbackFunc
}

// newListener initializes a new Listener instance with empty concurrency-safe maps.
func newListener() *Listener {
	return &Listener{
		callbacks:      sync.Map{},
		matchCallbacks: sync.Map{},
	}
}

// watch registers a new listener based on the provided options.
// It can register a listener for a specific key, a regex pattern, or both.
func (l *Listener) watch(key string, opt *watchOptions) {
	if opt == nil || opt.callback == nil {
		return
	}

	// Register as a pattern-based listener if a regex pattern is provided.
	if opt.pattern != "" {
		group, _ := l.matchCallbacks.LoadOrStore(opt.pattern, &callbackGroup{})
		group.(*callbackGroup).add(opt.callback)
	}

	// Register as a direct key listener if a specific key is provided.
	if key != "" {
		group, _ := l.callbacks.LoadOrStore(key, &callbackGroup{})
		group.(*callbackGroup).add(opt.callback)
	}
}

// remove unregisters all callbacks associated with a specific key or pattern string.
func (l *Listener) remove(key string) {
	l.callbacks.Delete(key)
	l.matchCallbacks.Delete(key)
}

// notify distributes a configuration event to all relevant subscribers.
func (l *Listener) notify(event *Event) {
	// Execute direct key notifications first.
	l.keyNotify(event)
	// Execute regex pattern notifications.
	l.matchNotify(event)
}

// keyNotify finds and executes callbacks registered for the exact key found in the event.
func (l *Listener) keyNotify(event *Event) {
	group, found := l.callbacks.Load(event.Key)
	if found {
		for _, callback := range group.(*callbackGroup).snapshot() {
			callback(event)
		}
	}
}

// matchNotify iterates through all registered regex patterns and executes
// callbacks for any pattern that matches the event key.
func (l *Listener) matchNotify(event *Event) {
	l.matchCallbacks.Range(func(key, value any) bool {
		// key here is the regex pattern string.
		if matched, err := regexp.MatchString(key.(string), event.Key); matched {
			for _, callback := range value.(*callbackGroup).snapshot() {
				callback(event)
			}
		} else if err != nil {
			logger.Error("regular expression fail",
				"key", event.Key,
				"pattern", key,
				"err", err)
		}
		return true
	})
}
