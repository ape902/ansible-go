package connection

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"
)

// LocalConnection 本地连接实现
type LocalConnection struct {
	connected bool
}

// NewLocalConnection 创建新的本地连接
func NewLocalConnection() *LocalConnection {
	return &LocalConnection{
		connected: false,
	}
}

// Connect 建立连接
func (c *LocalConnection) Connect() error {
	c.connected = true
	return nil
}

// Disconnect 断开连接
func (c *LocalConnection) Disconnect() error {
	c.connected = false
	return nil
}

// IsConnected 检查是否已连接
func (c *LocalConnection) IsConnected() bool {
	return c.connected
}

// ExecuteCommand 执行命令
func (c *LocalConnection) ExecuteCommand(command string) (*ConnectionResult, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("连接未建立")
	}

	// 记录开始时间
	startTime := time.Now()

	// 创建命令
	cmd := exec.Command("sh", "-c", command)

	// 捕获标准输出和错误
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 执行命令
	err := cmd.Run()

	// 计算执行时间
	duration := time.Since(startTime)

	// 创建结果
	result := &ConnectionResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	// 设置退出码
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.Error = err
			result.ExitCode = -1
		}
	}

	return result, nil
}

// CopyFile 复制文件到远程主机
func (c *LocalConnection) CopyFile(localPath, remotePath string) error {
	return fmt.Errorf("本地连接不支持文件复制操作")
}

// FetchFile 从远程主机获取文件
func (c *LocalConnection) FetchFile(remotePath, localPath string) error {
	return fmt.Errorf("本地连接不支持文件获取操作")
}

// GetType 获取连接类型
func (c *LocalConnection) GetType() ConnectionType {
	return ConnectionTypeLocal
}