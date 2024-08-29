package xmongo

// ClientManager 数据库连接维护
// type ClientManager struct {
// }

// ClientConn 数据库连接
// type ClientConn struct {
// 	name   string
// 	client *mongo.Client
// }

// Config 数据库配置
// type Config struct {
// 	Clients map[string]*ClientConnConfig `json:"clients"`
// 	Options Options                      `json:"options"`
// }

// type ClientConnConfig struct {
// 	URI     string  `json:"uri"`
// 	Options Options `json:"options"`
// }

// Options 数据库连接参数
// type Options struct{}

// ClientOptions 获取客户端参数
// type ClientOptions struct {
// 	clientName   string
// 	databaseName string
// }

// type Option func(options *ClientOptions)

// var (
// 	clientManager *ClientManager
// )

// // WithClientName 设置客户端名称
// func WithClientName(clientName string) Option {
// 	return func(options *ClientOptions) {
// 		options.clientName = clientName
// 	}
// }

// WithDatabase 设置数据库
// func WithDatabase(database string) Option {
// 	return func(options *ClientOptions) {
// 		options.databaseName = database
// 	}
// }

// // Client 获取mongo客户端
// func Client(ctx context.Context, opts ...Option) (*mongo.Client, error) {
// 	return nil, nil
// }

// func Database(ctx context.Context, opts ...Option) (*mongo.Client, error) {
// 	return nil, nil
// }

// func Collection(ctx context.Context, opts ...Option) (*mongo.Collection, error) {
// 	return nil, nil
// }

// func init() {
// 	dbManager := &ClientManager{}
// 	bootstrap.AddBootstrap(dbManager)
// }

// func (m *ClientManager) Bootstrap() error {
// 	return nil
// }

// func (m *ClientManager) Shutdown() {

// }

// 连接到所有mongo数据库
// func (m *ClientManager) connClients(clients map[string]*ClientConnConfig) error {
// 	for clientName, conf := range clients {
// 		if err := m.connClient(clientName, conf); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// 连接到数据库
// func (m *ClientManager) connClient(clientName string, conf *ClientConnConfig) error {
// 	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(conf.URI))
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
