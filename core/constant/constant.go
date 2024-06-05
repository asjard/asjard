package constant

const (
	// Framework 框架名称
	Framework = "asjard"
	// FrameworkVersion 框架版本号
	FrameworkVersion = "1.0.0"
	// DefaultDelimiter 默认分隔符
	DefaultDelimiter = ","
	// ConfigDelimiter 配置分隔符
	ConfigDelimiter = "."
	// ServerListenAddressName 服务监听地址
	// 代表内部地址
	ServerListenAddressName = "listen"
	// ServerAdvertiseAddressName 服务广播地址名称
	// 开放地址意味着可以在外部访问
	// 例如垮AZ，APP， Region访问可以通过此地址访问
	ServerAdvertiseAddressName = "advertise"

	// ServerProtocolKey 服务协议名称
	ServerProtocolKey = "protocol"
	// ServiceNameKey 服务名称
	ServiceNameKey = "serviceName"
	// ServiceAppKey 项目
	ServiceAppKey = "app"
	// ServiceEnvKey 环境
	ServiceEnvKey = "env"
	// ServiceRegionKey 区域
	ServiceRegionKey = "region"
	// ServiceAZKey 可用区
	ServiceAZKey = "az"
	// ServiceIDKey 服务ID
	ServiceIDKey = "serviceID"
	// ServiceVersionKey 版本
	ServiceVersionKey = "version"
	// DiscoverNameKey 服务发现者
	DiscoverNameKey = "discoverName"

	// DefaultCipherName 默认加解密名称
	DefaultCipherName = "default"
)
