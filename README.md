# Ansible-Go

## 项目介绍

Ansible-Go 是一个使用 Go 语言实现的自动化配置管理和应用部署工具，类似于 Ansible。它提供了一种简单、高效的方式来管理服务器配置和自动化部署流程。

### 主要特性

- **简单易用**：使用 YAML 格式定义任务和配置
- **模块化设计**：支持多种执行器模块（command、shell、file、template等）
- **变量管理**：支持全局变量、playbook变量和任务变量
- **任务编排**：支持任务依赖和条件执行
- **高性能**：Go语言实现，执行效率高
- **安全可靠**：支持变量加密和安全传输
- **跨平台**：支持Linux、macOS和Windows

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

### 初始化项目

使用 `ansible-go` 命令的 `--init` 参数来初始化一个新项目：

```bash
# 使用默认设置初始化项目
ansible-go --init

# 指定项目名称和路径
ansible-go --init --name=my-project --path=/path/to/workspace
```

参数说明：
- `--init`：启用项目初始化模式
- `--name`：指定项目名称，默认为 "ansible-go-project"
- `--path`：指定项目创建路径，默认为当前目录

初始化后的项目结构：
```
my-project/
├── pkg/
│   ├── config/
│   │   └── config.yaml  # 默认配置文件
│   ├── executor/
│   └── vars/
```

## 快速开始

### 创建配置文件

创建一个名为 `config.yaml` 的配置文件：

```yaml
# 主机清单
inventory:
  webservers: [192.168.1.101, 192.168.1.102]
  dbservers: [192.168.1.201]
  localhost: [127.0.0.1]

# 全局变量
vars:
  app_version: "1.0.0"
  deploy_path: "/opt/myapp"
  environment: "production"
```

### 创建Playbook

创建一个名为 `deploy.yaml` 的playbook文件：

```yaml
name: "部署应用"
hosts:
  - webservers
vars:
  app_port: 8080
  app_user: "www-data"

tasks:
  - name: "创建应用目录"
    module: file
    args:
      path: "{{ deploy_path }}/{{ app_version }}"
      state: directory
      mode: "0755"
      owner: "{{ app_user }}"

  - name: "部署应用配置"
    module: template
    args:
      src: "./templates/app.conf.j2"
      dest: "{{ deploy_path }}/config/app.conf"
      mode: "0644"
    notify:
      - restart app service

  - name: "启动应用服务"
    module: command
    args:
      cmd: "systemctl start myapp"
    when: "environment == 'production'"
```

### 执行Playbook

```bash
# 执行部署
ansible-go --config=config.yaml --playbook=deploy.yaml

# 启用详细日志
ansible-go --config=config.yaml --playbook=deploy.yaml --verbose
```

参数说明：
- `--config`：指定配置文件路径，默认为当前目录下的 config.yaml
- `--playbook`：指定要执行的playbook文件路径
- `--verbose`：启用详细日志输出，用于调试和问题排查
- `--parallel`：设置任务并行执行的数量，默认为1
- `--tags`：指定要执行的任务标签，多个标签用逗号分隔
- `--skip-tags`：指定要跳过的任务标签，多个标签用逗号分隔
- `--limit`：限制执行的主机范围，支持主机组和排除语法
- `--check`：启用检查模式，不实际执行任务
- `--diff`：显示文件更改的详细差异
- `--timeout`：设置任务执行超时时间（秒），默认为300秒
- `--retry`：设置任务失败后的重试次数，默认为0
- `--inventory`：指定额外的主机清单文件路径
- `--extra-vars`：设置额外的变量，格式为 key=value，多个变量用空格分隔

## 配置详解

### 配置文件结构

Ansible-Go 使用 YAML 格式的配置文件，主要包含以下部分：

#### 主机清单（inventory）

```yaml
inventory:
  # 简单格式：组名和IP列表
  webservers: [192.168.1.101, 192.168.1.102]
  
  # 详细格式：包含主机详细信息
  dbservers:
    - host: "192.168.1.201"
      port: 22
      alias: "db1"
      connection_type: "ssh"
      vars:
        role: "master"
        backup_time: "03:00"
```

#### SSH配置

```yaml
ssh:
  user: "admin"              # 默认用户名
  password: "your_password"  # 默认密码
  port: 22                   # 默认端口
  timeout: 30                # 连接超时时间(秒)
  use_key_auth: true         # 是否使用密钥认证
  key_file: "~/.ssh/id_rsa"  # 私钥文件路径
  disable_host_key_checking: true  # 是否禁用主机密钥检查
```

#### 变量配置

```yaml
# 全局变量
vars:
  app_version: "1.0.0"
  deploy_path: "/opt/myapp"

# 主机变量
host_vars:
  "192.168.1.101":
    http_port: 8080
    max_connections: 1000

# 组变量
group_vars:
  webservers:
    http_port: 80
    app_user: "www-data"
```

## 变量系统

### 变量优先级

Ansible-Go 中的变量优先级从高到低为：

1. 任务级变量（task vars）
2. 主机变量（host_vars）
3. 组变量（group_vars）
4. Playbook变量（playbook vars）
5. 全局变量（global vars）

### 变量引用

在配置文件、Playbook和模板中，可以使用 `{{ variable_name }}` 语法引用变量：

```yaml
- name: "创建应用目录"
  module: file
  args:
    path: "{{ deploy_path }}/{{ app_version }}"
    mode: "0755"
```

### 变量操作

Ansible-Go 支持以下变量操作：

- **字符串拼接**：`{{ var1 ~ var2 }}`
- **数学运算**：`{{ port + 1 }}`
- **条件表达式**：`{{ (env == 'prod') | ternary('production', 'development') }}`
- **默认值**：`{{ variable | default('default_value') }}`

### 变量加密

对于敏感信息，Ansible-Go 提供了变量加密功能：

```yaml
vars:
  db_password: !encrypted "your_encrypted_password"
```

## 任务编排

### 任务定义

每个任务包含以下属性：

- **name**：任务名称（必填）
- **module**：使用的模块（必填）
- **args**：模块参数（必填）
- **when**：执行条件（可选）
- **register**：注册变量（可选）
- **notify**：通知处理器（可选）
- **tags**：任务标签（可选）
- **ignore_errors**：是否忽略错误（可选）

### 条件执行

使用 `when` 属性定义任务执行条件：

```yaml
- name: "仅在生产环境执行"
  module: command
  args:
    cmd: "systemctl start myapp"
  when: "environment == 'production'"
```

### 任务依赖

使用 `depends` 属性定义任务依赖关系：

```yaml
- name: "部署应用"
  module: copy
  args:
    src: "./app.tar.gz"
    dest: "{{ deploy_path }}"
  register: deploy_result

- name: "解压应用"
  module: command
  args:
    cmd: "tar -xzf {{ deploy_path }}/app.tar.gz -C {{ deploy_path }}"
  depends: ["部署应用"]
  when: "deploy_result.changed"
```

### 处理器（Handlers）

处理器是特殊的任务，只有在被通知时才会执行：

```yaml
tasks:
  - name: "更新配置文件"
    module: template
    args:
      src: "./nginx.conf.j2"
      dest: "/etc/nginx/nginx.conf"
    notify:
      - restart nginx

handlers:
  - name: "restart nginx"
    module: command
    args:
      cmd: "systemctl restart nginx"
```

## 支持的模块

Ansible-Go 目前支持以下模块：

### command 模块

执行命令，不通过shell解释器。

```yaml
- name: "检查磁盘空间"
  module: command
  args:
    cmd: "df -h"
    chdir: "/var/log"  # 可选，指定工作目录
    creates: "/tmp/flag.txt"  # 可选，如果文件存在则跳过
    removes: "/tmp/lock"  # 可选，如果文件不存在则跳过
```

### shell 模块

通过shell解释器执行命令，支持管道、重定向等shell特性。

```yaml
- name: "查找大文件"
  module: shell
  args:
    cmd: "find /var -type f -size +100M | sort -nr"
    executable: "/bin/bash"  # 可选，指定shell解释器
```

### file 模块

管理文件和目录。

```yaml
- name: "创建目录"
  module: file
  args:
    path: "/var/www/app"
    state: directory  # 可选值：file, directory, link, absent
    mode: "0755"
    owner: "www-data"
    group: "www-data"
    recurse: true  # 可选，递归设置权限
```

### template 模块

使用模板生成文件。

```yaml
- name: "生成配置文件"
  module: template
  args:
    src: "./templates/app.conf.j2"
    dest: "/etc/app/app.conf"
    mode: "0644"
    owner: "root"
    group: "root"
    backup: true  # 可选，备份原文件
```

### copy 模块

复制文件到目标主机。

```yaml
- name: "复制文件"
  module: copy
  args:
    src: "./files/app.conf"
    dest: "/etc/app/app.conf"
    mode: "0644"
    owner: "root"
    group: "root"
    backup: true  # 可选，备份原文件
```

### fetch 模块

从远程主机获取文件。

```yaml
- name: "获取日志文件"
  module: fetch
  args:
    src: "/var/log/app.log"
    dest: "./logs/"
    flat: true  # 可选，不创建主机名目录
```

## 高级功能

### 并行执行

通过设置并行度来提高执行效率：

```bash
ansible-go --config=config.yaml --playbook=deploy.yaml --parallel=5
```

### 标签过滤

使用标签来选择性执行任务：

```bash
# 只执行带有指定标签的任务
ansible-go --config=config.yaml --playbook=deploy.yaml --tags=config,service

# 跳过带有指定标签的任务
ansible-go --config=config.yaml --playbook=deploy.yaml --skip-tags=notification
```

### 限制主机

限制playbook只在特定主机上执行：

```bash
ansible-go --config=config.yaml --playbook=deploy.yaml --limit="webservers:!web2"
```

### 插件系统

Ansible-Go 支持插件扩展，可以自定义模块和过滤器：

```go
// 自定义模块示例
package modules

import (
    "github.com/ape902/ansible-go/pkg/executor/models"
)

type CustomModule struct {}

func (m *CustomModule) Execute(args map[string]interface{}, context *models.ExecutionContext) (*models.ModuleResult, error) {
    // 实现模块逻辑
    return &models.ModuleResult{
        Changed: true,
        Output: "自定义模块执行成功",
    }, nil
}
```

## 贡献指南

### 开发环境设置

```bash
# 克隆仓库
git clone https://github.com/ape902/ansible-go.git
cd ansible-go

# 安装依赖
go mod download

# 运行测试
go test ./...
```

### 提交代码

1. Fork 项目仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 常见问题

### 连接问题

**问题**：无法连接到远程主机
**解决方案**：
- 检查网络连接和防火墙设置
- 确认SSH配置正确，包括用户名、密码或密钥
- 使用 `--verbose` 参数查看详细连接日志

### 权限问题

**问题**：执行任务时出现权限错误
**解决方案**：
- 确保远程用户具有足够的权限
- 对于需要提升权限的操作，在配置文件中设置 `become: true`
- 检查文件和目录的所有者和权限设置

### 变量问题

**问题**：变量未正确解析或替换
**解决方案**：
- 检查变量名称拼写是否正确
- 确认变量的作用域和优先级
- 使用 `--verbose` 参数查看变量解析过程

### 模块问题

**问题**：模块执行失败
**解决方案**：
- 查看模块的详细文档，确保参数使用正确
- 检查目标系统是否满足模块的依赖要求
- 尝试在目标主机上手动执行相应操作，确认可行性

## 性能优化

### 并行执行

通过调整并行度可以显著提高执行效率：

```bash
# 设置并行执行的任务数
ansible-go --config=config.yaml --playbook=deploy.yaml --parallel=10
```

### 连接复用

Ansible-Go 支持SSH连接复用，减少连接建立的开销：

```yaml
ssh:
  connection_reuse: true
  control_path: "~/.ssh/ansible-go-%%h-%%p-%%r"
```

### 缓存优化

对于频繁执行的任务，可以启用事实缓存：

```yaml
cache:
  enabled: true
  ttl: 3600  # 缓存有效期（秒）
  path: "~/.ansible-go/cache"
```

## 安全最佳实践

### 变量加密

对敏感信息使用变量加密：

```bash
# 加密变量
ansible-go --encrypt "your_sensitive_data" --key="your_encryption_key"

# 在配置文件中使用加密变量
vars:
  db_password: !encrypted "encrypted_string_here"
```

### 凭证管理

避免在配置文件中明文存储凭证：

```yaml
ssh:
  credential_file: "~/.ansible-go/credentials.yaml"  # 使用单独的凭证文件
  use_system_keyring: true  # 使用系统密钥环
```

### 权限控制

遵循最小权限原则：

```yaml
ssh:
  user: "deploy"  # 使用专用部署账户
  become: true  # 需要时提升权限
  become_method: "sudo"  # 提权方式
  become_user: "root"  # 提权后的用户
```

## 版本历史

### v1.0.0 (2023-10-01)

- 初始版本发布
- 支持基本的任务执行和变量管理
- 实现核心模块：command、shell、file、template、copy

### v1.1.0 (2023-12-15)

- 添加并行执行功能
- 增强变量系统，支持更多操作符
- 新增fetch模块
- 改进错误处理和日志记录

## 许可证

Ansible-Go 使用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。