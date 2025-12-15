package xamqp

import (
	"context"
	"fmt"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/stores/xamqp"
	"github.com/streadway/amqp"
)

const (
	Protocol = "amqp"
)

// AmqpServer amqp服务定义
// 相当于就是amqp的消费者
type AmqpServer struct {
	conf    Config
	options *server.ServerOptions
	app     runtime.APP

	conn *xamqp.ClientConn
	ch   *amqp.Channel
	// 通道是否已关闭
	closed chan *amqp.Error

	svcs  []Handler
	tasks atomic.Int32
}

var (
	_ server.Server = &AmqpServer{}
)

func init() {
	server.AddServer(Protocol, New)
}

// New 服务初始化
func New(options *server.ServerOptions) (server.Server, error) {
	conf := defaultConfig()
	if err := config.GetWithUnmarshal("asjard.servers.amqp", &conf); err != nil {
		return nil, err
	}
	return MustNew(conf, options)
}

// MustNew 根据配置初始化
func MustNew(conf Config, options *server.ServerOptions) (server.Server, error) {
	if !conf.Enabled {
		return &AmqpServer{}, nil
	}
	conn, err := xamqp.Client(xamqp.WithClientName(conf.ClientName))
	if err != nil {
		return nil, err
	}
	return &AmqpServer{
		conn:    conn,
		conf:    conf,
		options: options,
		closed:  make(chan *amqp.Error),
	}, nil
}

func (s *AmqpServer) AddHandler(handler any) error {
	h, ok := handler.(Handler)
	if !ok {
		return fmt.Errorf("invalid handler, %T must implement *amqp.Handler", handler)
	}
	return s.addHandler(h)
}

func (s *AmqpServer) Start(startErr chan error) error {
	if err := s.start(); err != nil {
		return err
	}
	return s.keepalive()
}

func (s *AmqpServer) Stop() {
	if s.ch != nil {
		select {
		case <-s.closed:
		default:
			s.ch.Close()
		}
	}
	for {
		if s.tasks.Load() <= 0 {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *AmqpServer) Protocol() string { return Protocol }
func (s *AmqpServer) ListenAddresses() server.AddressConfig {
	return server.AddressConfig{}
}

func (s *AmqpServer) Enabled() bool {
	return s.conf.Enabled
}

func (s *AmqpServer) addHandler(handler Handler) error {
	desc := handler.AmqpServiceDesc()
	if desc == nil {
		return nil
	}
	ht := reflect.TypeOf(desc.HandlerType).Elem()
	st := reflect.TypeOf(handler)
	if !st.Implements(ht) {
		return fmt.Errorf("found the handler of type %v that does not satisfy %v", st, ht)
	}
	s.svcs = append(s.svcs, handler)
	return nil
}

func (s *AmqpServer) keepalive() error {
	go func() {
		for {
			select {
			case err, ok := <-s.closed:
				if !ok {
					return
				}
				if err != nil {
					logger.Error("channel exit, start reconnect", "err", err)
					s.reconnect()
				}
			}
		}
	}()
	return nil
}

func (s *AmqpServer) reconnect() {
	duration := time.Second
	for {
		if err := s.start(); err == nil {
			logger.Info("server reconnect to amqp success")
			return
		} else {
			logger.Error("server reconnect to amqp fail", "err", err)
		}
		time.Sleep(duration)
		duration += time.Second
		if duration >= time.Second*10 {
			duration = time.Second * 10
		}
	}
}

func (s *AmqpServer) start() error {
	ch, err := s.conn.Channel()
	if err != nil {
		return err
	}
	s.ch = ch
	ch.NotifyClose(s.closed)

	if err := ch.Qos(s.conf.PrefetchCount, s.conf.PrefetchSize, s.conf.Global); err != nil {
		return err
	}

	for _, svc := range s.svcs {
		desc := svc.AmqpServiceDesc()
		if desc == nil {
			continue
		}
		for _, method := range desc.Methods {
			if method.Handler == nil {
				continue
			}
			if _, err := ch.QueueDeclare(method.Queue, method.Durable, method.AutoDelete, method.Exclusive, method.NoWait, method.Table); err != nil {
				return err
			}
			if method.Exchange != "" {
				if err := ch.ExchangeDeclare(method.Exchange,
					method.Kind, method.Durable, method.AutoDelete, method.Internal, method.NoWait, method.Table); err != nil {
					return err
				}
				if err := ch.QueueBind(method.Queue, method.Route, method.Exchange, method.NoWait, method.Table); err != nil {
					return err
				}
			}
			// ch.ExchangeDeclare(name string, kind string, durable bool, autoDelete bool, internal bool, noWait bool, args amqp.Table)
			// ch.QueueBind(name string, key string, exchange string, noWait bool, args amqp.Table)
			msgs, err := ch.Consume(method.Queue, method.Consumer, method.AutoAck, method.Exclusive, method.NoLocal, method.NoWait, method.Table)
			if err != nil {
				return err
			}
			go s.run(msgs, svc, method)
		}
	}
	return nil
}

func (s *AmqpServer) run(msgs <-chan amqp.Delivery, svc Handler, method MethodDesc) {
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				logger.Debug("channel exit, exit goroutine", "queue", method.Queue)
				return
			}
			s.tasks.Add(1)
			if _, err := method.Handler(&Context{Context: context.Background(), task: msg}, svc, s.options.Interceptor); err == nil {
				msg.Ack(false)
			} else {
				msg.Reject(true)
			}
			s.tasks.Add(-1)
		}
	}
}
