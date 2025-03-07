package connection

import (
	"fmt"
	"io/ioutil"
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
	Password string
	LastUsed time.Time
	IsInUse  bool
}

// NewSSHConnection 创建新的SSH连接
func NewSSHConnection(host string, port int) Connection {
	return &SSHConnection{
		Host: host,
		Port: port,
		User: "",
		Password: "",
		IsInUse: false,
	}
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

	// 使用全局SSH配置中的用户名和密码（如果未提供特定值）
	if user == "" {
		user = p.sshConfig.User
	}
	if password == "" {
		password = p.sshConfig.Password
	}

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

	// 根据UseKeyAuth字段选择认证方式
	if p.sshConfig.UseKeyAuth {
		// 使用密钥认证
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
	} else {
		// 使用密码认证
		if password != "" {
			authMethods = append(authMethods, ssh.Password(password))
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
		HostKeyCallback: getHostKeyCallback(p.sshConfig),
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
		Password: password,
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

// Connect 实现Connection接口的Connect方法
func (conn *SSHConnection) Connect() error {
	// 从连接实例获取认证信息
	config := &ssh.ClientConfig{
		User:            conn.User,
		Auth:            []ssh.AuthMethod{ssh.Password(conn.Password)}, // 使用连接实例中的密码
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 始终禁用主机密钥检查，避免首次连接时的交互确认
		Timeout:         time.Second * 10,
	}

	// 建立连接
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", conn.Host, conn.Port), config)
	if err != nil {
		return fmt.Errorf("SSH连接失败: %w", err)
	}

	conn.Client = client
	return nil
}

// Disconnect 实现Connection接口的Disconnect方法
func (conn *SSHConnection) Disconnect() error {
	if conn.Client != nil {
		return conn.Client.Close()
	}
	return nil
}

// IsConnected 实现Connection接口的IsConnected方法
func (conn *SSHConnection) IsConnected() bool {
	return conn.Client != nil
}

// GetType 实现Connection接口的GetType方法
func (conn *SSHConnection) GetType() ConnectionType {
	return ConnectionTypeSSH
}

// ExecuteCommand 实现Connection接口的ExecuteCommand方法
func (conn *SSHConnection) ExecuteCommand(command string) (*ConnectionResult, error) {
	// 创建会话
	session, err := conn.Client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer session.Close()

	// 获取标准输出和标准错误
	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("获取标准输出失败: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("获取标准错误失败: %w", err)
	}

	// 执行命令
	err = session.Start(command)
	if err != nil {
		return nil, fmt.Errorf("启动命令失败: %w", err)
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
	start := time.Now()
	err = session.Wait()
	duration := time.Since(start)
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		} else {
			return nil, fmt.Errorf("等待命令完成失败: %w", err)
		}
	}

	return &ConnectionResult{
		Stdout:   string(stdoutBytes),
		Stderr:   string(stderrBytes),
		ExitCode: exitCode,
		Duration: duration,
	}, nil
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
	return ioutil.ReadFile(keyPath)
}

// getHostKeyCallback 获取主机密钥验证回调函数
func getHostKeyCallback(config *types.SSHConfig) ssh.HostKeyCallback {
	// 如果配置了禁用主机密钥检查，直接返回InsecureIgnoreHostKey
	if config != nil && config.DisableHostKeyChecking {
		return ssh.InsecureIgnoreHostKey()
	}

	// 默认使用InsecureIgnoreHostKey，以避免首次连接时的交互确认
	// 在生产环境中，应该实现更安全的主机密钥验证机制
	return ssh.InsecureIgnoreHostKey()
}

// CopyFile 实现Connection接口的CopyFile方法
func (conn *SSHConnection) CopyFile(localPath, remotePath string) error {
	// 读取本地文件
	content, err := ioutil.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("读取本地文件失败: %w", err)
	}

	// 创建远程会话
	session, err := conn.Client.NewSession()
	if err != nil {
		return fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer session.Close()

	// 创建远程文件
	cmd := fmt.Sprintf("cat > %s", remotePath)
	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("获取标准输入失败: %w", err)
	}

	// 启动命令
	err = session.Start(cmd)
	if err != nil {
		return fmt.Errorf("启动命令失败: %w", err)
	}

	// 写入文件内容
	_, err = stdin.Write(content)
	if err != nil {
		return fmt.Errorf("写入文件内容失败: %w", err)
	}
	stdin.Close()

	// 等待命令完成
	err = session.Wait()
	if err != nil {
		return fmt.Errorf("等待命令完成失败: %w", err)
	}

	return nil
}

// FetchFile 实现Connection接口的FetchFile方法
func (conn *SSHConnection) FetchFile(remotePath, localPath string) error {
	// 创建远程会话
	session, err := conn.Client.NewSession()
	if err != nil {
		return fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer session.Close()

	// 获取标准输出
	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("获取标准输出失败: %w", err)
	}

	// 启动命令
	cmd := fmt.Sprintf("cat %s", remotePath)
	err = session.Start(cmd)
	if err != nil {
		return fmt.Errorf("启动命令失败: %w", err)
	}

	// 读取文件内容
	content, err := ioutil.ReadAll(stdout)
	if err != nil {
		return fmt.Errorf("读取文件内容失败: %w", err)
	}

	// 等待命令完成
	err = session.Wait()
	if err != nil {
		return fmt.Errorf("等待命令完成失败: %w", err)
	}

	// 写入本地文件
	err = ioutil.WriteFile(localPath, content, 0644)
	if err != nil {
		return fmt.Errorf("写入本地文件失败: %w", err)
	}

	return nil
}
