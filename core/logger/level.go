package logger

// Level 日志级别
type Level int

// 日志级别类型
const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String .
func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	}
	return ""
}
