package config

import (
	"regexp"
	"sync"

	"github.com/asjard/asjard/core/logger"
)

// Listener 配置变化监听着
type Listener struct {
	// 完全匹配回调
	callbacks map[string][]func(*Event)
	// 正则匹配回调
	matchCallbacks map[string][]func(*Event)
	// funcCallbacks  map[*WatchFunc][]func(*Event)
	cm     sync.RWMutex
	mm     sync.RWMutex
	watchs []*watch
}

type watch struct {
	f ListenFunc
	c ListenCallback
}

func newListener() *Listener {
	return &Listener{
		callbacks:      make(map[string][]func(*Event)),
		matchCallbacks: make(map[string][]func(*Event)),
	}
}

func (l *Listener) watch(key string, opt *watchOptions) {
	if opt == nil || opt.callback == nil {
		return
	}
	// 正则匹配
	if opt.pattern != "" {
		l.mm.Lock()
		l.matchCallbacks[opt.pattern] = append(l.matchCallbacks[opt.pattern], opt.callback)
		l.mm.Unlock()
		return
	}
	if key != "" {
		l.cm.Lock()
		l.callbacks[key] = append(l.callbacks[key], opt.callback)
		l.cm.Unlock()
		return
	}
}

// 移除监听器
func (l *Listener) remove(key string) {
	l.cm.Lock()
	delete(l.callbacks, key)
	l.cm.Unlock()

	l.mm.Lock()
	delete(l.matchCallbacks, key)
	l.mm.Unlock()
}

func (l *Listener) notify(event *Event) {
	l.keyNotify(event)
	l.matchNotify(event)
}

func (l *Listener) keyNotify(event *Event) {
	l.cm.RLock()
	callbacks, ok := l.callbacks[event.Key]
	l.cm.RUnlock()
	if ok {
		for _, callback := range callbacks {
			callback(event)
		}
	}
}

func (l *Listener) matchNotify(event *Event) {
	l.mm.RLock()
	for pattern, callbacks := range l.matchCallbacks {
		ok, err := regexp.MatchString(pattern, event.Key)
		if err != nil {
			logger.Error("regular expression fail[%s]",
				"key", event.Key,
				"pattern", pattern,
				"err", err.Error())
			continue
		}
		if ok {
			for _, callback := range callbacks {
				callback(event)
			}
		}
	}
	l.mm.RUnlock()
}
