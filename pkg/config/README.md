# 配置系统

## 目录说明

本目录包含了Ansible-Go的配置系统相关代码，负责处理配置文件的加载、解析和管理。

## 主要功能

- 配置文件加载和解析
- YAML格式配置支持
- 配置验证和默认值处理
- 配置合并和覆盖规则

## 目录结构

```
config/
├── config.go      # 配置加载和管理的核心代码
└── types/         # 配置相关的类型定义
    ├── log.go         # 日志配置类型
    ├── plugin.go      # 插件配置类型
    ├── server.go      # 服务器配置类型
    ├── task.go        # 任务配置类型
    ├── task_config.go # 任务详细配置类型
    └── types.go       # 通用类型定义
```

## 配置文件格式

配置文件使用YAML格式，主要包含以下部分：

1. 主机清单（inventory）
2. SSH配置
3. 变量配置
4. 任务配置
5. 插件配置

详细的配置示例和说明请参考项目根目录的README.md文件。

## 使用方法

```go
// 加载配置文件
config, err := config.LoadConfig("config.yaml")
if err != nil {
    log.Fatalf("加载配置失败: %v", err)
}

// 使用配置
inventory := config.Inventory
vars := config.Vars
```