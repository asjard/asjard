package config

import (
	"regexp"
	"sync"

	"github.com/asjard/asjard/core/logger"
)

// Listener manages configuration change subscribers.
// It maintains a registry of callbacks and matches incoming events against them.
type Listener struct {
	// callbacks stores direct key-to-callback mappings (map[string][]CallbackFunc).
	callbacks sync.Map
	// matchCallbacks stores regex-pattern-to-callback mappings (map[string][]CallbackFunc).
	matchCallbacks sync.Map

	// Internal slice for tracking watch relationships (reserved for future use).
	watchs []*watch
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
		cfuncs, ok := l.matchCallbacks.Load(opt.pattern)
		if !ok {
			cfuncs = []CallbackFunc{}
		}
		cfuncs = append(cfuncs.([]CallbackFunc), opt.callback)
		l.matchCallbacks.Store(opt.pattern, cfuncs)
	}

	// Register as a direct key listener if a specific key is provided.
	if key != "" {
		cfuncs, ok := l.callbacks.Load(key)
		if !ok {
			cfuncs = []CallbackFunc{}
		}
		cfuncs = append(cfuncs.([]CallbackFunc), opt.callback)
		l.callbacks.Store(key, cfuncs)
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
	callbacks, ok := l.callbacks.Load(event.Key)
	if ok {
		for _, callback := range callbacks.([]CallbackFunc) {
			callback(event)
		}
	}
}

// matchNotify iterates through all registered regex patterns and executes
// callbacks for any pattern that matches the event key.
func (l *Listener) matchNotify(event *Event) {
	l.matchCallbacks.Range(func(key, value any) bool {
		// key here is the regex pattern string.
		if ok, err := regexp.MatchString(key.(string), event.Key); ok {
			for _, callback := range value.([]CallbackFunc) {
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
