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

// CommandExecutor 命令执行器
type CommandExecutor struct{}

// NewCommandExecutor 创建新的命令执行器
func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{}
}

// Execute 执行命令任务
func (e *CommandExecutor) Execute(ctx context.Context, task *models.Task, conn connection.Connection, varStore *vars.Store) (*models.TaskResult, error) {
	// 检查任务参数
	cmd, ok := task.Spec.Args["cmd"]
	if !ok {
		return nil, fmt.Errorf("command模块必须提供cmd参数")
	}

	cmdStr, ok := cmd.(string)
	if !ok {
		return nil, fmt.Errorf("cmd参数必须是字符串类型")
	}

	// 替换变量
	cmdStr = replaceVars(cmdStr, task.Vars, varStore)

	// 记录开始时间
	startTime := time.Now()

	// 执行命令
	result, err := conn.ExecuteCommand(cmdStr)
	if err != nil {
		return nil, fmt.Errorf("执行命令失败: %w", err)
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
	taskResult.Extra["command"] = cmdStr

	return taskResult, nil
}

// replaceVars 替换命令中的变量
func replaceVars(cmd string, taskVars map[string]interface{}, varStore *vars.Store) string {
	// 先使用任务变量替换
	for k, v := range taskVars {
		placeholder := fmt.Sprintf("{{%s}}", k)
		value := fmt.Sprintf("%v", v)
		cmd = strings.ReplaceAll(cmd, placeholder, value)
	}

	// 再使用全局变量替换
	globalVars := varStore.GetAll()
	for k, v := range globalVars {
		placeholder := fmt.Sprintf("{{%s}}", k)
		value := fmt.Sprintf("%v", v)
		cmd = strings.ReplaceAll(cmd, placeholder, value)
	}

	return cmd
}