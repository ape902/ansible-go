package vars

import (
	"fmt"
	"os"
	"sort"
	"sync"
)

// Manager 定义变量管理器
type Manager struct {
	mutex  sync.RWMutex
	scopes []*Scope
	store  *Store
}

// NewManager 创建新的变量管理器
func NewManager() *Manager {
	return &Manager{
		scopes: make([]*Scope, 0),
		store:  NewStore(),
	}
}

// AddScope 添加变量作用域
func (m *Manager) AddScope(scope *Scope) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 按优先级排序插入
	pos := 0
	for i, s := range m.scopes {
		if scope.Priority > s.Priority {
			pos = i
			break
		}
		pos = i + 1
	}

	// 插入到指定位置
	if pos == len(m.scopes) {
		m.scopes = append(m.scopes, scope)
	} else {
		m.scopes = append(m.scopes[:pos+1], m.scopes[pos:]...)
		m.scopes[pos] = scope
	}
}

// RemoveScope 移除变量作用域
func (m *Manager) RemoveScope(scopeType ScopeType, name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for i, scope := range m.scopes {
		if scope.Type == scopeType && scope.Name == name {
			m.scopes = append(m.scopes[:i], m.scopes[i+1:]...)
			break
		}
	}
}

// GetScope 获取指定作用域
func (m *Manager) GetScope(scopeType ScopeType, name string) *Scope {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, scope := range m.scopes {
		if scope.Type == scopeType && scope.Name == name {
			return scope
		}
	}

	return nil
}

// Get 获取变量值，按优先级从高到低查找
func (m *Manager) Get(key string) (interface{}, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 按优先级从高到低查找
	for _, scope := range m.scopes {
		if val, exists := scope.Get(key); exists {
			return val, true
		}
	}

	// 最后查找环境变量
	if scopeType := m.GetScope(ScopeEnv, "env"); scopeType == nil {
		// 如果环境变量作用域不存在，则创建
		m.mutex.RUnlock()
		m.initEnvScope()
		m.mutex.RLock()
	}

	// 再次尝试从环境变量获取
	if val, exists := os.LookupEnv(key); exists {
		return val, true
	}

	return nil, false
}

// Set 设置变量值到指定作用域
func (m *Manager) Set(scopeType ScopeType, name string, key string, value interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	scope := m.GetScope(scopeType, name)
	if scope == nil {
		return fmt.Errorf("作用域不存在: %v - %s", scopeType, name)
	}

	scope.Set(key, value)
	return nil
}

// SetGlobal 设置全局变量
func (m *Manager) SetGlobal(key string, value interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	scope := m.GetScope(ScopeGlobal, "global")
	if scope == nil {
		scope = NewScope(ScopeGlobal, "global", 0)
		m.AddScope(scope)
	}

	scope.Set(key, value)
}

// SetGroupVars 设置主机组变量
func (m *Manager) SetGroupVars(groupName string, vars map[string]interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	scope := m.GetScope(ScopeGroup, groupName)
	if scope == nil {
		scope = NewScope(ScopeGroup, groupName, 10)
		m.AddScope(scope)
	}

	for k, v := range vars {
		scope.Set(k, v)
	}
}

// SetHostVars 设置主机变量
func (m *Manager) SetHostVars(hostname string, vars map[string]interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	scope := m.GetScope(ScopeHost, hostname)
	if scope == nil {
		scope = NewScope(ScopeHost, hostname, 20)
		m.AddScope(scope)
	}

	for k, v := range vars {
		scope.Set(k, v)
	}
}

// SetTaskVars 设置任务变量
func (m *Manager) SetTaskVars(taskID string, vars map[string]interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	scope := m.GetScope(ScopeTask, taskID)
	if scope == nil {
		scope = NewScope(ScopeTask, taskID, 30)
		m.AddScope(scope)
	}

	for k, v := range vars {
		scope.Set(k, v)
	}
}

// SetTempVars 设置临时变量
func (m *Manager) SetTempVars(name string, vars map[string]interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	scope := m.GetScope(ScopeTemp, name)
	if scope == nil {
		scope = NewScope(ScopeTemp, name, 40)
		m.AddScope(scope)
	}

	for k, v := range vars {
		scope.Set(k, v)
	}
}

// GetAllVars 获取所有变量，按优先级合并
func (m *Manager) GetAllVars() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 按优先级排序
	sortedScopes := make([]*Scope, len(m.scopes))
	copy(sortedScopes, m.scopes)
	sort.Slice(sortedScopes, func(i, j int) bool {
		return sortedScopes[i].Priority < sortedScopes[j].Priority
	})

	// 合并变量，优先级高的覆盖优先级低的
	result := make(map[string]interface{})
	for _, scope := range sortedScopes {
		for k, v := range scope.Vars {
			result[k] = v
		}
	}

	// 添加环境变量
	for _, env := range os.Environ() {
		for i := 0; i < len(env); i++ {
			if env[i] == '=' {
				result[env[:i]] = env[i+1:]
				break
			}
		}
	}

	return result
}

// initEnvScope 初始化环境变量作用域
func (m *Manager) initEnvScope() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 创建环境变量作用域
	envScope := NewScope(ScopeEnv, "env", -10) // 环境变量优先级最低

	// 加载所有环境变量
	for _, env := range os.Environ() {
		for i := 0; i < len(env); i++ {
			if env[i] == '=' {
				envScope.Set(env[:i], env[i+1:])
				break
			}
		}
	}

	m.AddScope(envScope)
}