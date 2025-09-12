package xamqp

type Handler interface {
	AmqpServiceDesc() *ServiceDesc
}
