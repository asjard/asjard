package config

import (
	"regexp"
	"sync"

	"github.com/asjard/asjard/core/logger"
)

// Listener 配置变化监听着
type Listener struct {
	// 完全匹配回调
	callbacks sync.Map
	// 正则匹配回调
	matchCallbacks sync.Map

	watchs []*watch
}

type watch struct {
	f ListenFunc
	c CallbackFunc
}

func newListener() *Listener {
	return &Listener{
		callbacks:      sync.Map{},
		matchCallbacks: sync.Map{},
	}
}

func (l *Listener) watch(key string, opt *watchOptions) {
	if opt == nil || opt.callback == nil {
		return
	}
	// 正则匹配
	if opt.pattern != "" {
		cfuncs, ok := l.matchCallbacks.Load(opt.pattern)
		if !ok {
			cfuncs = []CallbackFunc{}
		}
		cfuncs = append(cfuncs.([]CallbackFunc), opt.callback)
		l.matchCallbacks.Store(opt.pattern, cfuncs)
		// return
	}
	if key != "" {
		cfuncs, ok := l.callbacks.Load(key)
		if !ok {
			cfuncs = []CallbackFunc{}
		}
		cfuncs = append(cfuncs.([]CallbackFunc), opt.callback)
		l.callbacks.Store(key, cfuncs)
	}
}

// 移除监听器
func (l *Listener) remove(key string) {
	l.callbacks.Delete(key)
	l.matchCallbacks.Delete(key)
}

func (l *Listener) notify(event *Event) {
	l.keyNotify(event)
	l.matchNotify(event)
}

func (l *Listener) keyNotify(event *Event) {
	callbacks, ok := l.callbacks.Load(event.Key)
	if ok {
		for _, callback := range callbacks.([]CallbackFunc) {
			callback(event)
		}
	}
}

func (l *Listener) matchNotify(event *Event) {
	l.matchCallbacks.Range(func(key, value any) bool {
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
