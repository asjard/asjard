package xamqp

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/stores/xamqp"
	amqp "github.com/rabbitmq/amqp091-go"
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
	return &AmqpServer{
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
	conn, err := xamqp.Client(xamqp.WithClientName(s.conf.ClientName))
	if err != nil {
		return err
	}
	s.conn = conn
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
	r := ch.NotifyReturn(make(chan amqp.Return))
	go s.notifyReturn(r)

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
			if s.declareAndRun(svc, method); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *AmqpServer) declareAndRun(svc Handler, method MethodDesc) error {
	// Idempotently declare the queue.
	queue, err := s.ch.QueueDeclare(method.Queue, method.Durable, method.AutoDelete, method.Exclusive, method.NoWait, method.Table)
	if err != nil {
		return err
	}
	// If an exchange is specified, declare it and bind the queue.
	if method.Exchange != "" {
		if err := s.ch.ExchangeDeclare(method.Exchange,
			method.Kind, method.Durable, method.AutoDelete, method.Internal, method.NoWait, method.Table); err != nil {
			return err
		}
		if err := s.ch.QueueBind(queue.Name, method.Route, method.Exchange, method.NoWait, method.Table); err != nil {
			return err
		}
	}
	if method.RetryExchange != "" {
		retryTable := amqp.Table{}
		for k, v := range method.Table {
			retryTable[k] = v
		}
		retryTable["x-delayed-type"] = "direct"
		if err := s.ch.ExchangeDeclare(method.RetryExchange,
			"x-delayed-message", method.Durable, method.AutoDelete, method.Internal, method.NoWait, retryTable); err != nil {
			return err
		}
		if err := s.ch.QueueBind(queue.Name, method.RetryRoute, method.RetryExchange, method.NoWait, method.Table); err != nil {
			return err
		}
	}
	// Start the actual consumption process.
	msgs, err := s.ch.Consume(queue.Name, method.Consumer, method.AutoAck, method.Exclusive, method.NoLocal, method.NoWait, method.Table)
	if err != nil {
		return err
	}
	go s.run(msgs, svc, method)
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
			// Success results in an Ack, failure results in an Nack
			if _, err := method.Handler(&Context{Context: context.Background(), task: msg}, svc, s.options.Interceptor); err == nil {
				msg.Ack(false)
			} else {
				s.retry(msg, method)
			}
			s.tasks.Add(-1)
		}
	}
}

func (s *AmqpServer) notifyReturn(r chan amqp.Return) {
	for {
		select {
		case msg, ok := <-r:
			if !ok {
				return
			}
			logger.Error("delivery msg returned",
				"return", msg)
		}
	}
}

const (
	retryHeaderKey = "x-retry-counts"
)

func (s *AmqpServer) retry(msg amqp.Delivery, method MethodDesc) {
	if method.FixedRetry != nil {
		s.fixedRetry(msg, method)
		return
	}

	if method.BackoffRetry != nil {
		s.backoffRetry(msg, method)
		return
	}

	msg.Nack(false, method.ReQueue)
}

func (s *AmqpServer) fixedRetry(msg amqp.Delivery, method MethodDesc) {
	maxRetries := len(method.FixedRetry.RetryDelays)

	count, ok := msg.Headers[retryHeaderKey].(int32)
	if !ok {
		count = 0
	}

	if count >= int32(maxRetries) {
		msg.Nack(false, false)
		return
	}
	s.retryPublish(msg, method, count, method.FixedRetry.RetryDelays[count])
}

func (s *AmqpServer) backoffRetry(msg amqp.Delivery, method MethodDesc) {

	count, ok := msg.Headers[retryHeaderKey].(int32)
	if !ok {
		count = 0
	}

	if method.BackoffRetry.MaxRetries > 0 && count >= method.BackoffRetry.MaxRetries {
		msg.Nack(false, false)
		return
	}

	// initial * (multiplier ^ attempt)
	delay := method.BackoffRetry.InitialDelayMs * int32(math.Pow(float64(method.BackoffRetry.Multiplier), float64(count)))
	if delay < method.BackoffRetry.InitialDelayMs {
		delay = method.BackoffRetry.InitialDelayMs
	}

	if method.BackoffRetry.MaxDelayMs > 0 && delay > method.BackoffRetry.MaxDelayMs {
		delay = method.BackoffRetry.MaxDelayMs
	}

	s.retryPublish(msg, method, count, delay)
}

func (s *AmqpServer) retryPublish(msg amqp.Delivery, method MethodDesc, count, delay int32) {
	if err := s.ch.Publish(method.RetryExchange, method.RetryRoute, false, false, amqp.Publishing{
		Headers: amqp.Table{
			"x-delay":      int64(delay),
			retryHeaderKey: count + 1,
		},
		Body:         msg.Body,
		ContentType:  method.ContentType,
		DeliveryMode: amqp.Persistent,
	}); err != nil {
		logger.Error("republish msg to retry queue fail",
			"retry_exchange", method.RetryExchange, "retry_queue", method.RetryQueue, "err", err)
		msg.Nack(false, true)
		return
	}
	msg.Ack(false)
}
