package logger

import "log/slog"

// Level defines the severity of a log message.
// It is used to filter which messages are actually written to the output.
type Level int

// Log level constants defined in increasing order of severity.
const (
	// DEBUG is used for detailed information, typically of interest only when diagnosing problems.
	DEBUG Level = iota
	// INFO confirms that things are working as expected.
	INFO
	// WARN indicates that something unexpected happened, but the software is still working.
	WARN
	// ERROR indicates a more serious problem where the software has not been able to perform a function.
	ERROR
	// FATAL indicates a severe error that will lead the application to terminate.
	FATAL
)

// String returns the uppercase string representation of the log level.
// This is used for printing the level in text logs and for configuration lookups.
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
	// Note: FATAL usually maps to ERROR in many output formats unless handled specially.
	return ""
}

// GetLevel converts a string representation (e.g., "DEBUG") into the internal Level type.
// If the string is unrecognized, it defaults to DEBUG to ensure maximum visibility
// during initial setup/troubleshooting.
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

// getSlogLevel is an internal helper that maps asjard levels to the
// standard library's log/slog Level. This ensures compatibility with the
// Go ecosystem's structured logging performance and features.
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
