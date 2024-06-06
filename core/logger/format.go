package logger

// Format 日志格式
type Format int

const (
	// Text text格式
	Text Format = iota
	// Json json格式
	Json
)

func (f Format) String() string {
	switch f {
	case Text:
		return "text"
	case Json:
		return "json"
	default:
		return "text"
	}
}

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
