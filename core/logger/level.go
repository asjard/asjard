package logger

import "log/slog"

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

// GetLevel 根据字符串获取日志级别
func GetLevel(level string) Level {
	switch level {
	case DEBUG.String():
		return DEBUG
	case INFO.String():
		return INFO
	case WARN.String():
		return WARN
	case ERROR.String():
		return ERROR
	default:
		return DEBUG
	}
}

func getSlogLevel(level string) slog.Level {
	switch level {
	case DEBUG.String():
		return slog.LevelDebug
	case INFO.String():
		return slog.LevelInfo
	case WARN.String():
		return slog.LevelWarn
	case ERROR.String():
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}
