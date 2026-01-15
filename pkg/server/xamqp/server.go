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
	// Protocol defines the server type for the framework registry.
	Protocol = "amqp"
)

// AmqpServer defines the AMQP consumer server.
// It manages the underlying connection and maps RabbitMQ messages to service handlers.
type AmqpServer struct {
	conf    Config
	options *server.ServerOptions
	app     runtime.APP

	conn   *xamqp.ClientConn // Wrapped AMQP connection
	ch     *amqp.Channel     // Active AMQP channel
	closed chan *amqp.Error  // Listener for channel closure events

	svcs  []Handler    // Registered business logic handlers
	tasks atomic.Int32 // Counter for active processing tasks (for graceful shutdown)
}

var (
	_ server.Server = &AmqpServer{}
)

func init() {
	// Register the AMQP server factory with the core framework.
	server.AddServer(Protocol, New)
}

// New initializes the server by loading configuration from "asjard.servers.amqp".
func New(options *server.ServerOptions) (server.Server, error) {
	conf := defaultConfig()
	if err := config.GetWithUnmarshal("asjard.servers.amqp", &conf); err != nil {
		return nil, err
	}
	return MustNew(conf, options)
}

// MustNew establishes the initial connection and prepares the server struct.
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

// AddHandler validates and registers a service handler to the server.
func (s *AmqpServer) AddHandler(handler any) error {
	h, ok := handler.(Handler)
	if !ok {
		return fmt.Errorf("invalid handler, %T must implement *amqp.Handler", handler)
	}
	return s.addHandler(h)
}

// Start opens the channel and starts the keepalive monitor.
func (s *AmqpServer) Start(startErr chan error) error {
	if err := s.start(); err != nil {
		return err
	}
	return s.keepalive()
}

// Stop closes the channel and waits for active tasks to complete (graceful shutdown).
func (s *AmqpServer) Stop() {
	if s.ch != nil {
		select {
		case <-s.closed:
		default:
			s.ch.Close()
		}
	}
	// Block until all in-flight messages are processed.
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

// addHandler stores the service and validates it against the generated descriptor.
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

// keepalive listens for unexpected channel closures and triggers reconnection.
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

// reconnect implements an exponential backoff strategy to restore the connection.
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

// start performs the heavy lifting: declaring queues, exchanges, and bindings.
func (s *AmqpServer) start() error {
	ch, err := s.conn.Channel()
	if err != nil {
		return err
	}
	s.ch = ch
	// Register a listener for channel closure.
	ch.NotifyClose(s.closed)

	// Apply Quality of Service (QoS) settings like PrefetchCount.
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
			// Idempotently declare the queue.
			queue, err := ch.QueueDeclare(method.Queue, method.Durable, method.AutoDelete, method.Exclusive, method.NoWait, method.Table)
			if err != nil {
				return err
			}
			// If an exchange is specified, declare it and bind the queue.
			if method.Exchange != "" {
				if err := ch.ExchangeDeclare(method.Exchange,
					method.Kind, method.Durable, method.AutoDelete, method.Internal, method.NoWait, method.Table); err != nil {
					return err
				}
				if err := ch.QueueBind(queue.Name, method.Route, method.Exchange, method.NoWait, method.Table); err != nil {
					return err
				}
			}
			// Start the actual consumption process.
			msgs, err := ch.Consume(queue.Name, method.Consumer, method.AutoAck, method.Exclusive, method.NoLocal, method.NoWait, method.Table)
			if err != nil {
				return err
			}
			go s.run(msgs, svc, method)
		}
	}
	return nil
}

// run processes the message delivery stream for a specific queue.
func (s *AmqpServer) run(msgs <-chan amqp.Delivery, svc Handler, method MethodDesc) {
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				logger.Debug("channel exit, exit goroutine", "queue", method.Queue)
				return
			}
			s.tasks.Add(1)
			// Execute business logic via the descriptor's handler.
			// Success results in an Ack, failure results in a Reject (with requeue).
			if _, err := method.Handler(&Context{Context: context.Background(), task: msg}, svc, s.options.Interceptor); err == nil {
				msg.Ack(false)
			} else {
				msg.Reject(true)
			}
			s.tasks.Add(-1)
		}
	}
}
