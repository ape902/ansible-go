# API 文档

## SSH连接相关API

### 建立SSH连接
```go
func (p *SSHConnectionPool) GetConnection(host string, port int, user string, password string, keyFile string) (*SSHConnection, error)
```

### 执行命令
```go
func (conn *SSHConnection) ExecuteCommand(command string) (*ConnectionResult, error)
```

### 文件传输
```go
// 上传文件
func (conn *SSHConnection) CopyFile(localPath, remotePath string) error

// 下载文件
func (conn *SSHConnection) FetchFile(remotePath, localPath string) error
```

## 连接管理API

### 获取连接
```go
func (m *ConnectionManagerImpl) GetConnection(host string, port int, connType ConnectionType) (Connection, error)
```

### 关闭连接
```go
func (m *ConnectionManagerImpl) CloseConnection(host string, port int, connType ConnectionType) error
```