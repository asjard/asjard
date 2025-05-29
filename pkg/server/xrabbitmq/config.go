package xrabbitmq

import "github.com/asjard/asjard/core/server"

type Config struct {
	server.Config

	ClientName string `json:"clientName"`
	//当前消费者一次能接受的最大消息数量
	PrefetchCount int `json:"prefetchCount"`
	//服务器传递的最大容量
	PrefetchSize int `json:"prefetchSize"`
	//如果为true 对channel可用 false则只对当前队列可用
	Global bool `json:"global"`
}

func defaultConfig() Config {
	return Config{
		PrefetchCount: 1,
	}
}
