package rest

// Response 请求返回
type Response struct {
	*Status
	// 请求数据
	Data any `json:"data"`
}
