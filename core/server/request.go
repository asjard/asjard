package server

// Requester 请求实现
type Requester interface {
	// 获取请求头
	GetHeader(key string) string
	// 设置请求头
	SetHeader(key, value string)
	// 获取参数
	GetParam(name string) any
}

// Request 客户端发出的请求
type Request struct {
	// 请求地址
	Host string
	// 请求协议
	Proto  string
	Header map[string][]string
}
