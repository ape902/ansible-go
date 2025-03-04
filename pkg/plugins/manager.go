package plugins

import (
	"fmt"
	"log"
	"plugin"
	"sync"
)

// PluginType 定义插件类型
type PluginType string

// 插件类型常量
const (
	PluginTypeModule  PluginType = "module"  // 模块插件
	PluginTypeFilter  PluginType = "filter"  // 过滤器插件
	PluginTypeHandler PluginType = "handler" // 处理器插件
)

// Plugin 定义插件接口
type Plugin interface {
	GetName() string
	GetType() PluginType
	Init() error
	Close() error
}

// Manager 定义插件管理器
type Manager struct {
	mutex   sync.RWMutex
	plugins map[string]Plugin
	paths   []string
}

// NewManager 创建新的插件管理器
func NewManager(pluginPaths []string) *Manager {
	return &Manager{
		plugins: make(map[string]Plugin),
		paths:   pluginPaths,
	}
}

// LoadPlugin 加载插件
func (m *Manager) LoadPlugin(path string) (Plugin, error) {
	// 打开插件
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开插件失败: %w", err)
	}

	// 查找插件符号
	sym, err := p.Lookup("Plugin")
	if err != nil {
		return nil, fmt.Errorf("查找插件符号失败: %w", err)
	}

	// 类型断言
	plug, ok := sym.(Plugin)
	if !ok {
		return nil, fmt.Errorf("插件不实现Plugin接口")
	}

	// 初始化插件
	err = plug.Init()
	if err != nil {
		return nil, fmt.Errorf("初始化插件失败: %w", err)
	}

	// 注册插件
	m.mutex.Lock()
	defer m.mutex.Unlock()

	name := plug.GetName()
	if _, exists := m.plugins[name]; exists {
		return nil, fmt.Errorf("插件 %s 已存在", name)
	}

	m.plugins[name] = plug
	log.Printf("加载插件成功: %s (%s)", name, plug.GetType())

	return plug, nil
}

// GetPlugin 获取插件
func (m *Manager) GetPlugin(name string) (Plugin, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	plug, exists := m.plugins[name]
	return plug, exists
}

// GetPluginsByType 获取指定类型的所有插件
func (m *Manager) GetPluginsByType(pluginType PluginType) []Plugin {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make([]Plugin, 0)
	for _, plug := range m.plugins {
		if plug.GetType() == pluginType {
			result = append(result, plug)
		}
	}

	return result
}

// LoadAllPlugins 加载所有插件
func (m *Manager) LoadAllPlugins() error {
	for _, path := range m.paths {
		// TODO: 实现目录扫描和插件加载逻辑
		log.Printf("扫描插件目录: %s", path)
	}

	return nil
}

// CloseAllPlugins 关闭所有插件
func (m *Manager) CloseAllPlugins() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for name, plug := range m.plugins {
		err := plug.Close()
		if err != nil {
			log.Printf("关闭插件 %s 失败: %v", name, err)
		}
	}

	m.plugins = make(map[string]Plugin)
}