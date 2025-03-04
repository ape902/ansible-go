package executors

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ape902/ansible-go/pkg/executor/connection"
	"github.com/ape902/ansible-go/pkg/executor/models"
	"github.com/ape902/ansible-go/pkg/vars"
)

// FileExecutor 文件执行器
type FileExecutor struct{}

// NewFileExecutor 创建新的文件执行器
func NewFileExecutor() *FileExecutor {
	return &FileExecutor{}
}

// Execute 执行文件任务
func (e *FileExecutor) Execute(ctx context.Context, task *models.Task, conn connection.Connection, varStore *vars.Store) (*models.TaskResult, error) {
	// 检查任务参数
	path, ok := task.Spec.Args["path"]
	if !ok {
		return nil, fmt.Errorf("file模块必须提供path参数")
	}

	pathStr, ok := path.(string)
	if !ok {
		return nil, fmt.Errorf("path参数必须是字符串类型")
	}

	// 替换变量
	pathStr = replaceVars(pathStr, task.Vars, varStore)

	// 获取操作类型，默认为file
	state := "file"
	if stateArg, ok := task.Spec.Args["state"]; ok {
		if stateStr, ok := stateArg.(string); ok && stateStr != "" {
			state = stateStr
		}
	}

	// 记录开始时间
	startTime := time.Now()

	// 根据操作类型执行不同的命令
	var cmdStr string
	var changed bool

	switch state {
	case "absent":
		// 删除文件或目录
		cmdStr = fmt.Sprintf("rm -rf %s", pathStr)
		changed = true
	case "directory":
		// 创建目录
		cmdStr = fmt.Sprintf("mkdir -p %s", pathStr)
		changed = true
	case "touch":
		// 创建空文件
		cmdStr = fmt.Sprintf("touch %s", pathStr)
		changed = true
	case "file":
		// 确保文件存在
		cmdStr = fmt.Sprintf("test -f %s || touch %s", pathStr, pathStr)
		changed = true
	default:
		return nil, fmt.Errorf("不支持的state类型: %s", state)
	}

	// 处理权限设置
	if mode, ok := task.Spec.Args["mode"]; ok {
		var modeStr string
		switch v := mode.(type) {
		case string:
			modeStr = v
		case int:
			modeStr = strconv.FormatInt(int64(v), 8)
		default:
			return nil, fmt.Errorf("mode参数必须是字符串或整数类型")
		}
		cmdStr = fmt.Sprintf("%s && chmod %s %s", cmdStr, modeStr, pathStr)
		changed = true
	}

	// 处理所有者设置
	if owner, ok := task.Spec.Args["owner"]; ok {
		if ownerStr, ok := owner.(string); ok && ownerStr != "" {
			cmdStr = fmt.Sprintf("%s && chown %s %s", cmdStr, ownerStr, pathStr)
			changed = true
		}
	}

	// 处理组设置
	if group, ok := task.Spec.Args["group"]; ok {
		if groupStr, ok := group.(string); ok && groupStr != "" {
			cmdStr = fmt.Sprintf("%s && chgrp %s %s", cmdStr, groupStr, pathStr)
			changed = true
		}
	}

	// 执行命令
	result, err := conn.ExecuteCommand(cmdStr)
	if err != nil {
		return nil, fmt.Errorf("执行文件操作失败: %w", err)
	}

	// 计算执行时间
	duration := time.Since(startTime)

	// 创建任务结果
	taskResult := &models.TaskResult{
		ExitCode: result.ExitCode,
		Stdout:   result.Stdout,
		Stderr:   result.Stderr,
		Changed:  changed && result.ExitCode == 0,
		Failed:   result.ExitCode != 0,
		Duration: duration,
		Extra:    make(map[string]string),
	}

	// 添加额外信息
	taskResult.Extra["path"] = pathStr
	taskResult.Extra["state"] = state
	taskResult.Extra["command"] = cmdStr

	return taskResult, nil
}
