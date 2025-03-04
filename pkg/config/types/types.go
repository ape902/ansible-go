package types

// ConfigFormat 定义配置文件格式类型
type ConfigFormat string

const (
	// ConfigFormatYAML YAML格式配置
	ConfigFormatYAML ConfigFormat = "yaml"
	// ConfigFormatJSON JSON格式配置
	ConfigFormatJSON ConfigFormat = "json"
	// ConfigFormatTOML TOML格式配置
	ConfigFormatTOML ConfigFormat = "toml"
)

// ConnectionType 定义连接类型
type ConnectionType string

const (
	// ConnectionTypeSSH SSH连接
	ConnectionTypeSSH ConnectionType = "ssh"
	// ConnectionTypeWinRM WinRM连接
	ConnectionTypeWinRM ConnectionType = "winrm"
	// ConnectionTypeLocal 本地连接
	ConnectionTypeLocal ConnectionType = "local"
	// ConnectionTypeDocker Docker连接
	ConnectionTypeDocker ConnectionType = "docker"
)

// LogLevel 定义日志级别
type LogLevel string

const (
	// LogLevelDebug 调试级别
	LogLevelDebug LogLevel = "debug"
	// LogLevelInfo 信息级别
	LogLevelInfo LogLevel = "info"
	// LogLevelWarn 警告级别
	LogLevelWarn LogLevel = "warn"
	// LogLevelError 错误级别
	LogLevelError LogLevel = "error"
	// LogLevelFatal 致命级别
	LogLevelFatal LogLevel = "fatal"
)

// LogFormat 定义日志格式
type LogFormat string

const (
	// LogFormatJSON JSON格式日志
	LogFormatJSON LogFormat = "json"
	// LogFormatText 文本格式日志
	LogFormatText LogFormat = "text"
)

// PluginType 定义插件类型
type PluginType string

const (
	// PluginTypeModule 模块插件
	PluginTypeModule PluginType = "module"
	// PluginTypeConnection 连接插件
	PluginTypeConnection PluginType = "connection"
	// PluginTypeCallback 回调插件
	PluginTypeCallback PluginType = "callback"
	// PluginTypeFilter 过滤插件
	PluginTypeFilter PluginType = "filter"
)

// BaseConfig 定义基础配置接口
type BaseConfig interface {
	// Validate 验证配置是否有效
	Validate() error
}