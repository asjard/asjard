package server

type Endpoint struct {
	// 监听地址
	Listen []string
	// 广播地址
	Advertise []string
}
