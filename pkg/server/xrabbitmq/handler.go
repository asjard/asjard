package xrabbitmq

type Handler interface {
	RabbitmqServiceDesc() *ServiceDesc
}
