package executors

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/ape902/ansible-go/pkg/config"
	"github.com/ape902/ansible-go/pkg/executor/connection"
	"github.com/ape902/ansible-go/pkg/executor/engine"
	"github.com/ape902/ansible-go/pkg/executor/models"
	"github.com/ape902/ansible-go/pkg/vars"
)

// ImportExecutor 导入任务执行器
type ImportExecutor struct{}

// NewImportExecutor 创建新的导入任务执行器
func NewImportExecutor() *ImportExecutor {
	return &ImportExecutor{}
}

// Execute 执行导入任务
func (e *ImportExecutor) Execute(ctx context.Context, task *models.Task, conn connection.Connection, varStore *vars.Store) (*models.TaskResult, error) {
	// 检查任务参数
	file, ok := task.Spec.Args["file"]
	if !ok {
		return nil, fmt.Errorf("import模块必须提供file参数")
	}

	fileStr, ok := file.(string)
	if !ok {
		return nil, fmt.Errorf("file参数必须是字符串类型")
	}

	// 记录开始时间
	startTime := time.Now()

	// 获取当前任务文件的目录路径
	taskDir := filepath.Dir(task.FilePath)
	// 构建导入文件的完整路径
	importFilePath := filepath.Join(taskDir, fileStr)

	// 加载导入的任务文件
	taskConfig, err := config.LoadPlaybook(importFilePath)
	if err != nil {
		return nil, fmt.Errorf("加载导入任务文件失败: %w", err)
	}

	// 获取任务上下文
	taskCtx, ok := ctx.Value("taskContext").(*models.TaskContext)
	if !ok || taskCtx == nil {
		return nil, fmt.Errorf("无法获取任务上下文")
	}

	// 获取执行引擎
	engine, ok := ctx.Value("engine").(*engine.ExecutionEngine)
	if !ok || engine == nil {
		return nil, fmt.Errorf("无法获取执行引擎")
	}

	// 将导入的任务添加到执行队列
	for _, taskSpec := range taskConfig.Tasks {
		for taskName, spec := range taskSpec {
			// 为每个主机创建任务
			for _, host := range taskCtx.Hosts {
				importedTask := &models.Task{
					ID:       taskName,
					Spec:     &spec,
					Status:   models.TaskStatusPending,
					Priority: models.TaskPriorityNormal,
					Host:     host,
					Vars:     make(map[string]interface{}),
					FilePath: importFilePath,
				}
				// 添加任务到队列
				err := engine.AddTask(importedTask)
				if err != nil {
					return nil, fmt.Errorf("添加导入任务到队列失败: %w", err)
				}
			}
		}
	}

	// 计算执行时间
	duration := time.Since(startTime)

	// 创建任务结果
	result := &models.TaskResult{
		ExitCode: 0,
		Stdout:   fmt.Sprintf("成功导入任务文件: %s，并添加了 %d 个任务到执行队列", fileStr, len(taskConfig.Tasks)),
		Stderr:   "",
		Duration: duration,
		Changed:  true,
		ImportedTasks: taskConfig.Tasks,
	}

	return result, nil
}