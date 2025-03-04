package executors

import (
	"context"
	"fmt"
	"time"

	"github.com/ape902/ansible-go/pkg/executor/connection"
	"github.com/ape902/ansible-go/pkg/executor/models"
	"github.com/ape902/ansible-go/pkg/vars"
)

// CopyExecutor 复制执行器
type CopyExecutor struct{}

// NewCopyExecutor 创建新的复制执行器
func NewCopyExecutor() *CopyExecutor {
	return &CopyExecutor{}
}

// Execute 执行复制任务
func (e *CopyExecutor) Execute(ctx context.Context, task *models.Task, conn connection.Connection, varStore *vars.Store) (*models.TaskResult, error) {
	// 检查任务参数
	src, ok := task.Spec.Args["src"]
	if !ok {
		return nil, fmt.Errorf("copy模块必须提供src参数")
	}

	srcStr, ok := src.(string)
	if !ok {
		return nil, fmt.Errorf("src参数必须是字符串类型")
	}

	dest, ok := task.Spec.Args["dest"]
	if !ok {
		return nil, fmt.Errorf("copy模块必须提供dest参数")
	}

	destStr, ok := dest.(string)
	if !ok {
		return nil, fmt.Errorf("dest参数必须是字符串类型")
	}

	// 替换变量
	srcStr = replaceVars(srcStr, task.Vars, varStore)
	destStr = replaceVars(destStr, task.Vars, varStore)

	// 记录开始时间
	startTime := time.Now()

	// 复制文件到远程主机
	err := conn.CopyFile(srcStr, destStr)
	if err != nil {
		return nil, fmt.Errorf("复制文件失败: %w", err)
	}

	// 处理权限设置
	if mode, ok := task.Spec.Args["mode"]; ok {
		var modeStr string
		switch v := mode.(type) {
		case string:
			modeStr = v
		case int:
			modeStr = fmt.Sprintf("%o", v)
		default:
			return nil, fmt.Errorf("mode参数必须是字符串或整数类型")
		}
		cmdStr := fmt.Sprintf("chmod %s %s", modeStr, destStr)

		// 执行命令
		result, err := conn.ExecuteCommand(cmdStr)
		if err != nil {
			return nil, fmt.Errorf("设置文件权限失败: %w", err)
		}

		if result.ExitCode != 0 {
			return nil, fmt.Errorf("设置文件权限失败: %s", result.Stderr)
		}
	}

	// 计算执行时间
	duration := time.Since(startTime)

	// 创建任务结果
	taskResult := &models.TaskResult{
		ExitCode: 0,
		Stdout:   fmt.Sprintf("文件 %s 已成功复制到 %s", srcStr, destStr),
		Stderr:   "",
		Changed:  true,
		Failed:   false,
		Duration: duration,
		Extra:    make(map[string]string),
	}

	// 添加额外信息
	taskResult.Extra["src"] = srcStr
	taskResult.Extra["dest"] = destStr

	return taskResult, nil
}
