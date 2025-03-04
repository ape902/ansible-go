package types

// TaskStatus 定义任务执行状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 任务等待执行
	TaskStatusRunning   TaskStatus = "running"   // 任务正在执行
	TaskStatusSuccess   TaskStatus = "success"   // 任务执行成功
	TaskStatusFailed    TaskStatus = "failed"    // 任务执行失败
	TaskStatusSkipped   TaskStatus = "skipped"   // 任务被跳过
	TaskStatusCancelled TaskStatus = "cancelled" // 任务被取消
)

// TaskResult 定义任务执行结果
type TaskResult struct {
	TaskName    string                 `json:"task_name"`              // 任务名称
	Status      TaskStatus             `json:"status"`                 // 任务状态
	StartTime   int64                  `json:"start_time,omitempty"`   // 开始时间（Unix时间戳）
	EndTime     int64                  `json:"end_time,omitempty"`     // 结束时间（Unix时间戳）
	Host        string                 `json:"host"`                   // 执行主机
	Module      string                 `json:"module"`                 // 使用的模块
	Args        map[string]interface{} `json:"args,omitempty"`         // 模块参数
	Output      string                 `json:"output,omitempty"`       // 执行输出
	Error       string                 `json:"error,omitempty"`        // 错误信息
	RetryCount  int                    `json:"retry_count,omitempty"`  // 重试次数
	Changed     bool                   `json:"changed"`                // 是否发生变更
}

// TaskResultSummary 定义任务执行结果汇总
type TaskResultSummary struct {
	TotalTasks     int            `json:"total_tasks"`      // 总任务数
	SuccessfulTasks int            `json:"successful_tasks"` // 成功任务数
	FailedTasks    int            `json:"failed_tasks"`     // 失败任务数
	SkippedTasks   int            `json:"skipped_tasks"`    // 跳过任务数
	Results        []*TaskResult  `json:"results"`          // 详细结果列表
	StartTime      int64          `json:"start_time"`       // 开始时间（Unix时间戳）
	EndTime        int64          `json:"end_time"`         // 结束时间（Unix时间戳）
	Duration       int64          `json:"duration"`         // 执行时长（秒）
}