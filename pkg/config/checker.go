package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ape902/ansible-go/pkg/config/types"
	"github.com/ape902/ansible-go/pkg/logger"
	"gopkg.in/yaml.v3"
)

// LineInfo 存储YAML节点的行号信息
type LineInfo struct {
	Path     []string
	LineNum  int
	NodeType string
}

// ConfigChecker 配置检查器
type ConfigChecker struct {
	log     *logger.Logger
	lineMap map[string]int
}

// NewConfigChecker 创建配置检查器
func NewConfigChecker(log *logger.Logger) *ConfigChecker {
	return &ConfigChecker{
		log:     log,
		lineMap: make(map[string]int),
	}
}

// CheckTaskFile 检查任务文件
func (c *ConfigChecker) CheckTaskFile(filePath string) error {
	c.log.Info("检查任务文件: %s", filePath)

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取任务文件失败: %w", err)
	}

	// 解析YAML并收集行号信息
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return fmt.Errorf("解析任务文件失败: %w", err)
	}

	// 收集行号信息
	c.collectLineInfo(&node, []string{})

	// 解析为任务配置
	var taskConfig types.TaskConfig
	if err := yaml.Unmarshal(data, &taskConfig); err != nil {
		return fmt.Errorf("解析任务配置失败: %w", err)
	}

	// 验证任务配置
	errors := ValidateTaskConfig(&taskConfig)
	if len(errors) > 0 {
		c.log.Error("任务配置验证失败:")
		c.log.IncreaseIndent()
		for _, err := range errors {
			// 查找行号
			lineNum := c.findLineNumber(err.Field)
			if lineNum > 0 {
				c.log.Error("第%d行 - %s: %s", lineNum, err.Field, err.Message)
			} else {
				c.log.Error("%s: %s", err.Field, err.Message)
			}
		}
		c.log.DecreaseIndent()
		return fmt.Errorf("任务配置验证失败")
	}

	c.log.Success("任务配置验证通过: %s", filePath)
	return nil
}

// CheckProjectConfig 检查项目配置
func (c *ConfigChecker) CheckProjectConfig(configPath string) error {
	c.log.Info("检查项目配置: %s", configPath)

	// 读取文件内容
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML并收集行号信息
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 收集行号信息
	c.collectLineInfo(&node, []string{})

	// 获取项目根目录
	projectDir := filepath.Dir(configPath)

	// 检查所有相关目录
	dirs := []string{
		filepath.Join(projectDir, "tasks"),
		filepath.Join(projectDir, "vars"),
		filepath.Join(projectDir, "files"),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); err == nil {
			if err := c.checkYAMLDirectory(dir); err != nil {
				c.log.Error("检查目录失败: %s - %v", dir, err)
				return fmt.Errorf("检查目录失败: %s - %w", dir, err)
			}
		}
	}

	// 加载配置
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 验证配置
	errors := ValidateConfig(cfg)
	if len(errors) > 0 {
		c.log.Error("配置验证失败:")
		c.log.IncreaseIndent()
		for _, err := range errors {
			// 查找行号
			lineNum := c.findLineNumber(err.Field)
			if lineNum > 0 {
				c.log.Error("第%d行 - %s: %s", lineNum, err.Field, err.Message)
			} else {
				c.log.Error("%s: %s", err.Field, err.Message)
			}
		}
		c.log.DecreaseIndent()
		return fmt.Errorf("配置验证失败")
	}

	c.log.Success("配置验证通过")
	return nil
}

// checkYAMLDirectory 递归检查目录下的所有YAML文件
func (c *ConfigChecker) checkYAMLDirectory(dirPath string) error {
	if _, err := os.Stat(dirPath); err != nil {
		return fmt.Errorf("目录不存在: %w", err)
	}

	c.log.Info("检查目录: %s", dirPath)
	c.log.IncreaseIndent()

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 检查YAML文件
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
			// 根据文件所在目录决定检查方式
			if strings.Contains(path, "/tasks/") {
				// 任务文件检查
				if err := c.CheckTaskFile(path); err != nil {
					c.log.Error("检查任务文件失败: %s - %v", info.Name(), err)
					return fmt.Errorf("检查任务文件失败: %s - %w", info.Name(), err)
				}
			} else if strings.Contains(path, "/vars/") {
				// 变量文件只检查YAML语法
				if err := c.checkVarsYAMLFile(path); err != nil {
					c.log.Error("检查变量文件失败: %s - %v", info.Name(), err)
					return fmt.Errorf("检查变量文件失败: %s - %w", info.Name(), err)
				}
			} else {
				// 通用YAML文件检查
				if err := c.checkGenericYAMLFile(path); err != nil {
					c.log.Error("检查YAML文件失败: %s - %v", info.Name(), err)
					return fmt.Errorf("检查YAML文件失败: %s - %w", info.Name(), err)
				}
			}
		}

		return nil
	})

	c.log.DecreaseIndent()
	return err
}

// checkVarsYAMLFile 检查变量文件的YAML语法
func (c *ConfigChecker) checkVarsYAMLFile(filePath string) error {
	c.log.Info("检查变量文件: %s", filePath)

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取变量文件失败: %w", err)
	}

	// 解析YAML并收集行号信息
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return fmt.Errorf("解析变量文件失败: %w", err)
	}

	// 收集行号信息
	c.collectLineInfo(&node, []string{})

	// 对于变量文件，只检查YAML语法正确性
	c.log.Success("变量文件语法检查通过")
	return nil
}

// checkGenericYAMLFile 检查通用YAML文件的语法和数据结构
func (c *ConfigChecker) checkGenericYAMLFile(filePath string) error {
	c.log.Info("检查YAML文件: %s", filePath)

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取YAML文件失败: %w", err)
	}

	// 解析YAML并收集行号信息
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return fmt.Errorf("解析YAML文件失败: %w", err)
	}

	// 收集行号信息
	c.collectLineInfo(&node, []string{})

	// 尝试解析为任务配置
	var taskConfig types.TaskConfig
	if err := yaml.Unmarshal(data, &taskConfig); err != nil {
		c.log.Warning("文件不是有效的任务配置: %v", err)
		// 尝试解析为处理器配置
		var handlerSpec types.HandlerSpec
		if err := yaml.Unmarshal(data, &handlerSpec); err != nil {
			c.log.Warning("文件不是有效的处理器配置: %v", err)
			// 如果都不是，则只验证YAML语法
			c.log.Success("YAML文件语法检查通过")
			return nil
		}
		// 验证处理器配置
		if handlerSpec.Name == "" || handlerSpec.Module == "" {
			return fmt.Errorf("处理器配置无效: 名称和模块不能为空")
		}
		c.log.Success("处理器配置验证通过")
		return nil
	}

	// 验证任务配置
	errors := ValidateTaskConfig(&taskConfig)
	if len(errors) > 0 {
		c.log.Error("任务配置验证失败:")
		c.log.IncreaseIndent()
		for _, err := range errors {
			lineNum := c.findLineNumber(err.Field)
			if lineNum > 0 {
				c.log.Error("第%d行 - %s: %s", lineNum, err.Field, err.Message)
			} else {
				c.log.Error("%s: %s", err.Field, err.Message)
			}
		}
		c.log.DecreaseIndent()
		return fmt.Errorf("任务配置验证失败")
	}

	c.log.Success("任务配置验证通过")
	return nil
}

// CheckProject 检查整个项目
func (c *ConfigChecker) CheckProject(projectPath string) error {
	c.log.Info("检查项目: %s", projectPath)

	// 检查配置文件
	configPath := filepath.Join(projectPath, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		if err := c.CheckProjectConfig(configPath); err != nil {
			return err
		}
	} else {
		c.log.Warning("未找到配置文件: %s", configPath)
	}

	return nil
}

// collectLineInfo 收集YAML节点的行号信息
func (c *ConfigChecker) collectLineInfo(node *yaml.Node, path []string) {
	if node == nil {
		return
	}

	// 记录当前节点的行号
	if len(path) > 0 {
		key := strings.Join(path, ".")
		c.lineMap[key] = node.Line
	}

	// 根据节点类型处理
	switch node.Kind {
	case yaml.DocumentNode:
		// 处理文档节点
		for _, content := range node.Content {
			c.collectLineInfo(content, path)
		}

	case yaml.MappingNode:
		// 处理映射节点
		for i := 0; i < len(node.Content); i += 2 {
			if i+1 < len(node.Content) {
				key := node.Content[i].Value
				newPath := append(append([]string{}, path...), key)
				
				// 记录键的行号
				keyPath := strings.Join(newPath, ".")
				c.lineMap[keyPath] = node.Content[i].Line
				
				// 处理值节点
				c.collectLineInfo(node.Content[i+1], newPath)
			}
		}

	case yaml.SequenceNode:
		// 处理序列节点
		for i, item := range node.Content {
			newPath := append(append([]string{}, path...), fmt.Sprintf("%d", i))
			c.collectLineInfo(item, newPath)
		}

	case yaml.ScalarNode:
		// 标量节点不需要进一步处理
		return
	}
}

// findLineNumber 查找字段对应的行号
func (c *ConfigChecker) findLineNumber(field string) int {
	// 尝试直接匹配
	if line, ok := c.lineMap[field]; ok {
		return line
	}

	// 尝试部分匹配
	parts := strings.Split(field, ".")
	for i := len(parts); i > 0; i-- {
		partialField := strings.Join(parts[:i], ".")
		if line, ok := c.lineMap[partialField]; ok {
			return line
		}
	}

	// 尝试数组索引匹配
	for key, line := range c.lineMap {
		if strings.Contains(field, "[") && strings.Contains(field, "]") {
			// 将 tasks[0].name 转换为 tasks.0.name 格式进行匹配
			modifiedField := strings.ReplaceAll(field, "[", ".")
			modifiedField = strings.ReplaceAll(modifiedField, "]", "")
			if strings.HasPrefix(modifiedField, key) || strings.HasPrefix(key, modifiedField) {
				return line
			}
		}
	}

	return 0
}