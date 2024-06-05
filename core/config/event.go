package config

// EventType 事件类型
type EventType int

const (
	// EventTypeCreate 创建事件
	EventTypeCreate EventType = iota
	// EventTypeUpdate 更新事件
	EventTypeUpdate
	// EventTypeDelete 删除事件
	EventTypeDelete
)

// Event 配置事件
type Event struct {
	// 事件类型
	Type EventType
	// 事件源
	// Source string
	// 配置key
	Key string
	// 配置的值
	Value *Value
}

// String 返回字符串
func (e EventType) String() string {
	switch e {
	case EventTypeCreate:
		return "Create"
	case EventTypeUpdate:
		return "Update"
	case EventTypeDelete:
		return "Delete"
	default:
		return "Unknown"
	}
}
