package logger

// Logger 日志
type Logger interface {
	// info级别日志
	Info(format string, kv ...any)
	// debug级别日志
	Debug(format string, kv ...any)
	// warn级别日志
	Warn(format string, kv ...any)
	// error级别日志
	Error(format string, kv ...any)
}

// L 日志组件
var L Logger

func init() {
	L = NewDefaultLogger(defaultLoggerConfig)
}

// SetLogger 添加日志组件
func SetLogger(lg Logger) {
	L = lg
}

// Info .
func Info(format string, kv ...any) {
	L.Info(format, kv...)
}

// Debug .
func Debug(format string, kv ...any) {
	L.Debug(format, kv...)
}

// Warn .
func Warn(format string, kv ...any) {
	L.Warn(format, kv...)
}

// Error .
func Error(format string, kv ...any) {
	L.Error(format, kv...)
}
