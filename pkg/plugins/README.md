# 插件系统

## 目录说明

本目录包含了Ansible-Go的插件系统相关代码，负责处理插件的加载、注册和管理。

## 主要功能

- 插件发现和加载
- 插件注册和管理
- 自定义模块支持
- 过滤器扩展支持

## 目录结构

```
plugins/
├── manager.go    # 插件管理器，处理插件的加载和注册
└── types/        # 插件相关的类型定义（待实现）
```

## 插件类型

Ansible-Go 支持多种类型的插件：

1. **模块插件**：扩展系统的执行能力，如自定义执行器
2. **过滤器插件**：扩展变量处理能力，如自定义变量转换
3. **连接插件**：扩展连接方式，如自定义协议支持

## 使用方法

```go
// 加载插件
pluginManager := plugins.NewManager()
err := pluginManager.LoadPlugin("./plugins/custom.so")
if err != nil {
    log.Fatalf("加载插件失败: %v", err)
}

// 注册自定义模块
pluginManager.RegisterModule("custom_module", customModuleFunc)
```

## 开发插件

可以通过实现相应接口来开发新的插件：

```go
// 自定义模块插件示例
package main

import (
    "github.com/ape902/ansible-go/pkg/executor/models"
)

// 插件入口函数
func PluginInit() {
    // 在这里注册插件功能
}

// 自定义模块实现
func CustomModule(args map[string]interface{}, context *models.ExecutionContext) (*models.ModuleResult, error) {
    // 实现模块逻辑
    return &models.ModuleResult{
        Changed: true,
        Output: "自定义模块执行成功",
    }, nil
}
```

## 插件配置

在配置文件中可以启用和配置插件：

```yaml
plugins:
  # 启用的插件列表
  enabled:
    - "custom_module"
  
  # 插件路径
  paths:
    - "./plugins"
    - "/usr/local/lib/ansible-go/plugins"
  
  # 插件特定配置
  config:
    custom_module:
      option1: "value1"
      option2: "value2"
```