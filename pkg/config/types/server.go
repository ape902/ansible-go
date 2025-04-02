package types

// ServerConfig 定义服务器相关配置
type ServerConfig struct {
	// 主机配置
	Hosts HostsConfig `json:"hosts" yaml:"hosts" toml:"hosts"`

	// SSH配置
	SSH SSHConfig `json:"ssh" yaml:"ssh" toml:"ssh"`

	// 日志配置
	Log LogConfig `json:"log" yaml:"log" toml:"log"`

	// 插件配置
	Plugin PluginConfig `json:"plugin" yaml:"plugin" toml:"plugin"`

	// 配置文件格式 (yaml, json, toml)
	Format string `json:"format" yaml:"format" toml:"format"`

	// 配置热重载
	HotReload bool `json:"hot_reload" yaml:"hot_reload" toml:"hot_reload"`

	// 配置文件路径
	ConfigPath string `json:"config_path" yaml:"config_path" toml:"config_path"`
}

// HostsConfig 定义主机配置
type HostsConfig struct {
	// 主机列表，键为主机组名，值为主机列表
	Inventory map[string][]HostInfo `json:"inventory" yaml:"inventory" toml:"inventory"`

	// 默认主机组
	DefaultGroup string `json:"default_group" yaml:"default_group" toml:"default_group"`

	// 主机变量，键为主机名，值为变量映射
	HostVars map[string]map[string]interface{} `json:"host_vars" yaml:"host_vars" toml:"host_vars"`

	// 组变量，键为组名，值为变量映射
	GroupVars map[string]map[string]interface{} `json:"group_vars" yaml:"group_vars" toml:"group_vars"`
}

// HostInfo 定义主机信息
type HostInfo struct {
	// 主机名或IP地址
	Host string `json:"host" yaml:"host" toml:"host"`

	// 端口号
	Port int `json:"port" yaml:"port" toml:"port"`

	// 主机别名
	Alias string `json:"alias" yaml:"alias" toml:"alias"`

	// 连接类型 (ssh, winrm, local, docker, etc.)
	ConnectionType string `json:"connection_type" yaml:"connection_type" toml:"connection_type"`

	// 主机特定变量
	Vars map[string]interface{} `json:"vars" yaml:"vars" toml:"vars"`
}

// SSHConfig 定义SSH连接配置
type SSHConfig struct {
	// 默认用户名
	User string `json:"user" yaml:"user" toml:"user"`

	// 默认密码
	Password string `json:"password" yaml:"password" toml:"password"`

	// 私钥文件路径
	KeyFile string `json:"key_file" yaml:"key_file" toml:"key_file"`

	// 私钥密码
	KeyPassword string `json:"key_password" yaml:"key_password" toml:"key_password"`

	// 默认端口
	Port int `json:"port" yaml:"port" toml:"port"`

	// 连接超时时间(秒)
	Timeout int `json:"timeout" yaml:"timeout" toml:"timeout"`

	// 是否使用密钥认证（当为true时使用密钥认证，为false时使用密码认证）
	// 如果Password为空且KeyFile不为空，则自动设置为true
	UseKeyAuth bool `json:"use_key_auth" yaml:"use_key_auth" toml:"use_key_auth"`

	// 是否禁用主机密钥检查
	DisableHostKeyChecking bool `json:"disable_host_key_checking" yaml:"disable_host_key_checking" toml:"disable_host_key_checking"`

	// 是否启用代理跳转
	UseJumpHost bool `json:"use_jump_host" yaml:"use_jump_host" toml:"use_jump_host"`

	// 跳转主机配置
	JumpHost JumpHostConfig `json:"jump_host" yaml:"jump_host" toml:"jump_host"`

	// 最大并行连接数
	MaxParallel int `json:"max_parallel" yaml:"max_parallel" toml:"max_parallel"`
}

// JumpHostConfig 定义SSH跳转主机配置
type JumpHostConfig struct {
	// 主机名或IP地址
	Host string `json:"host" yaml:"host" toml:"host"`

	// 端口号
	Port int `json:"port" yaml:"port" toml:"port"`

	// 用户名
	User string `json:"user" yaml:"user" toml:"user"`

	// 密码
	Password string `json:"password" yaml:"password" toml:"password"`

	// 私钥文件路径
	KeyFile string `json:"key_file" yaml:"key_file" toml:"key_file"`
}
