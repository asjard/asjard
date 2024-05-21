package logger

import (
	"fmt"
	"log"
	"sync"
)

// defaultLogger 默认日志
type defaultLogger struct {
	level Level
}

var bufferPool = sync.Pool{New: func() any { return new([]byte) }}

func getBuffer() *[]byte {
	p := bufferPool.Get().(*[]byte)
	*p = (*p)[:0]
	return p
}
func putBuffer(p *[]byte) {
	// Proper usage of a sync.Pool requires each entry to have approximately
	// the same memory cost. To obtain this property when the stored type
	// contains a variably-sized buffer, we add a hard limit on the maximum buffer
	// to place back in the pool.
	//
	// See https://go.dev/issue/23199
	if cap(*p) > 64<<10 {
		*p = nil
	}
	bufferPool.Put(p)
}

func (dl defaultLogger) info(level Level, v ...interface{}) {
	if dl.level <= level {
		log.SetPrefix("[" + level.String() + "] ")
		// log.Println(v...)
		buf := getBuffer()
		defer putBuffer(buf)
		fmt.Appendln(*buf, v...)
		log.Output(4, string(fmt.Appendln(*buf, v...)))
	}
}

func (dl defaultLogger) infof(level Level, format string, v ...interface{}) {
	if dl.level <= level {
		log.SetPrefix("[" + level.String() + "] ")
		// log.Printf(format, v...)
		log.Output(4, fmt.Sprintf(format, v...))
	}
}

// Info .
func (dl defaultLogger) Info(v ...interface{}) {
	dl.info(INFO, v...)
}

// Infof .
func (dl defaultLogger) Infof(format string, v ...interface{}) {
	dl.infof(INFO, format, v...)

}

// Debug .
func (dl defaultLogger) Debug(v ...interface{}) {
	dl.info(DEBUG, v...)
}

// Debugf .
func (dl defaultLogger) Debugf(format string, v ...interface{}) {
	dl.infof(DEBUG, format, v...)
}

// Warn .
func (dl defaultLogger) Warn(v ...interface{}) {
	dl.info(WARN, v...)
}

// Warnf .
func (dl defaultLogger) Warnf(format string, v ...interface{}) {
	dl.infof(WARN, format, v...)
}

// Error .
func (dl defaultLogger) Error(v ...interface{}) {
	dl.info(ERROR, v...)
}

// Errorf .
func (dl defaultLogger) Errorf(format string, v ...interface{}) {
	dl.infof(ERROR, format, v...)
}

// SetLevel 设置日志级别
func (dl *defaultLogger) SetLevel(level Level) {
	dl.level = level
}
