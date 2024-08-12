package initator

// Initator 配置初始化后，其他组件初始化初始化之前需要执行的方法
type Initator interface {
	// 启动
	Start() error
	Stop()
}

var (
	initatorHandlers []Initator
	initatorMap      = make(map[Initator]struct{})
)

func AddInitator(handler Initator) {
	if _, ok := initatorMap[handler]; !ok {
		initatorHandlers = append(initatorHandlers, handler)
		initatorMap[handler] = struct{}{}
	}
}

func Start() error {
	for _, handler := range initatorHandlers {
		if err := handler.Start(); err != nil {
			return err
		}
	}
	return nil
}

func Stop() {
	for _, handler := range initatorHandlers {
		handler.Stop()
	}
}
