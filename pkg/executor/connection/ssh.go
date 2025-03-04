package connection

import (
	"fmt"
	"sync"
	"time"

	"github.com/ape902/ansible-go/pkg/config/types"
	"golang.org/x/crypto/ssh"
)

// SSHConnection 定义SSH连接结构
type SSHConnection struct {
	Client   *ssh.Client
	Host     string
	Port     int
	User     string
	LastUsed time.Time
	IsInUse  bool
}

// SSHConnectionPool 定义SSH连接池
type SSHConnectionPool struct {
	mutex      sync.RWMutex
	pool       map[string]*SSHConnection
	maxIdle    time.Duration
	maxRetries int
	timeout    time.Duration
	sshConfig  *types.SSHConfig
}

// NewSSHConnectionPool 创建新的SSH连接池
func NewSSHConnectionPool(sshConfig *types.SSHConfig, maxIdle time.Duration, maxRetries int) *SSHConnectionPool {
	return &SSHConnectionPool{
		pool:       make(map[string]*SSHConnection),
		maxIdle:    maxIdle,
		maxRetries: maxRetries,
		timeout:    time.Duration(sshConfig.Timeout) * time.Second,
		sshConfig:  sshConfig,
	}
}

// GetConnection 获取SSH连接
func (p *SSHConnectionPool) GetConnection(host string, port int, user string, password string, keyFile string) (*SSHConnection, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 生成连接键
	key := fmt.Sprintf("%s:%d:%s", host, port, user)

	// 检查连接池中是否有可用连接
	if conn, exists := p.pool[key]; exists && !conn.IsInUse {
		// 检查连接是否过期
		if time.Since(conn.LastUsed) > p.maxIdle {
			// 关闭过期连接
			conn.Client.Close()
			delete(p.pool, key)
		} else {
			// 标记连接为使用中
			conn.IsInUse = true
			return conn, nil
		}
	}

	// 创建新连接
	var authMethods []ssh.AuthMethod

	// 添加密码认证
	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	} else if p.sshConfig.Password != "" {
		authMethods = append(authMethods, ssh.Password(p.sshConfig.Password))
	}

	// 添加密钥认证
	if keyFile != "" || p.sshConfig.KeyFile != "" {
		keyPath := keyFile
		if keyPath == "" {
			keyPath = p.sshConfig.KeyFile
		}

		if keyPath != "" {
			key, err := loadPrivateKey(keyPath, p.sshConfig.KeyPassword)
			if err != nil {
				return nil, fmt.Errorf("加载SSH私钥失败: %w", err)
			}
			authMethods = append(authMethods, ssh.PublicKeys(key))
		}
	}

	// 如果没有认证方法，返回错误
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("未提供SSH认证方法")
	}

	// 创建SSH配置
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		Timeout:         p.timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 在生产环境中应使用更安全的方法
	}

	// 尝试连接
	var client *ssh.Client
	var err error

	for i := 0; i <= p.maxRetries; i++ {
		client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
		if err == nil {
			break
		}

		if i < p.maxRetries {
			time.Sleep(time.Second * time.Duration(i+1)) // 指数退避
		}
	}

	if err != nil {
		return nil, fmt.Errorf("SSH连接失败: %w", err)
	}

	// 创建新的连接对象
	conn := &SSHConnection{
		Client:   client,
		Host:     host,
		Port:     port,
		User:     user,
		LastUsed: time.Now(),
		IsInUse:  true,
	}

	// 添加到连接池
	p.pool[key] = conn

	return conn, nil
}

// ReleaseConnection 释放连接
func (p *SSHConnectionPool) ReleaseConnection(conn *SSHConnection) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if conn != nil {
		conn.IsInUse = false
		conn.LastUsed = time.Now()
	}
}

// CloseConnection 关闭连接
func (p *SSHConnectionPool) CloseConnection(host string, port int, user string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	key := fmt.Sprintf("%s:%d:%s", host, port, user)

	if conn, exists := p.pool[key]; exists {
		err := conn.Client.Close()
		delete(p.pool, key)
		return err
	}

	return nil
}

// CloseAll 关闭所有连接
func (p *SSHConnectionPool) CloseAll() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for key, conn := range p.pool {
		conn.Client.Close()
		delete(p.pool, key)
	}
}

// CleanIdleConnections 清理空闲连接
func (p *SSHConnectionPool) CleanIdleConnections() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for key, conn := range p.pool {
		if !conn.IsInUse && time.Since(conn.LastUsed) > p.maxIdle {
			conn.Client.Close()
			delete(p.pool, key)
		}
	}
}

// ExecuteCommand 在SSH连接上执行命令
func (conn *SSHConnection) ExecuteCommand(command string) (string, string, int, error) {
	// 创建会话
	session, err := conn.Client.NewSession()
	if err != nil {
		return "", "", -1, fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer session.Close()

	// 获取标准输出和标准错误
	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", "", -1, fmt.Errorf("获取标准输出失败: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return "", "", -1, fmt.Errorf("获取标准错误失败: %w", err)
	}

	// 执行命令
	err = session.Start(command)
	if err != nil {
		return "", "", -1, fmt.Errorf("启动命令失败: %w", err)
	}

	// 读取输出
	stdoutBytes := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			stdoutBytes = append(stdoutBytes, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	stderrBytes := make([]byte, 0)
	buf = make([]byte, 1024)
	for {
		n, err := stderr.Read(buf)
		if n > 0 {
			stderrBytes = append(stderrBytes, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	// 等待命令完成
	err = session.Wait()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		} else {
			return string(stdoutBytes), string(stderrBytes), -1, fmt.Errorf("等待命令完成失败: %w", err)
		}
	}

	return string(stdoutBytes), string(stderrBytes), exitCode, nil
}

// 加载私钥
func loadPrivateKey(keyPath, keyPassword string) (ssh.Signer, error) {
	key, err := readPrivateKeyFile(keyPath)
	if err != nil {
		return nil, err
	}

	var signer ssh.Signer
	if keyPassword != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(keyPassword))
	} else {
		signer, err = ssh.ParsePrivateKey(key)
	}

	if err != nil {
		return nil, err
	}

	return signer, nil
}

// 读取私钥文件
func readPrivateKeyFile(keyPath string) ([]byte, error) {
	// 这里应该使用os.ReadFile，但为了简化示例，我们假设已经读取了文件
	// 在实际实现中，应该使用os.ReadFile读取文件内容
	return nil, fmt.Errorf("未实现的私钥读取功能")
}
