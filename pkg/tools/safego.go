package tools

import (
	"runtime/debug"

	"github.com/asjard/asjard/core/logger"
)

// SafeGo auto add panic recover on a goroutine
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recover", "err", r, "stack", string(debug.Stack()))
			}
		}()
		fn()
	}()
}
