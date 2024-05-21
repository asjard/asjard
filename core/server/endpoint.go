package server

// Endpoint .
type Endpoint struct {
	// 协议
	Protocol string
	// 地址列表
	Addresses []*EndpointAddress
}

// EndpointAddress .
type EndpointAddress struct {
	Name    string
	Address string
}
