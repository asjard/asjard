package logger

import "log"

// Logger 日志
type Logger interface {
	Info(v ...any)
	Infof(format string, v ...any)

	Debug(v ...any)
	Debugf(format string, v ...any)

	Warn(v ...any)
	Warnf(format string, v ...any)

	Error(v ...any)
	Errorf(format string, v ...any)
}

// L 日志组件
var L Logger

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	L = &defaultLogger{}
}

// NewLoggerFunc 日志组件初始化
type NewLoggerFunc func() (Logger, error)

var newLoggerFunc NewLoggerFunc

// Init 日志初始化
func Init() error {
	L.Debug("Start Init logger")
	defer L.Debug("init logger Done")
	if newLoggerFunc != nil {
		lg, err := newLoggerFunc()
		if err != nil {
			return err
		}
		L = lg
	}
	return nil
}

// AddLogger 添加日志组件
func AddLogger(newFunc NewLoggerFunc) {
	newLoggerFunc = newFunc
}

// Info .
func Info(text string) {
	L.Info(text)
}

// Infof .
func Infof(format string, value ...any) {
	L.Infof(format, value...)
}

// Debug .
func Debug(text string) {
	L.Debug(text)
}

// Debugf .
func Debugf(format string, value ...any) {
	L.Debugf(format, value...)
}

// Warn .
func Warn(text string) {
	L.Warn(text)
}

// Warnf .
func Warnf(format string, value ...any) {
	L.Warnf(format, value...)
}

// Error .
func Error(text string) {
	L.Error(text)
}

// Errorf .
func Errorf(format string, value ...any) {
	L.Errorf(format, value...)
}
