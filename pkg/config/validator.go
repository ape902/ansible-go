package config

import (
	"fmt"
	"strings"

	"github.com/ape902/ansible-go/pkg/config/types"
)

// ConfigValidationError 定义配置验证错误
type ConfigValidationError struct {
	Field   string
	Message string
	Line    int
}

// ValidateConfig 验证配置文件
func ValidateConfig(cfg *Config) []ConfigValidationError {
	var errors []ConfigValidationError

	// 验证主机清单
	errors = append(errors, validateInventory(cfg.Inventory)...)

	// 验证SSH配置
	errors = append(errors, validateSSHConfig(cfg.SSH)...)

	// 验证变量
	errors = append(errors, validateVars(cfg.Vars)...)

	return errors
}

// ValidateTaskConfig 验证任务配置
func ValidateTaskConfig(taskCfg *types.TaskConfig) []ConfigValidationError {
	var errors []ConfigValidationError

	// 验证基本字段
	if taskCfg.Name == "" {
		errors = append(errors, ConfigValidationError{
			Field:   "name",
			Message: "任务名称不能为空",
		})
	}

	if len(taskCfg.Hosts) == 0 {
		errors = append(errors, ConfigValidationError{
			Field:   "hosts",
			Message: "目标主机列表不能为空",
		})
	}

	// 验证任务列表
	for i, task := range taskCfg.Tasks {
		for name, spec := range task {
			errs := validateTaskSpec(name, spec)
			for _, err := range errs {
				err.Field = fmt.Sprintf("tasks[%d].%s.%s", i, name, err.Field)
				errors = append(errors, err)
			}
		}
	}

	// 验证处理器
	for i, handler := range taskCfg.Handlers {
		if handler.Name == "" {
			errors = append(errors, ConfigValidationError{
				Field:   fmt.Sprintf("handlers[%d].name", i),
				Message: "处理器名称不能为空",
			})
		}

		if handler.Module == "" {
			errors = append(errors, ConfigValidationError{
				Field:   fmt.Sprintf("handlers[%d].module", i),
				Message: "处理器模块不能为空",
			})
		}
	}

	return errors
}

// validateTaskSpec 验证任务规格
func validateTaskSpec(name string, spec types.TaskSpec) []ConfigValidationError {
	var errors []ConfigValidationError

	if spec.Module == "" {
		errors = append(errors, ConfigValidationError{
			Field:   "module",
			Message: "模块名称不能为空",
		})
	}

	if spec.Retries < 0 {
		errors = append(errors, ConfigValidationError{
			Field:   "retries",
			Message: "重试次数不能为负数",
		})
	}

	// 检查notify列表中是否有重复项
	notifyMap := make(map[string]bool)
	for i, handler := range spec.Notify {
		if handler == "" {
			errors = append(errors, ConfigValidationError{
				Field:   fmt.Sprintf("notify[%d]", i),
				Message: "处理器名称不能为空",
			})
			continue
		}
		
		if notifyMap[handler] {
			errors = append(errors, ConfigValidationError{
				Field:   fmt.Sprintf("notify[%d]", i),
				Message: fmt.Sprintf("重复的处理器名称: %s", handler),
			})
		} else {
			notifyMap[handler] = true
		}
	}

	return errors
}

// validateInventory 验证主机清单
func validateInventory(inventory map[string][]types.HostInfo) []ConfigValidationError {
	var errors []ConfigValidationError

	for groupName, hosts := range inventory {
		if groupName == "" {
			errors = append(errors, ConfigValidationError{
				Field:   "inventory",
				Message: "主机组名称不能为空",
			})
		}

		for i, host := range hosts {
			if host.Host == "" {
				errors = append(errors, ConfigValidationError{
					Field:   fmt.Sprintf("inventory.%s[%d].host", groupName, i),
					Message: "主机地址不能为空",
				})
			}

			if host.Port <= 0 || host.Port > 65535 {
				errors = append(errors, ConfigValidationError{
					Field:   fmt.Sprintf("inventory.%s[%d].port", groupName, i),
					Message: "端口号必须在1-65535之间",
				})
			}
		}
	}

	return errors
}

// validateSSHConfig 验证SSH配置
func validateSSHConfig(cfg types.SSHConfig) []ConfigValidationError {
	var errors []ConfigValidationError

	if cfg.User == "" {
		errors = append(errors, ConfigValidationError{
			Field:   "ssh.user",
			Message: "SSH用户名不能为空",
		})
	}

	if cfg.UseKeyAuth && cfg.KeyFile == "" {
		errors = append(errors, ConfigValidationError{
			Field:   "ssh.key_file",
			Message: "使用密钥认证时，私钥路径不能为空",
		})
	}

	return errors
}

// validateVars 验证变量
func validateVars(vars map[string]interface{}) []ConfigValidationError {
	var errors []ConfigValidationError

	for key := range vars {
		if strings.TrimSpace(key) == "" {
			errors = append(errors, ConfigValidationError{
				Field:   "vars",
				Message: "变量名不能为空",
			})
		}
	}

	return errors
}