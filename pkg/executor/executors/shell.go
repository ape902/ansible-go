package executors

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ape902/ansible-go/pkg/executor/connection"
	"github.com/ape902/ansible-go/pkg/executor/models"
	"github.com/ape902/ansible-go/pkg/vars"
)

// ShellExecutor Shell执行器
type ShellExecutor struct{}

// NewShellExecutor 创建新的Shell执行器
func NewShellExecutor() *ShellExecutor {
	return &ShellExecutor{}
}

// Execute 执行Shell任务
func (e *ShellExecutor) Execute(ctx context.Context, task *models.Task, conn connection.Connection, varStore *vars.Store) (*models.TaskResult, error) {
	// 检查任务参数
	script, ok := task.Spec.Args["script"]
	if !ok {
		return nil, fmt.Errorf("shell模块必须提供script参数")
	}

	scriptStr, ok := script.(string)
	if !ok {
		return nil, fmt.Errorf("script参数必须是字符串类型")
	}

	// 获取shell类型，默认为/bin/sh
	shellType := "/bin/sh"
	if shellArg, ok := task.Spec.Args["shell"]; ok {
		if shellStr, ok := shellArg.(string); ok && shellStr != "" {
			shellType = shellStr
		}
	}

	// 替换变量
	scriptStr = replaceVars(scriptStr, task.Vars, varStore)

	// 构建完整命令
	cmdStr := fmt.Sprintf("%s -c '%s'", shellType, escapeQuotes(scriptStr))

	// 记录开始时间
	startTime := time.Now()

	// 执行命令
	result, err := conn.ExecuteCommand(cmdStr)
	if err != nil {
		return nil, fmt.Errorf("执行Shell脚本失败: %w", err)
	}

	// 计算执行时间
	duration := time.Since(startTime)

	// 创建任务结果
	taskResult := &models.TaskResult{
		ExitCode: result.ExitCode,
		Stdout:   result.Stdout,
		Stderr:   result.Stderr,
		Changed:  result.ExitCode == 0 && (result.Stdout != "" || result.Stderr != ""),
		Failed:   result.ExitCode != 0,
		Duration: duration,
		Extra:    make(map[string]string),
	}

	// 添加额外信息
	taskResult.Extra["shell"] = shellType
	taskResult.Extra["script"] = scriptStr

	return taskResult, nil
}

// escapeQuotes 转义引号
func escapeQuotes(script string) string {
	// 将单引号替换为 '\'' 以在shell中正确处理
	return strings.ReplaceAll(script, "'", "'\\''")
}