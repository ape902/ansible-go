package config

import (
	"fmt"
	"io/ioutil"

	"github.com/ape902/ansible-go/pkg/config/types"
	"gopkg.in/yaml.v3"
)

// Config 定义配置结构
type Config struct {
	Inventory map[string][]types.HostInfo `yaml:"inventory"`
	Vars      map[string]interface{}      `yaml:"vars"`
	SSH       types.SSHConfig             `yaml:"ssh"`
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		return &Config{
			Inventory: make(map[string][]types.HostInfo),
			Vars:      make(map[string]interface{}),
		}, nil
	}

	// 读取配置文件
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML
	var rawConfig struct {
		Inventory map[string]interface{}    `yaml:"inventory"`
		Vars      map[string]interface{}    `yaml:"vars"`
		SSH       types.SSHConfig           `yaml:"ssh"`
	}

	err = yaml.Unmarshal(data, &rawConfig)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 创建最终配置对象
	cfg := &Config{
		Inventory: make(map[string][]types.HostInfo),
		Vars:      rawConfig.Vars,
		SSH:       rawConfig.SSH,
	}

	// 处理inventory部分
	for groupName, hosts := range rawConfig.Inventory {
		switch hostList := hosts.(type) {
		case []interface{}: // 简化格式 ["192.168.1.1", "192.168.1.2"]
			hostInfos := make([]types.HostInfo, 0, len(hostList))
			for _, host := range hostList {
				if hostStr, ok := host.(string); ok {
					hostInfos = append(hostInfos, types.HostInfo{
						Host:           hostStr,
						Port:           22, // 默认SSH端口
						ConnectionType: "ssh", // 默认连接类型
						Vars:           make(map[string]interface{}),
					})
				}
			}
			cfg.Inventory[groupName] = hostInfos

		case []map[string]interface{}: // 详细格式 [{"host": "192.168.1.1", "port": 22, ...}]
			hostInfos := make([]types.HostInfo, 0, len(hostList))
			for _, hostMap := range hostList {
				hostInfo := types.HostInfo{
					Port:           22, // 默认值
					ConnectionType: "ssh", // 默认值
					Vars:           make(map[string]interface{}),
				}

				// 解析主机信息
				if host, ok := hostMap["host"].(string); ok {
					hostInfo.Host = host
				}

				// 解析端口
				if port, ok := hostMap["port"].(int); ok {
					hostInfo.Port = port
				}

				// 解析别名
				if alias, ok := hostMap["alias"].(string); ok {
					hostInfo.Alias = alias
				}

				// 解析连接类型
				if connType, ok := hostMap["connection_type"].(string); ok {
					hostInfo.ConnectionType = connType
				}

				// 解析变量
				if vars, ok := hostMap["vars"].(map[string]interface{}); ok {
					hostInfo.Vars = vars
				}

				hostInfos = append(hostInfos, hostInfo)
			}
			cfg.Inventory[groupName] = hostInfos

		case []types.HostInfo: // 已经是HostInfo类型，直接使用
			cfg.Inventory[groupName] = hostList

		default:
			// 尝试将单个map转换为主机列表
			if hostMap, ok := hosts.(map[string]interface{}); ok {
				hostInfo := types.HostInfo{
					Port:           22, // 默认值
					ConnectionType: "ssh", // 默认值
					Vars:           make(map[string]interface{}),
				}

				// 解析主机信息
				if host, ok := hostMap["host"].(string); ok {
					hostInfo.Host = host
				}

				// 解析端口
				if port, ok := hostMap["port"].(int); ok {
					hostInfo.Port = port
				}

				// 解析别名
				if alias, ok := hostMap["alias"].(string); ok {
					hostInfo.Alias = alias
				}

				// 解析连接类型
				if connType, ok := hostMap["connection_type"].(string); ok {
					hostInfo.ConnectionType = connType
				}

				// 解析变量
				if vars, ok := hostMap["vars"].(map[string]interface{}); ok {
					hostInfo.Vars = vars
				}

				cfg.Inventory[groupName] = []types.HostInfo{hostInfo}
			} else {
				return nil, fmt.Errorf("不支持的主机列表格式: %v", hosts)
			}
		}
	}

	return cfg, nil
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