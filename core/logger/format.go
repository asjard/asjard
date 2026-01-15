package logger

// Format defines the output structure of the logs.
type Format int

const (
	// Text represents a human-readable, plain-text log format.
	// Usually contains timestamp, level, and message separated by spaces/tabs.
	Text Format = iota
	// Json represents a structured JSON log format.
	// Ideal for machine parsing and log analysis tools.
	Json
)

// String returns the lowercase string representation of the log format.
// This is used for mapping configuration strings to the internal Format type.
func (f Format) String() string {
	switch f {
	case Text:
		return "text"
	case Json:
		return "json"
	default:
		// Defaults to "text" if the format is undefined.
		return "text"
	}
}

// GetFormat converts a string input (e.g., from a config file) into a Format type.
// If the input is unrecognized, it defaults to Json, following a "production-first"
// security and observability posture.
func GetFormat(format string) Format {
	switch format {
	case Text.String():
		return Text
	case Json.String():
		return Json
	default:
		return Json
	}
}
