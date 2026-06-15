package utils

import (
	"github.com/asjard/asjard/pkg/protobuf/mqpb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

type MQOption struct {
	Exchange string
	Kind     string

	Route      string
	Consumer   string
	Table      map[string]any
	DataFormat string

	AutoAck      *bool
	Durable      *bool
	AutoDelete   *bool
	Exclusive    *bool
	NoLocal      *bool
	NoWait       *bool
	Internal     *bool
	Requeue      *bool
	IsDelayTask  *bool
	FixedRetry   *mqpb.FixedRetryPolicy
	BackoffRetry *mqpb.BackoffRetryPolicy
}

type isMQ_Retry interface {
	isMQ_Retry()
}

func ParseMethodMqOption(service *protogen.Service, h *mqpb.MQ) *MQOption {
	return parseMethodMqOption(h, parseServiceMqOption(service))
}

//gocyclo:ignore
func parseMethodMqOption(h *mqpb.MQ, serviceOption *MQOption) *MQOption {
	table := make(map[string]any, len(h.GetTable()))
	for k, v := range h.GetTable() {
		table[k] = v
	}
	option := &MQOption{
		Route:       h.GetRoute(),
		Consumer:    h.GetConsumer(),
		Table:       table,
		DataFormat:  h.DataFormat,
		AutoAck:     h.AutoAck,
		Durable:     h.Durable,
		AutoDelete:  h.AutoDelete,
		Exclusive:   h.Exclusive,
		NoLocal:     h.NoLocal,
		NoWait:      h.NoWait,
		Internal:    h.Internal,
		Requeue:     h.Requeue,
		IsDelayTask: h.IsDelayTask,
	}
	if option.AutoAck == nil {
		option.AutoAck = serviceOption.AutoAck
	}
	if option.Durable == nil {
		option.Durable = serviceOption.Durable
	}
	if option.AutoDelete == nil {
		option.AutoDelete = serviceOption.AutoDelete
	}
	if option.Exclusive == nil {
		option.Exclusive = serviceOption.Exclusive
	}
	if option.NoLocal == nil {
		option.NoLocal = serviceOption.NoLocal
	}
	if option.NoWait == nil {
		option.NoWait = serviceOption.NoWait
	}
	if option.Requeue == nil {
		option.Requeue = serviceOption.Requeue
	}

	if option.IsDelayTask == nil {
		option.IsDelayTask = serviceOption.IsDelayTask
	}

	switch h.GetRetry().(type) {
	case *mqpb.MQ_FixedRetry:
		option.FixedRetry = h.GetFixedRetry()
	case *mqpb.MQ_BackoffRetry:
		option.BackoffRetry = h.GetBackoffRetry()
	}
	if option.FixedRetry == nil {
		option.FixedRetry = serviceOption.FixedRetry
	}
	if option.BackoffRetry == nil {
		option.BackoffRetry = serviceOption.BackoffRetry
	}
	if option.FixedRetry != nil && len(option.FixedRetry.RetryDelays) == 0 {
		option.FixedRetry.RetryDelays = []int32{30000, 60000, 180000, 300000}
	}
	if option.BackoffRetry != nil && option.BackoffRetry.Multiplier == 0 {
		option.BackoffRetry = &mqpb.BackoffRetryPolicy{
			InitialDelayMs: 6000,
			Multiplier:     1,
			MaxRetries:     10,
		}
	}

	if option.Internal == nil {
		option.Internal = serviceOption.Internal
	}
	switch h.GetExchange().(type) {
	case *mqpb.MQ_Direct:
		option.Kind = "direct"
		option.Exchange = h.GetDirect()
	case *mqpb.MQ_Fanout:
		option.Kind = "fanout"
		option.Exchange = h.GetFanout()
	case *mqpb.MQ_Topic:
		option.Kind = "topic"
		option.Exchange = h.GetTopic()
	case *mqpb.MQ_Headers:
		option.Kind = "headers"
		option.Exchange = h.GetHeaders()
	default:
		option.Kind = "direct"
	}
	return option
}

func parseServiceMqOption(service *protogen.Service) *MQOption {
	option := &MQOption{}
	if serviceMqOption, ok := proto.GetExtension(service.Desc.Options(), mqpb.E_ServiceMq).(*mqpb.MQ); ok && serviceMqOption != nil {
		option.AutoAck = serviceMqOption.AutoAck
		option.Durable = serviceMqOption.Durable
		option.AutoDelete = serviceMqOption.AutoDelete
		option.Exclusive = serviceMqOption.Exclusive
		option.NoLocal = serviceMqOption.NoLocal
		option.NoWait = serviceMqOption.NoWait
		option.Internal = serviceMqOption.Internal
		option.Requeue = serviceMqOption.Requeue
		option.IsDelayTask = serviceMqOption.IsDelayTask
		switch serviceMqOption.GetRetry().(type) {
		case *mqpb.MQ_FixedRetry:
			option.FixedRetry = serviceMqOption.GetFixedRetry()
		case *mqpb.MQ_BackoffRetry:
			option.BackoffRetry = serviceMqOption.GetBackoffRetry()
		}
	}
	return option
}
