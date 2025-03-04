package vars

// ScopeType 定义变量作用域类型
type ScopeType int

// 变量作用域类型常量
const (
	ScopeGlobal ScopeType = iota // 全局作用域
	ScopeGroup                    // 主机组作用域
	ScopeHost                     // 主机作用域
	ScopeTask                     // 任务作用域
	ScopeTemp                     // 临时作用域
	ScopeEnv                      // 环境变量作用域
)

// Scope 定义变量作用域
type Scope struct {
	Type     ScopeType              // 作用域类型
	Name     string                 // 作用域名称（如主机组名、主机名、任务ID等）
	Priority int                    // 优先级（数字越大优先级越高）
	Vars     map[string]interface{} // 该作用域下的变量
	Parent   *Scope                // 父作用域
}

// NewScope 创建新的作用域
func NewScope(scopeType ScopeType, name string, priority int) *Scope {
	return &Scope{
		Type:     scopeType,
		Name:     name,
		Priority: priority,
		Vars:     make(map[string]interface{}),
	}
}

// SetParent 设置父作用域
func (s *Scope) SetParent(parent *Scope) {
	s.Parent = parent
}

// Get 获取变量值，如果本作用域没有则向上查找
func (s *Scope) Get(key string) (interface{}, bool) {
	// 先在当前作用域查找
	if val, exists := s.Vars[key]; exists {
		return val, true
	}

	// 如果有父作用域，则向上查找
	if s.Parent != nil {
		return s.Parent.Get(key)
	}

	return nil, false
}

// Set 设置变量值
func (s *Scope) Set(key string, value interface{}) {
	s.Vars[key] = value
}

// Delete 删除变量
func (s *Scope) Delete(key string) {
	delete(s.Vars, key)
}

// GetAll 获取当前作用域的所有变量
func (s *Scope) GetAll() map[string]interface{} {
	result := make(map[string]interface{})

	// 如果有父作用域，先获取父作用域的变量
	if s.Parent != nil {
		for k, v := range s.Parent.GetAll() {
			result[k] = v
		}
	}

	// 当前作用域的变量覆盖父作用域的变量
	for k, v := range s.Vars {
		result[k] = v
	}

	return result
}