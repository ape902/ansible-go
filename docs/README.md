# Ansible-Go 项目文档

## 项目概述
Ansible-Go 是一个用Go语言实现的Ansible-like自动化工具。

## 目录结构
```
├── cmd
│   └── ansible-go
│       └── main.go
├── pkg
│   ├── config
│   ├── executor
│   ├── logger
│   ├── plugins
│   └── vars
├── test-project
└── docs
```

## 技术选型
- 语言: Go
- SSH库: golang.org/x/crypto/ssh
- 配置管理: 自定义YAML解析

## 启动命令
```bash
go run cmd/ansible-go/main.go
```