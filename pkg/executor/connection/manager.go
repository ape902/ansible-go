package connection

import (
	"fmt"
	"sync"
	"github.com/ape902/ansible-go/pkg/config/types"
)

// ConnectionManagerImpl 实现连接管理器
type ConnectionManagerImpl struct {
	pool      *Pool
	mutex     sync.RWMutex
	sshConfig *types.SSHConfig
}

// NewConnectionManager 创建新的连接管理器
func NewConnectionManager(pool *Pool, sshConfig *types.SSHConfig) ConnectionManager {
	return &ConnectionManagerImpl{
		pool:      pool,
		sshConfig: sshConfig,
	}
}

// GetConnection 获取指定主机的连接
func (m *ConnectionManagerImpl) GetConnection(host string, port int, connType ConnectionType) (Connection, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 尝试从连接池获取连接
	conn, err := m.pool.Get(host)
	if err != nil {
		return nil, fmt.Errorf("从连接池获取连接失败: %w", err)
	}

	// 如果连接不存在，创建新连接
	if conn == nil {
		// 根据连接类型创建不同的连接
		switch connType {
		case ConnectionTypeSSH:
			conn = NewSSHConnection(host, port)
			// 设置SSH连接的用户名和密码
			if sshConn, ok := conn.(*SSHConnection); ok {
				// 使用配置文件中的用户名和密码
				sshConn.User = m.sshConfig.User
				sshConn.Password = m.sshConfig.Password
			}
		case ConnectionTypeLocal:
			conn = NewLocalConnection()
		default:
			return nil, fmt.Errorf("不支持的连接类型: %s", connType)
		}

		// 建立连接
		err = conn.Connect()
		if err != nil {
			return nil, fmt.Errorf("建立连接失败: %w", err)
		}

		// 添加到连接池
		m.pool.Add(host, conn)
	}

	return conn, nil
}

// ReleaseConnection 释放连接
func (m *ConnectionManagerImpl) ReleaseConnection(conn Connection) {
	// 当前简单实现，保持连接在池中
}

// CloseConnection 关闭指定连接
func (m *ConnectionManagerImpl) CloseConnection(host string, port int, connType ConnectionType) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 从连接池获取连接
	conn, err := m.pool.Get(host)
	if err != nil {
		return fmt.Errorf("从连接池获取连接失败: %w", err)
	}

	if conn != nil {
		// 断开连接
		err = conn.Disconnect()
		if err != nil {
			return fmt.Errorf("断开连接失败: %w", err)
		}

		// 从连接池移除
		m.pool.Remove(host)
	}

	return nil
}

// CloseAll 关闭所有连接
func (m *ConnectionManagerImpl) CloseAll() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 遍历所有连接并关闭
	for host, conn := range m.pool.connections {
		conn.Disconnect()
		m.pool.Remove(host)
	}
}

// CleanIdleConnections 清理空闲连接
func (m *ConnectionManagerImpl) CleanIdleConnections() {
	// 当前简单实现，不做任何操作
}