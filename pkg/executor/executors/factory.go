package executors

import (
	"fmt"
	"sync"

	"github.com/ape902/ansible-go/pkg/executor/engine"
)

// ExecutorFactory 实现执行器工厂
type ExecutorFactory struct {
	executors map[string]engine.TaskExecutor
	mutex     sync.RWMutex
}

// NewExecutorFactory 创建新的执行器工厂
func NewExecutorFactory() *ExecutorFactory {
	factory := &ExecutorFactory{
		executors: make(map[string]engine.TaskExecutor),
	}

	// 注册内置执行器
	factory.RegisterExecutor("command", NewCommandExecutor())
	factory.RegisterExecutor("shell", NewShellExecutor())
	factory.RegisterExecutor("file", NewFileExecutor())
	factory.RegisterExecutor("template", NewTemplateExecutor())
	factory.RegisterExecutor("copy", NewCopyExecutor())
	factory.RegisterExecutor("fetch", NewFetchExecutor())

	return factory
}

// RegisterExecutor 注册执行器
func (f *ExecutorFactory) RegisterExecutor(name string, executor engine.TaskExecutor) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.executors[name] = executor
}

// CreateExecutor 创建执行器
func (f *ExecutorFactory) CreateExecutor(taskType string) (engine.TaskExecutor, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	executor, exists := f.executors[taskType]
	if !exists {
		return nil, fmt.Errorf("未知的任务类型: %s", taskType)
	}

	return executor, nil
}

// GetSupportedExecutors 获取支持的执行器列表
func (f *ExecutorFactory) GetSupportedExecutors() []string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	executors := make([]string, 0, len(f.executors))
	for name := range f.executors {
		executors = append(executors, name)
	}

	return executors
}