package models

import (
	"time"

	"github.com/ape902/ansible-go/pkg/config/types"
)

// TaskStatus 定义任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 等待执行
	TaskStatusRunning   TaskStatus = "running"   // 正在执行
	TaskStatusSuccess   TaskStatus = "success"   // 执行成功
	TaskStatusFailed    TaskStatus = "failed"    // 执行失败
	TaskStatusCancelled TaskStatus = "cancelled" // 已取消
	TaskStatusSkipped   TaskStatus = "skipped"   // 已跳过
)

// TaskPriority 定义任务优先级
type TaskPriority int

const (
	TaskPriorityLow    TaskPriority = 0
	TaskPriorityNormal TaskPriority = 1
	TaskPriorityHigh   TaskPriority = 2
)

// Task 定义任务结构
type Task struct {
	ID          string                 // 任务ID
	Spec        *types.TaskSpec        // 任务规格
	Status      TaskStatus             // 任务状态
	Priority    TaskPriority           // 任务优先级
	DependsOn   []string               // 依赖任务ID列表
	Host        string                 // 目标主机
	Vars        map[string]interface{} // 任务变量
	RetryCount  int                    // 已重试次数
	StartTime   *time.Time            // 开始时间
	EndTime     *time.Time            // 结束时间
	Result      *TaskResult           // 执行结果
	Error       error                 // 错误信息
}

// TaskResult 定义任务执行结果
type TaskResult struct {
	ExitCode    int               // 退出码
	Stdout      string            // 标准输出
	Stderr      string            // 标准错误
	Changed     bool              // 是否发生变更
	Failed      bool              // 是否失败
	Skipped     bool              // 是否跳过
	Unreachable bool              // 是否不可达
	Duration    time.Duration     // 执行时长
	Extra       map[string]string // 额外信息
}

// TaskQueue 定义任务队列接口
type TaskQueue interface {
	Push(task *Task) error           // 添加任务
	Pop() (*Task, error)             // 获取任务
	Remove(taskID string) error      // 移除任务
	Get(taskID string) (*Task, bool) // 获取指定任务
	List() []*Task                   // 获取所有任务
	Len() int                        // 获取任务数量
}