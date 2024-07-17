package server

// Endpoint .
// type Endpoint struct {
// 	// 协议
// 	Protocol string
// 	// 地址列表
// 	Addresses []*EndpointAddress
// }

// Endpoint .
// type Endpoint struct {
// 	Name    string
// 	Address string
// }

type Endpoint struct {
	Listen    []string
	Advertise []string
}
