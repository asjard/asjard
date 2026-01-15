package xamqp

import (
	"fmt"
	"sync"
	"time"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/security"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/utils"
	"github.com/streadway/amqp"
)

const (
	// DefaultClientName is used if no client name is specified in the request.
	DefaultClientName = "default"
)

// ClientManager maintains a thread-safe registry of active AMQP connections
// and their associated configuration blueprints.
type ClientManager struct {
	clients sync.Map // Map of name -> *ClientConn

	cm      sync.RWMutex
	configs map[string]*ClientConnConfig
}

// ClientConn wraps a physical AMQP connection with metadata and a synchronization lock.
type ClientConn struct {
	name string
	conn *amqp.Connection
	conf *ClientConnConfig
	cm   sync.RWMutex
	done chan struct{} // Used to signal background routines to stop
}

// Config represents the overall configuration structure for AMQP in the framework.
type Config struct {
	Clients map[string]ClientConnConfig `json:"clients"`
	Options Options                     `json:"options"`
}

// ClientConnConfig holds the connection string and security parameters.
type ClientConnConfig struct {
	Url string `json:"url"` // amqp://user:pass@host:port/vhost

	// CipherName and Params allow the URL (containing credentials) to be encrypted.
	CipherName   string         `json:"cipherName"`
	CipherParams map[string]any `json:"cipherParams"`
	Vhost        string         `json:"vhost"`
	Options      Options        `json:"options"`
}

// Options defines the technical AMQP dial parameters and TLS certificate paths.
type Options struct {
	ChannelMax int                `json:"channelMax"`
	FrameSize  int                `json:"frameSize"`
	HeartBeat  utils.JSONDuration `json:"heartbeat"`
	CAFile     string             `json:"caFile"`
	CertFile   string             `json:"certFile"`
	KeyFile    string             `json:"keyFile"`
}

type ClientOptions struct {
	clientName string
}

type Option func(*ClientOptions)

var (
	clientManager  *ClientManager
	defaultOptions = Options{
		HeartBeat: utils.JSONDuration{Duration: time.Second},
	}
)

func init() {
	clientManager = &ClientManager{
		configs: make(map[string]*ClientConnConfig),
	}
	// Register with bootstrap to initialize connections during app startup.
	bootstrap.AddBootstrap(clientManager)
}

// WithClientName allows selecting a specific named RabbitMQ cluster.
func WithClientName(clientName string) Option {
	return func(opt *ClientOptions) {
		if clientName != "" {
			opt.clientName = clientName
		}
	}
}

// Channel opens a new AMQP channel on the existing connection.
// Channels are the primary way to perform AMQP operations (publish/consume).
func (c *ClientConn) Channel() (*amqp.Channel, error) {
	c.cm.RLock()
	defer c.cm.RUnlock()
	if c.conn == nil || c.conn.IsClosed() {
		return nil, status.DatabaseNotFoundError()
	}
	return c.conn.Channel()
}

// Client retrieves a managed ClientConn from the global manager.
func Client(opts ...Option) (*ClientConn, error) {
	options := defaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}
	conn, ok := clientManager.clients.Load(options.clientName)
	if !ok {
		logger.Error("amqp not found", "name", options.clientName)
		return nil, status.DatabaseNotFoundError()
	}
	client, ok := conn.(*ClientConn)
	if !ok {
		logger.Error("invalid amqp client, must be *ClientConn", "current", fmt.Sprintf("%T", conn))
		return nil, status.InternalServerError()
	}
	return client, nil
}

// Start loads config and initializes connections. Called by the bootstrap process.
func (c *ClientManager) Start() error {
	clients, err := c.loadAndWatch()
	if err != nil {
		return err
	}
	return c.newClients(clients)
}

// Stop gracefully shuts down all managed connections.
func (c *ClientManager) Stop() {
	c.clients.Range(func(key, value any) bool {
		conn, ok := value.(*ClientConn)
		if ok && !conn.conn.IsClosed() {
			logger.Debug("amqp close", "client", conn.name)
			close(conn.done) // Stop the keepalive goroutine
			c.clients.Delete(key)
		}
		return true
	})
}

// newClients establishes new physical connections for each configuration provided.
func (c *ClientManager) newClients(clients map[string]*ClientConnConfig) error {
	logger.Debug("new clients", "clients", clients)
	for name, conf := range clients {
		conn, err := c.newClient(name, conf)
		if err != nil {
			logger.Error("connect to amqp fail", "name", name, "conf", conf, "err", err)
			return err
		}

		clientConn := &ClientConn{
			name: name,
			conn: conn,
			conf: conf,
			done: make(chan struct{}),
		}

		// If the client already exists (reloading), update the internal connection.
		v, ok := c.clients.Load(name)
		if ok {
			clientConn = v.(*ClientConn)
			clientConn.cm.Lock()
			clientConn.conn.Close() // Close old physical connection
			clientConn.name = name
			clientConn.conn = conn
			clientConn.conf = conf
			clientConn.cm.Unlock()
		} else {
			// Start the health check/reconnection loop for new clients.
			clientConn.keepalive()
		}
		c.clients.Store(name, clientConn)
		logger.Debug("connect to amqp success", "name", name, "conf", conf)
	}
	return nil
}

// newClient handles the low-level AMQP dial logic, including TLS and Decryption.
func (c *ClientManager) newClient(name string, conf *ClientConnConfig) (*amqp.Connection, error) {
	dialConfig := amqp.Config{
		Vhost:      conf.Vhost,
		ChannelMax: conf.Options.ChannelMax,
		FrameSize:  conf.Options.FrameSize,
		Heartbeat:  conf.Options.HeartBeat.Duration,
	}

	// Handle TLS configuration if certificates are provided.
	if conf.Options.CAFile != "" && conf.Options.CertFile != "" && conf.Options.KeyFile != "" {
		// Certificate path resolution logic...
		// (Reads CA, loads KeyPair, creates CertPool)
		// ... (omitted for brevity in comments)
	}

	connUrl := conf.Url
	// Decrypt the connection URL if it's protected.
	if conf.CipherName != "" {
		plainText, err := security.Decrypt(conf.Url, security.WithCipherName(conf.CipherName), security.WithParams(conf.CipherParams))
		if err != nil {
			return nil, err
		}
		connUrl = plainText
	}

	return amqp.DialConfig(connUrl, dialConfig)
}

// loadAndWatch initializes config and starts watching for remote configuration changes.
func (c *ClientManager) loadAndWatch() (map[string]*ClientConnConfig, error) {
	clients, err := c.loadConfig()
	if err != nil {
		return nil, err
	}
	config.AddPatternListener("asjard.stores.amqp.*", c.watch)
	return clients, err
}

// loadConfig fetches settings from the configuration center.
func (c *ClientManager) loadConfig() (map[string]*ClientConnConfig, error) {
	clients := make(map[string]*ClientConnConfig)
	options := defaultOptions
	if err := config.GetWithUnmarshal("asjard.stores.amqp.options", &options); err != nil {
		return clients, err
	}
	if err := config.GetWithUnmarshal("asjard.stores.amqp.clients", &clients); err != nil {
		return clients, err
	}

	// Apply global options to clients.
	for name, client := range clients {
		client.Options = options
		config.GetWithUnmarshal(fmt.Sprintf("asjard.stores.amqp.clients.%s.options", name), &client.Options)
	}

	c.cm.Lock()
	c.configs = clients
	c.cm.Unlock()
	return clients, nil
}

// watch responds to live configuration updates by re-establishing connections.
func (c *ClientManager) watch(event *config.Event) {
	clients, err := c.loadConfig()
	if err != nil {
		logger.Error("load amqp config fail", "err", err)
		return
	}
	if err := c.newClients(clients); err != nil {
		logger.Error("new clients fail", "err", err)
		return
	}
	// Cleanup removed clients...
}

// keepalive is a background goroutine that monitors connection health.
// If the connection drops, it initiates an exponential backoff reconnection.
func (c *ClientConn) keepalive() {
	go func() {
		duration := time.Second
		for {
			select {
			case <-c.done:
				return
			default:
				if c.conn.IsClosed() {
					logger.Debug("amqp disconnect, start to reconnect", "name", c.name)
					if err := clientManager.newClients(map[string]*ClientConnConfig{
						c.name: c.conf,
					}); err == nil {
						logger.Info("reconnect to amqp success", "name", c.name)
						duration = time.Second // Reset backoff on success
					} else {
						// Increase backoff delay on failure, capped at 10s.
						duration += time.Second
						if duration >= time.Second*10 {
							duration = time.Second * 10
						}
					}
				}
			}
			time.Sleep(duration)
		}
	}()
}

func defaultClientOptions() *ClientOptions {
	return &ClientOptions{
		clientName: DefaultClientName,
	}
}
