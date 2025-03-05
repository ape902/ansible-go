package engine

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ape902/ansible-go/pkg/executor/connection"
	"github.com/ape902/ansible-go/pkg/executor/models"
	"github.com/ape902/ansible-go/pkg/vars"
)

// ExecutionMode 定义执行模式
type ExecutionMode int

const (
	// ExecutionModeSerial 串行执行模式
	ExecutionModeSerial ExecutionMode = iota
	// ExecutionModeParallel 并行执行模式
	ExecutionModeParallel
	// ExecutionModeParallelByHost 按主机并行执行模式
	ExecutionModeParallelByHost
)

// ExecutionOptions 定义执行选项
type ExecutionOptions struct {
	// 执行模式
	Mode ExecutionMode
	// 最大并行度
	MaxParallel int
	// 执行超时时间
	Timeout time.Duration
	// 重试次数
	MaxRetries int
	// 重试间隔
	RetryInterval time.Duration
	// 是否忽略错误继续执行
	IgnoreErrors bool
	// 是否启用调试模式
	Debug bool
}

// DefaultExecutionOptions 默认执行选项
var DefaultExecutionOptions = ExecutionOptions{
	Mode:          ExecutionModeParallel,
	MaxParallel:   10,
	Timeout:       30 * time.Minute,
	MaxRetries:    3,
	RetryInterval: 5 * time.Second,
	IgnoreErrors:  false,
	Debug:         false,
}

// ExecutionEngine 定义执行引擎
type ExecutionEngine struct {
	queue           models.TaskQueue
	connManager     connection.ConnectionManager
	varStore        *vars.Store
	options         ExecutionOptions
	executorFactory ExecutorFactory
	running         bool
	mutex           sync.RWMutex
	workers         int
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

// ExecutorFactory 定义执行器工厂接口
type ExecutorFactory interface {
	// CreateExecutor 创建执行器
	CreateExecutor(taskType string) (TaskExecutor, error)
}

// TaskExecutor 定义任务执行器接口
type TaskExecutor interface {
	// Execute 执行任务
	Execute(ctx context.Context, task *models.Task, conn connection.Connection, varStore *vars.Store) (*models.TaskResult, error)
}

// NewExecutionEngine 创建新的执行引擎
func NewExecutionEngine(
	queue models.TaskQueue,
	connManager connection.ConnectionManager,
	varStore *vars.Store,
	executorFactory ExecutorFactory,
	options ExecutionOptions,
) *ExecutionEngine {
	ctx, cancel := context.WithCancel(context.Background())
	return &ExecutionEngine{
		queue:           queue,
		connManager:     connManager,
		varStore:        varStore,
		options:         options,
		executorFactory: executorFactory,
		running:         false,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Start 启动执行引擎
func (e *ExecutionEngine) Start() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.running {
		return fmt.Errorf("执行引擎已经在运行")
	}

	e.running = true
	workers := e.options.MaxParallel
	if workers <= 0 {
		workers = 1
	}
	e.workers = workers

	// 启动工作协程
	for i := 0; i < workers; i++ {
		e.wg.Add(1)
		go e.worker(i)
	}

	return nil
}

// Stop 停止执行引擎
func (e *ExecutionEngine) Stop() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if !e.running {
		return
	}

	e.cancel()
	e.wg.Wait()
	e.running = false
}

// IsRunning 检查执行引擎是否正在运行
func (e *ExecutionEngine) IsRunning() bool {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.running
}

// worker 工作协程
func (e *ExecutionEngine) worker(id int) {
	defer e.wg.Done()

	for {
		select {
		case <-e.ctx.Done():
			return
		default:
			// 尝试获取任务
			task, err := e.queue.Pop()
			if err != nil {
				// 队列为空，等待一段时间再尝试
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// 执行任务
			_ = e.ExecuteTask(task, nil) // 使用ExecuteTask而不是executeTask
		}
	}
}

// ExecuteTask 执行单个任务
func (e *ExecutionEngine) ExecuteTask(task *models.Task, taskCtx *models.TaskContext) error {
	// 更新任务状态为运行中
	task.Status = models.TaskStatusRunning
	startTime := time.Now()
	task.StartTime = &startTime

	// 创建任务上下文，带超时控制
	execCtx, cancel := context.WithTimeout(e.ctx, e.options.Timeout)
	defer cancel()

	// 获取连接
	conn, err := e.connManager.GetConnection(task.Host, 22, connection.ConnectionTypeSSH) // 默认使用SSH连接，端口22
	if err != nil {
		task.Status = models.TaskStatusFailed
		task.Error = fmt.Errorf("获取连接失败: %w", err)
		endTime := time.Now()
		task.EndTime = &endTime
		return err
	}
	defer e.connManager.ReleaseConnection(conn)

	// 获取执行器
	executor, err := e.executorFactory.CreateExecutor(task.Spec.Module)
	if err != nil {
		task.Status = models.TaskStatusFailed
		task.Error = fmt.Errorf("创建执行器失败: %w", err)
		endTime := time.Now()
		task.EndTime = &endTime
		return err
	}

	// 执行任务，支持重试
	var result *models.TaskResult
	for retry := 0; retry <= e.options.MaxRetries; retry++ {
		if retry > 0 {
			// 重试前等待一段时间
			time.Sleep(e.options.RetryInterval)
		}

		task.RetryCount = retry
		result, err = executor.Execute(execCtx, task, conn, e.varStore)
		if err == nil || !shouldRetry(err) {
			break
		}
	}

	// 更新任务状态和结果
	endTime := time.Now()
	task.EndTime = &endTime
	task.Result = result

	if err != nil {
		task.Status = models.TaskStatusFailed
		task.Error = err
	} else if result.Failed {
		task.Status = models.TaskStatusFailed
	} else if result.Skipped {
		task.Status = models.TaskStatusSkipped
	} else {
		task.Status = models.TaskStatusSuccess
		// 添加命令执行结果的输出
		// if result.Stdout != "" {
		// 	log.Printf("命令输出:\n%s\n", result.Stdout)
		// }
		if result.Stderr != "" {
			log.Printf("错误输出:\n%s\n", result.Stderr)
		}
	}

	return nil
}

// shouldRetry 判断是否应该重试
func shouldRetry(err error) bool {
	// 这里可以根据错误类型判断是否应该重试
	// 例如，网络超时、连接重置等错误可以重试
	// 而语法错误、权限错误等不应该重试
	return true // 简化实现，默认所有错误都重试
}

// AddTask 添加任务到队列
func (e *ExecutionEngine) AddTask(task *models.Task) error {
	return e.queue.Push(task)
}

// GetTaskStatus 获取任务状态
func (e *ExecutionEngine) GetTaskStatus(taskID string) (models.TaskStatus, error) {
	task, exists := e.queue.Get(taskID)
	if !exists {
		return "", fmt.Errorf("任务ID %s 不存在", taskID)
	}
	return task.Status, nil
}

// GetTaskResult 获取任务结果
func (e *ExecutionEngine) GetTaskResult(taskID string) (*models.TaskResult, error) {
	task, exists := e.queue.Get(taskID)
	if !exists {
		return nil, fmt.Errorf("任务ID %s 不存在", taskID)
	}
	if task.Result == nil {
		return nil, fmt.Errorf("任务ID %s 尚未完成", taskID)
	}
	return task.Result, nil
}

// SetVerbose 设置详细输出模式
func (e *ExecutionEngine) SetVerbose(verbose bool) {
	e.options.Debug = verbose
}
