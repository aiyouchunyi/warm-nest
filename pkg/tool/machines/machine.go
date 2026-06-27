// Package machines @Author Larry
// @Date 2024/10/22 19:09
// @Desc

package machines

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/tool/machines/impls"
	"warm-nest/pkg/tool/machines/model/config"
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/dto"
	"warm-nest/pkg/tool/machines/model/task"
	"warm-nest/pkg/tool/machines/model/variable"
	"warm-nest/pkg/tool/pools"
	"warm-nest/pkg/tool/timers"
)

type Machine[T task.IMachineTask] struct {
	config.MachineConfig //状态机配置
	context.MachineStatus
	machinePool       *pools.Pool
	machineTransition impls.IMachineTransition
	MachineAction     impls.IMachineAction //状态机行为
	MachineAudit      impls.MachineAudit
	MachineRepository impls.IMachineRepository[T]
	EventListeners    []impls.IMachineEventListener
}

// NewMachine 创建状态机
func NewMachine[T task.IMachineTask](name string, machineTransition impls.IMachineTransition, options ...config.Option) *Machine[T] {
	machineConfig := config.GetMachineConfig(name)
	for _, option := range options {
		option(&machineConfig)
	}
	m := &Machine[T]{
		MachineConfig:     machineConfig,
		MachineStatus:     machineTransition.GetMachineStatus(),
		MachineRepository: impls.NewRepository[T](),
		MachineAction:     impls.NewAction(),
		MachineAudit:      impls.NewAudit(),
		machineTransition: machineTransition,
		machinePool:       pools.NewPool(name + "Pool"),
	}
	if m.ScanLock > 0 {
		timers.Timer(fmt.Sprintf("%s-ReLock", m.Name), m.ScanLock, m.ReLock)
	}
	logrus.WithFields(logrus.Fields{
		"machine": m.MachineStatus,
	}).Infof("%s machine started...", m.Name)
	return m
}

// Execute 执行任务节点，适用于单个任务的场景，返回执行结果和错误信息
func (m *Machine[T]) Execute(task T, options ...variable.Option) (dto.TaskExecuteInfo, error) {
	ctx := context.NewContext(task, m.MachineStatus, m.MachineConfig, options...)
	err := m.Run(ctx)
	if err != nil {
		ctx.Log().WithError(err).Error("Machine do tasks failed!")
		return dto.NewExecuteInfo(ctx.ToExecuteStatus(), ctx), err
	}
	return dto.NewExecuteInfo(ctx.ToExecuteStatus(), ctx), nil
}

// Executes 批量执行任务节点
func (m *Machine[T]) Executes(tasks []T, options ...variable.Option) ([]dto.TaskExecuteInfo, error) {
	var results []dto.TaskExecuteInfo
	for index := range tasks {
		result, err := m.Execute(tasks[index], options...)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}

// Submit 提交任务节点，适用于异步执行的场景，无返回值
func (m *Machine[T]) Submit(task T, options ...variable.Option) {
	ctx := context.NewContext(task, m.MachineStatus, m.MachineConfig, options...)
	err := m.machinePool.Submit(func() {
		err2 := m.Run(ctx)
		if err2 != nil {
			ctx.Log().WithError(err2).Error("Machine async do tasks failed!")
		}
	})
	if err != nil {
		ctx.Log().WithError(err).Error("Machine submit tasks failed!")
	}
}

// Submits 批量提交任务节点
func (m *Machine[T]) Submits(tasks []T, options ...variable.Option) {
	for index := range tasks {
		m.Submit(tasks[index], options...)
	}
}

// AddEventListeners 设置状态机事件监听器
func (m *Machine[T]) AddEventListeners(eventListeners ...impls.IMachineEventListener) *Machine[T] {
	m.EventListeners = append(m.EventListeners, eventListeners...)
	return m
}
