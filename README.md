# Ansible-Go

## 项目概述

Ansible-Go 是一个基于 Go 语言开发的自动化配置管理和应用部署工具，提供类似于 Ansible 的功能。本项目旨在提供一个高性能、易用且可扩展的自动化运维解决方案。

### 核心特性

- **高性能执行引擎**：基于 Go 语言实现，提供卓越的执行效率
- **声明式配置**：使用 YAML 格式定义任务和配置，简单直观
- **模块化架构**：
  - 内置多种执行器模块（command、shell、file、template等）
  - 支持自定义模块扩展
  - 灵活的插件系统
- **强大的变量系统**：
  - 多层级变量管理（全局、主机组、主机、任务）
  - 支持变量加密
  - 丰富的变量操作符
- **任务编排能力**：
  - 支持任务依赖关系
  - 条件执行
  - 错误处理机制
- **跨平台支持**：完整支持 Linux、macOS 和 Windows
- **安全性设计**：
  - 内置变量加密机制
  - 安全的凭证管理
  - 最小权限原则

## 快速开始

### 环境要求

- Go 1.23.0 或更高版本
- SSH 客户端（用于远程执行）

### 安装

```bash
# 从源码安装
git clone https://github.com/ape902/ansible-go.git
cd ansible-go

# 编译
go build -o ansible-go ./cmd/ansible-go

# 安装到系统路径（可选）
sudo mv ansible-go /usr/local/bin/  # Linux/macOS
```

### 项目初始化

```bash
# 基本初始化
ansible-go init

# 自定义初始化
ansible-go init --name=my-project --path=/path/to/workspace --template=standard
```

初始化参数说明：
- `init`：初始化子命令
- `--name`：项目名称（默认：ansible-go-project）
- `--path`：项目路径（默认：当前目录）
- `--template`：项目模板类型（可选：standard, minimal, full，默认：standard）

初始化将创建以下目录结构：
```
./
├── config.yaml     # 主配置文件
├── tasks/         # 任务文件目录
│   ├── main.yaml  # 主任务文件
│   └── example.yaml # 示例任务文件
├── vars/          # 变量文件目录
│   └── main.yaml  # 主变量文件
├── files/         # 静态文件目录
└── executor/      # 执行器配置目录
```

### 基础配置

创建 `config.yaml`：

```yaml
# 主机清单配置
inventory:
  # 主机组名: 主机列表
  localhost: 
    - host: 127.0.0.1
      port: 22
      connection_type: ssh
  webservers:
    - host: "192.168.1.101"
      port: 22
      connection_type: ssh
      alias: "web1"
    - host: "192.168.1.102"
      port: 22
      connection_type: ssh
      alias: "web2"
  dbservers:
    - host: "192.168.1.201"
      port: 22
      connection_type: ssh
      alias: "db1"

# SSH连接配置
ssh:
  user: root
  private_key_path: ~/.ssh/id_rsa
  timeout: 10s
  known_hosts_file: ~/.ssh/known_hosts
  use_key_auth: true  # 是否使用密钥认证
  max_parallel: 5     # 最大并行执行数

# 全局变量
vars:
  app_version: "1.0.0"
  deploy_path: "/opt/myapp"
  environment: "production"

# 日志配置
log:
  level: info
  format: text
  file: ./ansible-go.log
```

### 创建任务

创建 `tasks/main.yaml`：

```yaml
name: "应用部署"
description: "这是一个应用部署任务配置"

# 目标主机组
hosts: ["webservers", "localhost"]

# 任务变量
vars:
  deploy_path: "/opt/myapp"
  app_port: 8080
  app_user: "www-data"

# 任务列表
tasks:
  - "创建应用目录":
      module: "command"
      args:
        cmd: "mkdir -p {{ deploy_path }}/{{ app_version }}"

  - "部署配置文件":
      module: "command"
      args:
        cmd: "cp ./templates/app.conf {{ deploy_path }}/config/app.conf"

# 处理器
handlers:
  - name: "重启应用"
    module: "command"
    args:
      cmd: "systemctl restart myapp"
```

### 执行任务

```bash
# 基本执行
ansible-go --config=config.yaml

# 高级选项
ansible-go --config=config.yaml \
          --parallel=5 \
          --tags=config,service \
          --verbose
```

### 检查配置

```bash
# 检查配置文件合规性
ansible-go check --config=config.yaml

# 检查特定任务文件
ansible-go check --task=tasks/example.yaml

# 检查整个项目
ansible-go check --project=./

# 详细检查报告
ansible-go check --config=config.yaml --verbose
```

检查功能会验证：
- 配置文件语法正确性
- 任务文件格式规范性
- 变量引用完整性
- 模块参数有效性
- 主机组配置有效性
- 文件权限合规性
- SSH连接配置完整性

## 核心概念

### 配置系统

#### 配置层级

1. 系统配置：全局默认设置
2. 项目配置：项目级别的配置
3. Playbook配置：特定任务的配置
4. 运行时配置：命令行参数指定的配置

#### 变量系统

变量优先级（从高到低）：
1. 命令行变量（--extra-vars）
2. 任务变量（task vars）
3. 主机变量（host_vars）
4. 组变量（group_vars）
5. Playbook变量
6. 全局变量

### 任务系统

#### 任务类型

- **标准任务**：常规的操作任务
- **处理器任务**：由其他任务触发的任务
- **条件任务**：基于条件执行的任务
- **循环任务**：重复执行的任务

#### 执行流程

1. 配置加载和验证
2. 主机清单解析
3. 变量处理和替换
4. 任务依赖分析
5. 任务执行和状态跟踪
6. 处理器触发和执行

## 模块参考

### 内置模块

#### command
```yaml
- name: "执行命令"
  module: command
  args:
    cmd: "df -h"
    chdir: "/var/log"  # 工作目录
    creates: "/tmp/flag"  # 文件存在则跳过
    removes: "/tmp/lock"  # 文件不存在则跳过
```

#### shell
```yaml
- name: "Shell操作"
  module: shell
  args:
    cmd: "find /var -type f -size +100M | sort -nr"
    executable: "/bin/bash"
```

#### file
```yaml
- name: "文件操作"
  module: file
  args:
    path: "/var/www/app"
    state: directory  # file/directory/link/absent
    mode: "0755"
    owner: "www-data"
    recurse: true
```

## 最佳实践

### 安全性建议

1. **凭证管理**
   - 使用密钥认证替代密码
   - 对敏感信息进行加密
   - 使用专用的部署账户

2. **权限控制**
   - 遵循最小权限原则
   - 合理使用 sudo 提权
   - 定期轮换密钥和凭证

### 性能优化

1. **并行执行**
   - 合理设置并行度
   - 注意资源限制
   - 避免过度并行

2. **连接优化**
   - 启用 SSH 连接复用
   - 设置合适的超时时间
   - 使用连接池

### 开发规范

1. **代码风格**
   - 遵循 Go 语言规范
   - 使用统一的命名约定
   - 编写完整的注释

2. **测试要求**
   - 单元测试覆盖率 > 80%
   - 提供集成测试
   - 包含性能测试

## 故障排查

### 常见问题

1. **连接问题**
   - 检查网络连通性
   - 验证 SSH 配置
   - 确认防火墙规则

2. **权限问题**
   - 检查用户权限
   - 验证文件权限
   - 确认 sudo 配置

3. **执行问题**
   - 查看详细日志
   - 检查变量值
   - 验证模块参数

### 调试方法

1. **日志分析**
   ```bash
   ansible-go --verbose --log-level=debug
   ```

2. **状态检查**
   ```bash
   ansible-go --check --diff
   ```


## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。