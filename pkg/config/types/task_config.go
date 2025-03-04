package types

// TaskConfig 定义任务配置结构
type TaskConfig struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description,omitempty"`
	Hosts       []string               `yaml:"hosts"`
	Tasks       []map[string]TaskSpec  `yaml:"tasks"`
	Vars        map[string]interface{} `yaml:"vars,omitempty"`
	Handlers    []HandlerSpec          `yaml:"handlers,omitempty"`
}

// TaskSpec 定义具体任务规格
type TaskSpec struct {
	Name        string                 `yaml:"name,omitempty"`
	Module      string                 `yaml:"module"`
	Args        map[string]interface{} `yaml:"args,omitempty"`
	Vars        map[string]interface{} `yaml:"vars,omitempty"`
	When        string                 `yaml:"when,omitempty"`
	Notify      []string               `yaml:"notify,omitempty"`
	IgnoreError bool                   `yaml:"ignore_error,omitempty"`
	Retries     int                    `yaml:"retries,omitempty"`
	Delay       string                 `yaml:"delay,omitempty"`
}

// HandlerSpec 定义处理器规格
type HandlerSpec struct {
	Name   string                 `yaml:"name"`
	Module string                 `yaml:"module"`
	Args   map[string]interface{} `yaml:"args,omitempty"`
}
