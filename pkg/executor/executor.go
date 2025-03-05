package executor

import (
	"fmt"
	"log"
	"sync"

	"github.com/ape902/ansible-go/pkg/config"
	"github.com/ape902/ansible-go/pkg/config/types"
	"github.com/ape902/ansible-go/pkg/executor/connection"
	"github.com/ape902/ansible-go/pkg/executor/engine"
	"github.com/ape902/ansible-go/pkg/executor/executors"
	"github.com/ape902/ansible-go/pkg/executor/models"
	"github.com/ape902/ansible-go/pkg/vars"
)

// Executor 定义执行器
type Executor struct {
	config     *config.Config
	varManager *vars.Manager
	connPool   *connection.Pool
	engine     *engine.ExecutionEngine
}

// NewExecutor 创建新的执行器
func NewExecutor(cfg *config.Config) *Executor {
	connPool := connection.NewPool()
	executorFactory := executors.NewExecutorFactory()
	queue := models.NewPriorityTaskQueue()
	connManager := connection.NewConnectionManager(connPool, &cfg.SSH)
	localVarStore := vars.NewStore()
	return &Executor{
		config:     cfg,
		varManager: vars.NewManager(),
		connPool:   connPool,
		engine: engine.NewExecutionEngine(
			queue,
			connManager,
			localVarStore,
			executorFactory,
			engine.DefaultExecutionOptions,
		),
	}
}

// Execute 执行playbook
func (e *Executor) Execute(playbookPath string) error {
	// 加载playbook
	taskConfig, err := config.LoadPlaybook(playbookPath)
	if err != nil {
		return err
	}

	// 初始化变量
	localVarStore := vars.NewStore()

	// 设置全局变量
	for k, v := range e.config.Vars {
		localVarStore.Set(k, v)
	}

	// 执行任务
	return e.executeTasks(taskConfig, localVarStore)
}

// executeTasks 执行任务列表
func (e *Executor) executeTasks(taskConfig *types.TaskConfig, varStore *vars.Store) error {
	// 检查主机组是否存在
	hosts := make([]string, 0)
	log.Printf("开始解析主机组，共有 %d 个主机组", len(taskConfig.Hosts))
	for _, hostGroup := range taskConfig.Hosts {
		log.Printf("正在处理主机组: %s", hostGroup)
		if hostList, ok := e.config.Inventory[hostGroup]; ok {
			log.Printf("找到主机组 %s，包含 %d 个主机", hostGroup, len(hostList))
			for i, hostInfo := range hostList {
				log.Printf("主机组 %s 中的主机 #%d: %s (端口: %d, 连接类型: %s)",
					hostGroup, i, hostInfo.Host, hostInfo.Port, hostInfo.ConnectionType)
				hosts = append(hosts, hostInfo.Host)
			}
		} else {
			log.Printf("警告: 主机组 %s 不存在", hostGroup)
		}
	}

	log.Printf("主机解析完成，共找到 %d 个可用主机", len(hosts))
	if len(hosts) == 0 {
		return fmt.Errorf("没有找到可用的主机")
	}

	// 创建任务上下文
	ctx := &models.TaskContext{
		Hosts:    hosts,
		VarStore: varStore,
	}

	// 执行任务列表
	var wg sync.WaitGroup
	for _, taskSpec := range taskConfig.Tasks {
		// 处理任务规格映射
		for taskName, spec := range taskSpec {
			wg.Add(1)
			go func(name string, ts types.TaskSpec) {
				defer wg.Done()
				// 创建models.Task类型的任务对象
				modelsTask := &models.Task{
					ID:       name,
					Spec:     &ts,
					Status:   models.TaskStatusPending,
					Priority: models.TaskPriorityNormal,
					Host:     hosts[0], // 使用第一个主机，实际应用中应该循环处理所有主机
					Vars:     make(map[string]interface{}),
				}

				err := e.engine.ExecuteTask(modelsTask, ctx)
				if err != nil {
					log.Printf("执行任务失败: %v", err)
				}
			}(taskName, spec)
		}
	}

	wg.Wait()
	return nil
}
