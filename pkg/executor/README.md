# 执行器系统

## 目录说明

本目录包含了Ansible-Go的执行器系统相关代码，负责处理任务的执行、连接管理和结果处理。

## 主要功能

- 任务执行和编排
- 模块化执行器实现
- 远程连接管理
- 执行结果处理和错误管理

## 目录结构

```
executor/
├── executor.go         # 执行器核心代码
├── connection/         # 连接管理相关代码
│   ├── connection.go   # 连接接口定义
│   └── ssh.go          # SSH连接实现
├── engine/             # 执行引擎
│   └── engine.go       # 执行引擎实现
├── executors/          # 具体执行器实现
│   ├── command.go      # 命令执行器
│   ├── copy.go         # 文件复制执行器
│   ├── factory.go      # 执行器工厂
│   ├── fetch.go        # 文件获取执行器
│   ├── file.go         # 文件管理执行器
│   ├── shell.go        # Shell执行器
│   └── template.go     # 模板执行器
└── models/             # 数据模型
    ├── queue.go        # 任务队列
    └── task.go         # 任务模型
```

## 执行器类型

系统支持多种执行器类型，每种类型负责不同的任务执行：

1. **Command执行器**：执行命令，不通过shell解释器
2. **Shell执行器**：通过shell解释器执行命令，支持管道、重定向等shell特性
3. **File执行器**：管理文件和目录，包括创建、删除、权限设置等
4. **Template执行器**：使用模板生成文件，支持变量替换
5. **Copy执行器**：复制文件到目标主机
6. **Fetch执行器**：从远程主机获取文件

## 使用方法

```go
// 创建执行器
executor, err := executor.NewExecutor(config, varStore)
if err != nil {
    log.Fatalf("创建执行器失败: %v", err)
}

// 执行任务
err = executor.Execute(task)
if err != nil {
    log.Fatalf("执行任务失败: %v", err)
}
```

## 扩展执行器

可以通过实现相应接口来扩展新的执行器类型：

```go
// 自定义执行器示例
type CustomExecutor struct {}

func (e *CustomExecutor) Execute(args map[string]interface{}, context *models.ExecutionContext) (*models.ExecutionResult, error) {
    // 实现执行逻辑
    return &models.ExecutionResult{
        Changed: true,
        Output: "自定义执行器执行成功",
    }, nil
}

// 注册自定义执行器
executor.RegisterExecutor("custom", &CustomExecutor{})
```