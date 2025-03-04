package executors

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/ape902/ansible-go/pkg/executor/connection"
	"github.com/ape902/ansible-go/pkg/executor/models"
	"github.com/ape902/ansible-go/pkg/vars"
)

// FetchExecutor 获取文件执行器
type FetchExecutor struct{}

// NewFetchExecutor 创建新的获取文件执行器
func NewFetchExecutor() *FetchExecutor {
	return &FetchExecutor{}
}

// Execute 执行获取文件任务
func (e *FetchExecutor) Execute(ctx context.Context, task *models.Task, conn connection.Connection, varStore *vars.Store) (*models.TaskResult, error) {
	// 检查任务参数
	src, ok := task.Spec.Args["src"]
	if !ok {
		return nil, fmt.Errorf("fetch模块必须提供src参数")
	}

	srcStr, ok := src.(string)
	if !ok {
		return nil, fmt.Errorf("src参数必须是字符串类型")
	}

	dest, ok := task.Spec.Args["dest"]
	if !ok {
		return nil, fmt.Errorf("fetch模块必须提供dest参数")
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

	// 构建目标路径，可能包含主机名
	flatDest := destStr
	if flat, ok := task.Spec.Args["flat"]; ok {
		if flatBool, ok := flat.(bool); ok && !flatBool {
			// 如果flat为false，则在目标路径中添加主机名
			flatDest = filepath.Join(destStr, task.Host, srcStr)
		}
	}

	// 从远程主机获取文件
	err := conn.FetchFile(srcStr, flatDest)
	if err != nil {
		return nil, fmt.Errorf("获取文件失败: %w", err)
	}

	// 计算执行时间
	duration := time.Since(startTime)

	// 创建任务结果
	taskResult := &models.TaskResult{
		ExitCode: 0,
		Stdout:   fmt.Sprintf("文件 %s 已成功获取到 %s", srcStr, flatDest),
		Stderr:   "",
		Changed:  true,
		Failed:   false,
		Duration: duration,
		Extra:    make(map[string]string),
	}

	// 添加额外信息
	taskResult.Extra["src"] = srcStr
	taskResult.Extra["dest"] = flatDest

	return taskResult, nil
}
