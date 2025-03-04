package executors

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/ape902/ansible-go/pkg/executor/connection"
	"github.com/ape902/ansible-go/pkg/executor/models"
	"github.com/ape902/ansible-go/pkg/vars"
)

// TemplateExecutor 模板执行器
type TemplateExecutor struct{}

// NewTemplateExecutor 创建新的模板执行器
func NewTemplateExecutor() *TemplateExecutor {
	return &TemplateExecutor{}
}

// Execute 执行模板任务
func (e *TemplateExecutor) Execute(ctx context.Context, task *models.Task, conn connection.Connection, varStore *vars.Store) (*models.TaskResult, error) {
	// 检查任务参数
	src, ok := task.Spec.Args["src"]
	if !ok {
		return nil, fmt.Errorf("template模块必须提供src参数")
	}

	srcStr, ok := src.(string)
	if !ok {
		return nil, fmt.Errorf("src参数必须是字符串类型")
	}

	dest, ok := task.Spec.Args["dest"]
	if !ok {
		return nil, fmt.Errorf("template模块必须提供dest参数")
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

	// 读取模板文件
	tmplContent, err := ioutil.ReadFile(srcStr)
	if err != nil {
		return nil, fmt.Errorf("读取模板文件失败: %w", err)
	}

	// 解析模板
	tmpl, err := template.New(filepath.Base(srcStr)).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("解析模板失败: %w", err)
	}

	// 合并变量
	vars := make(map[string]interface{})
	for k, v := range varStore.GetAll() {
		vars[k] = v
	}
	for k, v := range task.Vars {
		vars[k] = v
	}

	// 渲染模板
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, vars)
	if err != nil {
		return nil, fmt.Errorf("渲染模板失败: %w", err)
	}

	// 创建临时文件
	tmpFile, err := ioutil.TempFile("", "ansible-go-template-*")
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// 写入渲染后的内容
	_, err = tmpFile.Write(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("写入临时文件失败: %w", err)
	}
	tmpFile.Close()

	// 复制文件到远程主机
	err = conn.CopyFile(tmpFile.Name(), destStr)
	if err != nil {
		return nil, fmt.Errorf("复制文件到远程主机失败: %w", err)
	}

	// 处理权限设置
	var cmdStr string
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
		cmdStr = fmt.Sprintf("chmod %s %s", modeStr, destStr)
		
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
		Stdout:   fmt.Sprintf("模板 %s 已成功部署到 %s", srcStr, destStr),
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