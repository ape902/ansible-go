package config

import (
	"fmt"
	"io/ioutil"

	"github.com/ape902/ansible-go/pkg/config/types"
	"gopkg.in/yaml.v3"
)

// Config 定义配置结构
type Config struct {
	Inventory map[string][]string
	Vars      map[string]interface{}
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		return &Config{
			Inventory: make(map[string][]string),
			Vars:      make(map[string]interface{}),
		}, nil
	}

	// 读取配置文件
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 初始化默认值
	if cfg.Inventory == nil {
		cfg.Inventory = make(map[string][]string)
	}
	if cfg.Vars == nil {
		cfg.Vars = make(map[string]interface{})
	}

	return &cfg, nil
}

// LoadPlaybook 加载playbook文件
func LoadPlaybook(playbookPath string) (*types.TaskConfig, error) {
	// 读取playbook文件
	data, err := ioutil.ReadFile(playbookPath)
	if err != nil {
		return nil, fmt.Errorf("读取playbook文件失败: %w", err)
	}

	// 解析YAML
	var taskConfig types.TaskConfig
	err = yaml.Unmarshal(data, &taskConfig)
	if err != nil {
		return nil, fmt.Errorf("解析playbook文件失败: %w", err)
	}

	// 验证必要字段
	if taskConfig.Name == "" {
		return nil, fmt.Errorf("playbook必须包含name字段")
	}
	if len(taskConfig.Hosts) == 0 {
		return nil, fmt.Errorf("playbook必须包含hosts字段")
	}
	if len(taskConfig.Tasks) == 0 {
		return nil, fmt.Errorf("playbook必须包含至少一个任务")
	}

	return &taskConfig, nil
}