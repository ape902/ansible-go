package types

// LogConfig 定义日志配置结构
type LogConfig struct {
	// 日志级别 (debug, info, warn, error, fatal)
	Level string `json:"level" yaml:"level" toml:"level"`

	// 日志输出路径
	Output string `json:"output" yaml:"output" toml:"output"`

	// 是否输出到控制台
	Console bool `json:"console" yaml:"console" toml:"console"`

	// 日志文件最大大小(MB)
	MaxSize int `json:"max_size" yaml:"max_size" toml:"max_size"`

	// 日志文件最大保留天数
	MaxAge int `json:"max_age" yaml:"max_age" toml:"max_age"`

	// 日志文件最大保留个数
	MaxBackups int `json:"max_backups" yaml:"max_backups" toml:"max_backups"`

	// 是否压缩日志文件
	Compress bool `json:"compress" yaml:"compress" toml:"compress"`

	// 是否记录时间戳
	Timestamp bool `json:"timestamp" yaml:"timestamp" toml:"timestamp"`

	// 是否记录调用文件和行号
	Caller bool `json:"caller" yaml:"caller" toml:"caller"`

	// 是否记录完整路径
	FullPath bool `json:"full_path" yaml:"full_path" toml:"full_path"`

	// 日志格式 (json, text)
	Format string `json:"format" yaml:"format" toml:"format"`
}