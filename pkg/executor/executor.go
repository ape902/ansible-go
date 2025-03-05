package executor

import (
	"fmt"
	"sync"

	"github.com/ape902/ansible-go/pkg/config"
	"github.com/ape902/ansible-go/pkg/config/types"
	"github.com/ape902/ansible-go/pkg/executor/connection"
	"github.com/ape902/ansible-go/pkg/executor/engine"
	"github.com/ape902/ansible-go/pkg/executor/executors"
	"github.com/ape902/ansible-go/pkg/executor/models"
	"github.com/ape902/ansible-go/pkg/logger"
	"github.com/ape902/ansible-go/pkg/vars"
)

// Executor 定义执行器
type Executor struct {
	config     *config.Config
	varManager *vars.Manager
	connPool   *connection.Pool
	engine     *engine.ExecutionEngine
	logger     *logger.Logger
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
		logger:     logger.New(),
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
	e.logger.Info("开始解析主机组，共有 %d 个主机组", len(taskConfig.Hosts))
	e.logger.IncreaseIndent()

	for _, hostGroup := range taskConfig.Hosts {
		e.logger.Info("正在处理主机组: %s", hostGroup)
		e.logger.IncreaseIndent()

		if hostList, ok := e.config.Inventory[hostGroup]; ok {
			e.logger.Success("找到主机组 %s，包含 %d 个主机", hostGroup, len(hostList))
			e.logger.IncreaseIndent()

			for i, hostInfo := range hostList {
				e.logger.Info("主机 #%d: %s (端口: %d, 连接类型: %s)",
					i, hostInfo.Host, hostInfo.Port, hostInfo.ConnectionType)
				hosts = append(hosts, hostInfo.Host)
			}

			e.logger.DecreaseIndent()
		} else {
			e.logger.Warning("警告: 主机组 %s 不存在", hostGroup)
		}

		e.logger.DecreaseIndent()
	}

	e.logger.DecreaseIndent()
	e.logger.Success("主机解析完成，共找到 %d 个可用主机", len(hosts))
	if len(hosts) == 0 {
		return fmt.Errorf("没有找到可用的主机")
	}

	// 创建任务上下文
	ctx := &models.TaskContext{
		Hosts:    hosts,
		VarStore: varStore,
	}

	// 创建工作池
	workerCount := 10 // 默认并发数
	if e.config.SSH.MaxParallel > 0 {
		workerCount = e.config.SSH.MaxParallel
	}

	// 创建任务通道
	taskChan := make(chan *models.Task, len(hosts)*len(taskConfig.Tasks))
	errChan := make(chan error, len(hosts)*len(taskConfig.Tasks))

	// 启动工作池
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				err := e.engine.ExecuteTask(task, ctx)
				if err != nil {
					e.logger.Error("在主机 %s 上执行任务 %s 失败: %v", task.Host, task.ID, err)
					errChan <- err
				} else if task.Result != nil {
					// 输出任务执行结果
					if task.Result.Stdout != "" {
						e.logger.Output(task.Host, task.ID, task.Result.Stdout)
					}
					if task.Result.Stderr != "" {
						e.logger.Error("主机 %s 上的任务 %s 的错误输出:", task.Host, task.ID)
						e.logger.Output(task.Host, task.ID, task.Result.Stderr)
					}
				}
			}
		}()
	}

	// 分发任务
	for _, taskSpec := range taskConfig.Tasks {
		for taskName, spec := range taskSpec {
			for _, host := range hosts {
				modelsTask := &models.Task{
					ID:       taskName,
					Spec:     &spec,
					Status:   models.TaskStatusPending,
					Priority: models.TaskPriorityNormal,
					Host:     host,
					Vars:     make(map[string]interface{}),
				}
				taskChan <- modelsTask
			}
		}
	}

	// 关闭任务通道
	close(taskChan)

	// 等待所有工作协程完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误发生
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("执行任务时发生错误: %v", errs)
	}

	return nil
}

// SetVerboseMode 设置详细输出模式
func (e *Executor) SetVerboseMode(verbose bool) {
	e.engine.SetVerbose(verbose)
}
