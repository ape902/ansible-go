# Ansible-Go

## 项目介绍

Ansible-Go 是一个使用 Go 语言实现的自动化配置管理和应用部署工具，类似于 Ansible。它提供了一种简单、高效的方式来管理服务器配置和自动化部署流程。

### 主要特性

- **简单易用**：使用 YAML 格式定义任务和配置
- **模块化设计**：支持多种执行器模块（command、shell、file、template等）
- **变量管理**：支持全局变量、playbook变量和任务变量
- **任务编排**：支持任务依赖和条件执行
- **高性能**：Go语言实现，执行效率高

## 安装部署

### 前置条件

- Go 1.23.0 或更高版本

### 从源码安装

```bash
# 克隆仓库
git clone https://github.com/ape902/ansible-go.git
cd ansible-go

# 编译项目
go build -o ansible-go ./cmd/ansible-go

# 安装到系统路径（可选）
mv ansible-go /usr/local/bin/  # Linux/macOS
# 或者将编译后的可执行文件添加到系统PATH中 (Windows)
```

## 配置说明

### 配置文件

Ansible-Go 使用 YAML 格式的配置文件，主要包含主机清单（inventory）、SSH配置和全局变量（vars）。

#### 基本配置示例

```yaml
# config.yaml
# 主机配置
hosts:
  # 主机清单
  inventory:
    webservers:
      - host: "192.168.1.101"
        port: 22
        alias: "web1"
        connection_type: "ssh"
        vars:
          role: "web"
          user: "webadmin"
          password: "web_password"
      - host: "192.168.1.102"
        port: 22
        alias: "web2"
        connection_type: "ssh"
    dbservers:
      - host: "192.168.1.201"
        port: 22
        alias: "db1"
        connection_type: "ssh"

  # 默认主机组
  default_group: "webservers"

  # 主机变量
  host_vars:
    "192.168.1.101":
      app_port: 8080

  # 组变量
  group_vars:
    webservers:
      http_port: 80
      max_connections: 1000

# SSH配置
ssh:
  # 默认用户名
  user: "admin"
  # 默认密码
  password: "your_password"
  # 默认端口
  port: 22
  # 连接超时时间(秒)
  timeout: 30
  # 是否使用密钥认证
  use_key_auth: false
  # 私钥文件路径（如果使用密钥认证）
  key_file: ""
  # 私钥密码（如果私钥有密码保护）
  key_password: ""
  # 是否禁用主机密钥检查
  disable_host_key_checking: true
  # 是否启用代理跳转
  use_jump_host: false
  # 跳转主机配置
  jump_host:
    host: ""
    port: 22
    user: ""
    password: ""
    key_file: ""
```

### Playbook 文件

Playbook 文件定义了要在目标主机上执行的任务序列。

示例 playbook 文件：

```yaml
# deploy.yaml
name: "部署Web应用"
hosts:
  - webservers
vars:
  app_path: "/var/www/app"
  app_user: "www-data"

tasks:
  - name: "确保应用目录存在"
    module: file
    args:
      path: "{{ app_path }}"
      state: directory
      mode: "0755"
      owner: "{{ app_user }}"

  - name: "部署配置文件"
    module: template
    args:
      src: "./templates/app.conf.j2"
      dest: "/etc/app/app.conf"
      mode: "0644"
    notify:
      - restart app service
```

## 使用方法

### 基本命令

```bash
# 使用指定配置文件执行playbook
ansible-go --config=config.yaml --playbook=deploy.yaml

# 启用详细日志输出
ansible-go --config=config.yaml --playbook=deploy.yaml --verbose
```

### 支持的模块

Ansible-Go 目前支持以下模块：

- **command**：执行命令
- **shell**：执行shell命令
- **file**：文件操作
- **template**：模板渲染
- **copy**：文件复制
- **fetch**：从远程主机获取文件

### 模块使用示例

#### Command 模块

```yaml
- name: "检查磁盘空间"
  module: command
  args:
    cmd: "df -h"
```

#### Template 模块

```yaml
- name: "生成nginx配置"
  module: template
  args:
    src: "./templates/nginx.conf.j2"
    dest: "/etc/nginx/nginx.conf"
    mode: "0644"
```

#### File 模块

```yaml
- name: "创建目录"
  module: file
  args:
    path: "/var/log/app"
    state: directory
    mode: "0755"
```

## 变量使用

Ansible-Go 支持在任务中使用变量，变量引用格式为 `{{ variable_name }}`。变量优先级从高到低为：

1. 任务变量
2. Playbook 变量
3. 全局变量

## 扩展开发

### 添加自定义执行器

可以通过实现 `engine.TaskExecutor` 接口来添加自定义执行器：

```go
type TaskExecutor interface {
    Execute(ctx context.Context, task *models.Task, conn connection.Connection, varStore *vars.Store) (*models.TaskResult, error)
}
```

然后在执行器工厂中注册：

```go
factory.RegisterExecutor("my-module", NewMyModuleExecutor())
```

## 常见问题

### 连接问题

如果遇到连接目标主机失败的问题，请检查：

1. 网络连接是否正常
2. SSH 配置是否正确
3. 目标主机防火墙设置

#### SSH连接配置说明

Ansible-Go支持两种方式配置SSH连接信息：

1. **全局SSH配置**：在配置文件的`ssh`部分设置默认的连接参数，适用于所有主机
   ```yaml
   ssh:
     user: "admin"          # 默认用户名
     password: "password"    # 默认密码
     port: 22               # 默认SSH端口
     timeout: 30            # 连接超时时间(秒)
     use_key_auth: false    # 是否使用密钥认证
     key_file: ""           # 私钥文件路径
     key_password: ""       # 私钥密码
   ```

2. **主机特定配置**：在主机定义中通过`vars`字段设置特定主机的连接参数，会覆盖全局设置
   ```yaml
   hosts:
     inventory:
       webservers:
         - host: "192.168.1.101"
           port: 2222           # 特定主机的SSH端口
           vars:
             user: "webadmin"   # 特定主机的用户名
             password: "web_password" # 特定主机的密码
   ```

这种灵活的配置方式允许您为不同的主机设置不同的连接参数，同时通过全局配置减少重复设置。

### 权限问题

执行某些操作可能需要特定权限，请确保：

1. 使用具有足够权限的用户执行命令
2. 文件和目录权限设置正确

## 许可证

[添加许可证信息]

## 贡献指南

[添加贡献指南]