package connection

import (
	"time"
)

// ConnectionType 定义连接类型
type ConnectionType string

const (
	// ConnectionTypeSSH SSH连接类型
	ConnectionTypeSSH ConnectionType = "ssh"
	// ConnectionTypeWinRM WinRM连接类型
	ConnectionTypeWinRM ConnectionType = "winrm"
	// ConnectionTypeLocal 本地连接类型
	ConnectionTypeLocal ConnectionType = "local"
	// ConnectionTypeDocker Docker连接类型
	ConnectionTypeDocker ConnectionType = "docker"
)

// ConnectionResult 定义连接执行结果
type ConnectionResult struct {
	Stdout   string        // 标准输出
	Stderr   string        // 标准错误
	ExitCode int           // 退出码
	Duration time.Duration // 执行时长
	Error    error         // 错误信息
}

// Connection 定义连接接口
type Connection interface {
	// Connect 建立连接
	Connect() error
	
	// Disconnect 断开连接
	Disconnect() error
	
	// IsConnected 检查是否已连接
	IsConnected() bool
	
	// ExecuteCommand 执行命令
	ExecuteCommand(command string) (*ConnectionResult, error)
	
	// CopyFile 复制文件到远程主机
	CopyFile(localPath, remotePath string) error
	
	// FetchFile 从远程主机获取文件
	FetchFile(remotePath, localPath string) error
	
	// GetType 获取连接类型
	GetType() ConnectionType
}

// ConnectionManager 定义连接管理器接口
type ConnectionManager interface {
	// GetConnection 获取指定主机的连接
	GetConnection(host string, port int, connType ConnectionType) (Connection, error)
	
	// ReleaseConnection 释放连接
	ReleaseConnection(conn Connection)
	
	// CloseConnection 关闭指定连接
	CloseConnection(host string, port int, connType ConnectionType) error
	
	// CloseAll 关闭所有连接
	CloseAll()
	
	// CleanIdleConnections 清理空闲连接
	CleanIdleConnections()
}