# 数据字典

## 核心数据结构

### SSHConnection
```go
type SSHConnection struct {
    Client   *ssh.Client
    Host     string
    Port     int
    User     string
    Password string
    LastUsed time.Time
    IsInUse  bool
}
```

### ConnectionResult
```go
type ConnectionResult struct {
    Stdout   string        // 标准输出
    Stderr   string        // 标准错误
    ExitCode int           // 退出码
    Duration time.Duration // 执行时长
    Error    error         // 错误信息
}
```

### TaskEngine
```go
type TaskEngine struct {
    ConnectionPool *SSHConnectionPool
    MaxRetries     int
    RetryInterval  time.Duration
    Timeout        time.Duration
}
```

## 技术术语

| 术语 | 说明 |
|------|------|
| SSH连接池 | 管理多个SSH连接，提高连接复用率，支持并发安全访问 |
| 指数退避 | 连接失败时按指数增长等待时间重试，初始间隔1秒，最大间隔30秒 |
| 主机密钥 | SSH连接时用于验证主机身份的密钥，支持RSA/ECDSA/Ed25519算法 |
| 任务队列 | 先进先出(FIFO)的任务执行队列，支持优先级调度 |
| 连接复用 | 空闲连接自动回收机制，默认5分钟未使用自动关闭 |