package connection

// Pool 定义连接池接口
type Pool struct {
	connections map[string]Connection
}



// NewPool 创建新的连接池
func NewPool() *Pool {
	return &Pool{
		connections: make(map[string]Connection),
	}
}

// Get 获取连接
func (p *Pool) Get(host string) (Connection, error) {
	if conn, ok := p.connections[host]; ok {
		return conn, nil
	}
	return nil, nil
}

// Add 添加连接
func (p *Pool) Add(host string, conn Connection) {
	p.connections[host] = conn
}

// Remove 移除连接
func (p *Pool) Remove(host string) {
	delete(p.connections, host)
}