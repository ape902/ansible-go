package vars

import (
	"fmt"
	"sync"
)

// Store 定义变量存储结构
type Store struct {
	mutex sync.RWMutex
	vars  map[string]interface{}
}

// NewStore 创建新的变量存储实例
func NewStore() *Store {
	return &Store{
		vars: make(map[string]interface{}),
	}
}

// Set 设置变量值
func (s *Store) Set(key string, value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.vars[key] = value
}

// Get 获取变量值
func (s *Store) Get(key string) (interface{}, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	val, exists := s.vars[key]
	return val, exists
}

// Delete 删除变量
func (s *Store) Delete(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.vars, key)
}

// GetAll 获取所有变量
func (s *Store) GetAll() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make(map[string]interface{}, len(s.vars))
	for k, v := range s.vars {
		result[k] = v
	}
	return result
}

// Merge 合并变量
func (s *Store) Merge(vars map[string]interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for k, v := range vars {
		if _, exists := s.vars[k]; exists {
			return fmt.Errorf("变量 %s 已存在", k)
		}
		s.vars[k] = v
	}
	return nil
}