package models

import (
	"github.com/ape902/ansible-go/pkg/vars"
)

// TaskContext 定义任务执行上下文
type TaskContext struct {
	// 目标主机列表
	Hosts []string

	// 变量存储
	VarStore *vars.Store

	// 任务状态
	Status TaskStatus

	// 任务优先级
	Priority TaskPriority

	// 错误信息
	Error error
}