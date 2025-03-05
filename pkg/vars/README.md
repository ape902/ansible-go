# 变量系统

## 目录说明

本目录包含了Ansible-Go的变量系统相关代码，负责处理变量的存储、作用域管理和渲染。

## 主要功能

- 变量存储和管理
- 变量作用域控制
- 变量渲染和模板处理
- 变量加密和安全处理

## 目录结构

```
vars/
├── manager.go    # 变量管理器，处理变量作用域和优先级
├── renderer.go   # 变量渲染器，处理模板中的变量替换
├── scope.go      # 变量作用域定义
├── security.go   # 变量加密和安全处理
├── store.go      # 变量存储实现
└── types.go      # 变量系统相关类型定义
```

## 变量优先级

Ansible-Go 中的变量优先级从高到低为：

1. 任务级变量（task vars）
2. 主机变量（host_vars）
3. 组变量（group_vars）
4. Playbook变量（playbook vars）
5. 全局变量（global vars）

## 变量作用域

变量系统支持多种作用域类型，每种作用域有不同的优先级：

- **全局作用域**：适用于所有主机和任务
- **组作用域**：适用于特定组内的所有主机
- **主机作用域**：适用于特定主机
- **Playbook作用域**：适用于特定playbook中的所有任务
- **任务作用域**：仅适用于特定任务

## 使用方法

```go
// 创建变量管理器
manager := vars.NewManager()

// 添加变量作用域
globalScope := vars.NewScope(vars.ScopeTypeGlobal, "global", 0)
globalScope.Set("app_version", "1.0.0")
manager.AddScope(globalScope)

// 获取变量值
value, exists := manager.Get("app_version")

// 渲染模板
renderer := vars.NewRenderer(manager)
result, err := renderer.Render("应用版本: {{ app_version }}")
```

## 变量加密

对于敏感信息，变量系统提供了加密功能：

```go
// 加密变量
encrypted, err := vars.Encrypt("sensitive_data", "encryption_key")

// 解密变量
decrypted, err := vars.Decrypt(encrypted, "encryption_key")
```