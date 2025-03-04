package executor

import (
	"fmt"
	"log"

	"github.com/ape902/ansible-go/pkg/config"
	"github.com/ape902/ansible-go/pkg/config/types"
	"github.com/ape902/ansible-go/pkg/vars"
)

// Executor 定义执行器结构
type Executor struct {
	config   *config.Config
	varStore *vars.Store
	handlers map[string]HandlerFunc
}

// HandlerFunc 定义处理器函数类型
type HandlerFunc func(args map[string]interface{}, varStore *vars.Store) error

// NewExecutor 创建新的执行器实例
func NewExecutor(cfg *config.Config, varStore *vars.Store) (*Executor, error) {
	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}
	if varStore == nil {
		return nil, fmt.Errorf("变量存储不能为空")
	}

	// 初始化执行器
	exec := &Executor{
		config:   cfg,
		varStore: varStore,
		handlers: make(map[string]HandlerFunc),
	}

	// 注册内置处理器
	exec.registerBuiltinHandlers()

	// 合并全局变量
	if err := varStore.Merge(cfg.Vars); err != nil {
		log.Printf("警告: 合并全局变量时出现冲突: %v", err)
	}

	return exec, nil
}

// registerBuiltinHandlers 注册内置处理器
func (e *Executor) registerBuiltinHandlers() {
	// 注册命令处理器
	e.handlers["command"] = func(args map[string]interface{}, varStore *vars.Store) error {
		cmd, ok := args["cmd"].(string)
		if !ok {
			return fmt.Errorf("command模块必须提供cmd参数")
		}
		log.Printf("执行命令: %s", cmd)
		// TODO: 实现命令执行逻辑
		return nil
	}

	// 注册文件处理器
	e.handlers["file"] = func(args map[string]interface{}, varStore *vars.Store) error {
		path, ok := args["path"].(string)
		if !ok {
			return fmt.Errorf("file模块必须提供path参数")
		}
		log.Printf("处理文件: %s", path)
		// TODO: 实现文件操作逻辑
		return nil
	}

	// 注册模板处理器
	e.handlers["template"] = func(args map[string]interface{}, varStore *vars.Store) error {
		src, ok := args["src"].(string)
		if !ok {
			return fmt.Errorf("template模块必须提供src参数")
		}
		dest, ok := args["dest"].(string)
		if !ok {
			return fmt.Errorf("template模块必须提供dest参数")
		}
		log.Printf("处理模板: %s -> %s", src, dest)
		// TODO: 实现模板渲染逻辑
		return nil
	}
}

// RegisterHandler 注册自定义处理器
func (e *Executor) RegisterHandler(name string, handler HandlerFunc) {
	e.handlers[name] = handler
}

// ExecutePlaybook 执行playbook
func (e *Executor) ExecutePlaybook(playbookPath string) error {
	// 加载playbook
	taskConfig, err := config.LoadPlaybook(playbookPath)
	if err != nil {
		return fmt.Errorf("加载playbook失败: %w", err)
	}

	// 创建本地变量存储
	localVarStore := vars.NewStore()

	// 合并playbook变量
	if err := localVarStore.Merge(e.varStore.GetAll()); err != nil {
		return fmt.Errorf("合并全局变量失败: %w", err)
	}
	if err := localVarStore.Merge(taskConfig.Vars); err != nil {
		return fmt.Errorf("合并playbook变量失败: %w", err)
	}

	// 执行任务
	return e.executeTasks(taskConfig, localVarStore)
}

// executeTasks 执行任务列表
func (e *Executor) executeTasks(taskConfig *types.TaskConfig, varStore *vars.Store) error {
	// 检查主机组是否存在
	hosts := make([]string, 0)
	for _, hostGroup := range taskConfig.Hosts {
		if hostList, ok := e.config.Inventory[hostGroup]; ok {
			hosts = append(hosts, hostList...)
		} else {
			log.Printf("警告: 主机组 %s 不存在", hostGroup)
		}
	}

	if len(hosts) == 0 {
		return fmt.Errorf("没有找到可用的主机")
	}

	// 执行每个任务
	for _, taskMap := range taskConfig.Tasks {
		for taskName, taskSpec := range taskMap {
			log.Printf("执行任务: %s", taskName)

			// 合并任务变量
			taskVarStore := vars.NewStore()
			if err := taskVarStore.Merge(varStore.GetAll()); err != nil {
				return fmt.Errorf("合并任务变量失败: %w", err)
			}
			if err := taskVarStore.Merge(taskSpec.Vars); err != nil {
				return fmt.Errorf("合并任务变量失败: %w", err)
			}

			// 检查条件
			if taskSpec.When != "" {
				// TODO: 实现条件判断逻辑
				log.Printf("条件判断: %s", taskSpec.When)
			}

			// 获取处理器
			handler, ok := e.handlers[taskSpec.Module]
			if !ok {
				return fmt.Errorf("未知的模块: %s", taskSpec.Module)
			}

			// 执行任务
			err := handler(taskSpec.Args, taskVarStore)
			if err != nil {
				if taskSpec.IgnoreError {
					log.Printf("任务执行失败，但已忽略: %v", err)
				} else {
					return fmt.Errorf("任务执行失败: %w", err)
				}
			}

			// 处理通知
			if len(taskSpec.Notify) > 0 {
				// TODO: 实现处理器通知逻辑
				log.Printf("触发通知: %v", taskSpec.Notify)
			}
		}
	}

	return nil
}
