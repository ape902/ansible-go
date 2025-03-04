package models

import (
	"container/heap"
	"fmt"
	"sync"
)

// PriorityTaskQueue 实现基于优先级的任务队列
type PriorityTaskQueue struct {
	mutex      sync.RWMutex
	tasks      taskHeap
	taskMap    map[string]*Task // 用于快速查找任务
	dependents map[string][]string // 记录依赖关系：key是被依赖的任务ID，value是依赖它的任务ID列表
}

// NewPriorityTaskQueue 创建新的优先级任务队列
func NewPriorityTaskQueue() *PriorityTaskQueue {
	pq := &PriorityTaskQueue{
		taskMap:    make(map[string]*Task),
		dependents: make(map[string][]string),
	}
	heap.Init(&pq.tasks)
	return pq
}

// Push 添加任务到队列
func (pq *PriorityTaskQueue) Push(task *Task) error {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	// 检查任务ID是否已存在
	if _, exists := pq.taskMap[task.ID]; exists {
		return fmt.Errorf("任务ID %s 已存在", task.ID)
	}

	// 记录依赖关系
	for _, depID := range task.DependsOn {
		// 检查依赖的任务是否存在
		if _, exists := pq.taskMap[depID]; !exists {
			return fmt.Errorf("依赖的任务ID %s 不存在", depID)
		}
		
		// 记录反向依赖关系
		pq.dependents[depID] = append(pq.dependents[depID], task.ID)
	}

	// 只有没有依赖或依赖已完成的任务才能加入堆
	canQueue := true
	for _, depID := range task.DependsOn {
		depTask, exists := pq.taskMap[depID]
		if !exists || depTask.Status != TaskStatusSuccess {
			canQueue = false
			break
		}
	}

	// 保存任务到映射
	pq.taskMap[task.ID] = task

	// 如果可以加入队列，则加入堆
	if canQueue {
		heap.Push(&pq.tasks, task)
	}

	return nil
}

// Pop 从队列获取下一个任务
func (pq *PriorityTaskQueue) Pop() (*Task, error) {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	if pq.tasks.Len() == 0 {
		return nil, fmt.Errorf("队列为空")
	}

	task := heap.Pop(&pq.tasks).(*Task)
	return task, nil
}

// Remove 从队列移除任务
func (pq *PriorityTaskQueue) Remove(taskID string) error {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	// 检查任务是否存在
	task, exists := pq.taskMap[taskID]
	if !exists {
		return fmt.Errorf("任务ID %s 不存在", taskID)
	}

	// 从堆中移除任务（如果在堆中）
	for i, t := range pq.tasks {
		if t.ID == taskID {
			heap.Remove(&pq.tasks, i)
			break
		}
	}

	// 从映射中移除任务
	delete(pq.taskMap, taskID)

	// 处理依赖关系
	if deps, ok := pq.dependents[taskID]; ok {
		delete(pq.dependents, taskID)
		
		// 如果任务成功完成，检查依赖它的任务是否可以加入队列
		if task.Status == TaskStatusSuccess {
			for _, depID := range deps {
				depTask, exists := pq.taskMap[depID]
				if !exists {
					continue
				}
				
				// 检查该任务的所有依赖是否都已完成
				allDepsCompleted := true
				for _, id := range depTask.DependsOn {
					t, exists := pq.taskMap[id]
					if !exists || t.Status != TaskStatusSuccess {
						allDepsCompleted = false
						break
					}
				}
				
				// 如果所有依赖都已完成，将任务加入队列
				if allDepsCompleted {
					heap.Push(&pq.tasks, depTask)
				}
			}
		}
	}

	return nil
}

// Get 获取指定任务
func (pq *PriorityTaskQueue) Get(taskID string) (*Task, bool) {
	pq.mutex.RLock()
	defer pq.mutex.RUnlock()

	task, exists := pq.taskMap[taskID]
	return task, exists
}

// List 获取所有任务
func (pq *PriorityTaskQueue) List() []*Task {
	pq.mutex.RLock()
	defer pq.mutex.RUnlock()

	tasks := make([]*Task, 0, len(pq.taskMap))
	for _, task := range pq.taskMap {
		tasks = append(tasks, task)
	}

	return tasks
}

// Len 获取任务数量
func (pq *PriorityTaskQueue) Len() int {
	pq.mutex.RLock()
	defer pq.mutex.RUnlock()

	return len(pq.taskMap)
}

// UpdateTaskStatus 更新任务状态并处理依赖关系
func (pq *PriorityTaskQueue) UpdateTaskStatus(taskID string, status TaskStatus) error {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	// 检查任务是否存在
	task, exists := pq.taskMap[taskID]
	if !exists {
		return fmt.Errorf("任务ID %s 不存在", taskID)
	}

	// 更新状态
	task.Status = status

	// 如果任务成功完成，检查依赖它的任务
	if status == TaskStatusSuccess {
		if deps, ok := pq.dependents[taskID]; ok {
			for _, depID := range deps {
				depTask, exists := pq.taskMap[depID]
				if !exists {
					continue
				}
				
				// 检查该任务的所有依赖是否都已完成
				allDepsCompleted := true
				for _, id := range depTask.DependsOn {
					t, exists := pq.taskMap[id]
					if !exists || t.Status != TaskStatusSuccess {
						allDepsCompleted = false
						break
					}
				}
				
				// 如果所有依赖都已完成，将任务加入队列
				if allDepsCompleted {
					heap.Push(&pq.tasks, depTask)
				}
			}
		}
	}

	return nil
}

// 实现container/heap接口所需的方法
type taskHeap []*Task

func (h taskHeap) Len() int { return len(h) }

func (h taskHeap) Less(i, j int) bool {
	// 优先级高的任务先执行
	return h[i].Priority > h[j].Priority
}

func (h taskHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *taskHeap) Push(x interface{}) {
	*h = append(*h, x.(*Task))
}

func (h *taskHeap) Pop() interface{} {
	old := *h
	n := len(old)
	task := old[n-1]
	*h = old[0 : n-1]
	return task
}