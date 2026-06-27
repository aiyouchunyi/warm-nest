// Package machines @Author larry
// @Date 2025/3/7 16:51
// @Desc

package machines

import (
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/machines/code"
	"warm-nest/pkg/tool/machines/model/context"
)

// Sync 任务状态同步
func (m *Machine[T]) Sync(ctx *context.MachineContext) error {
	err := m.MachineRepository.Sync(ctx)
	if err != nil {
		ctx.Log().WithError(err).Errorf("Task sync status failed")
		return errors.NewWithArgs(code.TaskSyncException, ctx.State(), err)
	}
	return nil
}

// FindTask 获取任务
func (m *Machine[T]) FindTask(taskId string) (T, error) {
	t, err := m.MachineRepository.FindTask(taskId)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"taskId": taskId,
		}).Errorf("Machine find task failed!")
		return t, errors.NewWithArgs(code.TaskNotFound, taskId, err)
	}
	return t, nil
}

// FindTasks 获取任务列表
func (m *Machine[T]) FindTasks(taskIds []string) ([]T, error) {
	tasks, err := m.MachineRepository.FindTasks(taskIds)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"taskIds": taskIds,
		}).Errorf("Machine find tasks failed!")
		return nil, errors.NewWithArgs(code.TaskNotFound, taskIds, err)
	}
	return tasks, nil
}

// MonitorTasks 查询监控任务列表
func (m *Machine[T]) MonitorTasks() ([]T, error) {
	tasks, err := m.MachineRepository.MonitorTasks(m.MachineStatus.AutoStatues, m.MachineConfig)
	if err != nil {
		logrus.WithError(err).Errorf("Machine query tasks failed!")
		return nil, err
	}
	return tasks, nil
}

// Lock 锁定任务
func (m *Machine[T]) Lock(ctx *context.MachineContext) error {
	if ctx.Lock {
		ctx.Log().Warnf("Task already locked!")
		return errors.NewWithArgs(code.TaskAlreadyLocked, ctx.State())
	}
	succ, err := m.MachineRepository.Lock(ctx)
	if err != nil {
		ctx.Log().WithError(err).Errorf("Task lock failed")
		return errors.NewWithArgs(code.TaskLockException, ctx.State(), err)
	}
	if !succ {
		ctx.Log().Warnf("Task already locked!")
		return errors.NewWithArgs(code.TaskAlreadyLocked, ctx.State())
	}
	return nil
}

// Unlock 解锁任务
func (m *Machine[T]) Unlock(ctx *context.MachineContext) {
	succ, err := m.MachineRepository.Unlock(ctx)
	if err != nil {
		ctx.Log().WithError(err).Errorf("Task unlock failed")
		return
	}
	if !succ {
		ctx.Log().Warnf("Task already unlocked!")
	}
}

// ReLock 释放锁
func (m *Machine[T]) ReLock() error {
	log := logrus.WithFields(logrus.Fields{
		"name": m.Name,
	})
	rows, err := m.MachineRepository.ReLock(m.LockTime)
	if err != nil {
		log.WithError(err).Errorf("Machine release lock failed!")
		return nil
	}
	log.WithFields(logrus.Fields{
		"rows": rows,
	}).Infof("Machine release lock heart ...")
	return nil
}
