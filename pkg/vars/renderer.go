package vars

import (
	"bytes"
	"fmt"
	"text/template"
)

// Renderer 定义变量渲染器
type Renderer struct {
	manager *Manager
}

// NewRenderer 创建新的变量渲染器
func NewRenderer(manager *Manager) *Renderer {
	return &Renderer{
		manager: manager,
	}
}

// RenderString 渲染包含变量引用的字符串
func (r *Renderer) RenderString(text string, extraVars map[string]interface{}) (string, error) {
	// 创建模板
	tmpl, err := template.New("expr").Funcs(r.getFuncMap()).Parse(text)
	if err != nil {
		return "", fmt.Errorf("解析模板失败: %w", err)
	}

	// 准备变量
	vars := r.manager.GetAllVars()
	if extraVars != nil {
		for k, v := range extraVars {
			vars[k] = v
		}
	}

	// 渲染模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("渲染模板失败: %w", err)
	}

	return buf.String(), nil
}

// RenderValue 渲染变量值中的变量引用
func (r *Renderer) RenderValue(value interface{}, extraVars map[string]interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return r.RenderString(v, extraVars)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			renderedVal, err := r.RenderValue(val, extraVars)
			if err != nil {
				return nil, err
			}
			result[k] = renderedVal
		}
		return result, nil
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			renderedVal, err := r.RenderValue(val, extraVars)
			if err != nil {
				return nil, err
			}
			result[i] = renderedVal
		}
		return result, nil
	default:
		return v, nil
	}
}

// getFuncMap 获取模板函数映射
func (r *Renderer) getFuncMap() template.FuncMap {
	return template.FuncMap{
		// 变量访问函数
		"var": func(name string) interface{} {
			if val, exists := r.manager.Get(name); exists {
				return val
			}
			return nil
		},

		// 条件判断函数
		"if": func(cond bool, trueVal, falseVal interface{}) interface{} {
			if cond {
				return trueVal
			}
			return falseVal
		},

		// 默认值函数
		"default": func(val, defaultVal interface{}) interface{} {
			if val == nil {
				return defaultVal
			}
			return val
		},

		// 环境变量函数
		"env": func(name string) interface{} {
			if val, exists := r.manager.Get(name); exists {
				return val
			}
			return nil
		},

		// 主机变量函数
		"hostvars": func(hostname, varname string) interface{} {
			scope := r.manager.GetScope(ScopeHost, hostname)
			if scope == nil {
				return nil
			}
			if val, exists := scope.Get(varname); exists {
				return val
			}
			return nil
		},

		// 组变量函数
		"groupvars": func(groupname, varname string) interface{} {
			scope := r.manager.GetScope(ScopeGroup, groupname)
			if scope == nil {
				return nil
			}
			if val, exists := scope.Get(varname); exists {
				return val
			}
			return nil
		},
	}
}
