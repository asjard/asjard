package server

// ServerHandler 服务处理方法
type ServerHandler interface{}

// BeforeRequestServerHandler 请求前服务处理方法
type BeforeRequestServerHandler interface {
	GetParam(key string) string
}

// AfterRequestServerHandler 请求后服务处理方法
type AfterRequestServerHandler interface{}
