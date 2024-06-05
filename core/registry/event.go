package registry

// EventType 事件类型
type EventType int

const (
	// EventTypeCreate 创建事件
	EventTypeCreate EventType = 0
	// EventTypeUpdate 更新
	EventTypeUpdate EventType = 1
	// EventTypeDelete 删除
	EventTypeDelete EventType = 2
)

// Event 服务发现注册事件
type Event struct {
	// 事件类型
	Type EventType
	// 服务实例详情
	Instance *Instance
}
