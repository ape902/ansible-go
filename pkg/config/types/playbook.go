package types

// Task 定义任务
type Task struct {
	// 任务名称
	Name string `json:"name" yaml:"name" toml:"name"`

	// 模块名称
	Module string `json:"module" yaml:"module" toml:"module"`

	// 模块参数
	Args map[string]interface{} `json:"args" yaml:"args" toml:"args"`

	// 条件表达式
	When string `json:"when" yaml:"when" toml:"when"`

	// 重试次数
	Retries int `json:"retries" yaml:"retries" toml:"retries"`

	// 重试间隔(秒)
	RetryInterval int `json:"retry_interval" yaml:"retry_interval" toml:"retry_interval"`

	// 超时时间(秒)
	Timeout int `json:"timeout" yaml:"timeout" toml:"timeout"`

	// 忽略错误
	IgnoreErrors bool `json:"ignore_errors" yaml:"ignore_errors" toml:"ignore_errors"`

	// 注册变量
	Register string `json:"register" yaml:"register" toml:"register"`
}