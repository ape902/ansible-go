package types

// PluginConfig 定义插件相关配置
type PluginConfig struct {
	// 插件目录
	Dir string `json:"dir" yaml:"dir" toml:"dir"`

	// 已启用的插件列表
	Enabled []string `json:"enabled" yaml:"enabled" toml:"enabled"`

	// 已禁用的插件列表
	Disabled []string `json:"disabled" yaml:"disabled" toml:"disabled"`

	// 插件加载顺序
	LoadOrder []string `json:"load_order" yaml:"load_order" toml:"load_order"`

	// 插件配置
	Configs map[string]map[string]interface{} `json:"configs" yaml:"configs" toml:"configs"`

	// 插件超时设置(秒)
	Timeout int `json:"timeout" yaml:"timeout" toml:"timeout"`

	// 自动发现插件
	AutoDiscover bool `json:"auto_discover" yaml:"auto_discover" toml:"auto_discover"`

	// 插件热重载
	HotReload bool `json:"hot_reload" yaml:"hot_reload" toml:"hot_reload"`
}

// PluginMetadata 定义插件元数据
type PluginMetadata struct {
	// 插件名称
	Name string `json:"name" yaml:"name" toml:"name"`

	// 插件版本
	Version string `json:"version" yaml:"version" toml:"version"`

	// 插件描述
	Description string `json:"description" yaml:"description" toml:"description"`

	// 插件作者
	Author string `json:"author" yaml:"author" toml:"author"`

	// 插件许可证
	License string `json:"license" yaml:"license" toml:"license"`

	// 插件依赖
	Dependencies []string `json:"dependencies" yaml:"dependencies" toml:"dependencies"`

	// 插件标签
	Tags []string `json:"tags" yaml:"tags" toml:"tags"`

	// 插件类型
	Type string `json:"type" yaml:"type" toml:"type"`

	// 插件入口点
	EntryPoint string `json:"entry_point" yaml:"entry_point" toml:"entry_point"`
}