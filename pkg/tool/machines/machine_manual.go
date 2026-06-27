// Package machines @Author larry
// @Date 2025/2/12 10:34
// @Desc

package machines

import (
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/machines/code"
	"warm-nest/pkg/tool/machines/impls"
	"warm-nest/pkg/tool/machines/model/constant"
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/manual"
	"warm-nest/pkg/tool/machines/model/task/info"
	"warm-nest/pkg/tool/machines/model/variable"
	"warm-nest/pkg/utils/slices"
	"warm-nest/pkg/utils/strings"
	"warm-nest/pkg/utils/times"
)

// Manual 执行手动状态流转，带权限校验
func (m *Machine[T]) Manual(task T, event manual.ManualEvent, options ...variable.Option) error {
	ctx := context.NewContext(task, m.MachineStatus, m.MachineConfig, options...)
	ctx.Log().WithFields(logrus.Fields{
		"event": event,
	}).Info("Manual operation received")

	ctx.Log().Infof("==> Machine manual begin")

	// 1.锁定任务
	if err := m.Lock(ctx); err != nil {
		return err
	}

	var err error
	defer func() {
		m.Unlock(ctx)
		if err == nil {
			m.Submit(task, options...)
		}
		ctx.Log().Infof("<== Machine manual complete")
	}()

	// 2.获取下个状态
	nextStatus, err := m.manual(ctx, event)
	if strings.IsBlank(nextStatus) {
		return errors.Union(err, errors.NewWithArgs(code.TaskManualException, ctx.State()))
	}

	// 3.执行状态转移
	err = m.Transfer(ctx, nextStatus, err, info.ManualInfo{
		Operator: event.Operator.UserId,
		Action:   event.Action,
		Opinion:  event.Opinion,
		Params:   event.Params,
		Time:     times.UnixMilli(),
	})
	if err != nil {
		return err
	}
	// event 事件监听器，监听到事件后执行相应操作，如发送通知、记录日志等
	if ctx.Change {
		slices.ForEach(m.EventListeners, func(listener impls.IMachineEventListener) {
			if !slices.Contains(listener.Match(), event.Action) {
				return
			}
			err = listener.Listen(ctx, event)
			if err != nil {
				ctx.Log().WithError(err).Errorf("Machine event listener %s failed", listener.Name())
			}
		})
	}
	return nil
}

// manual 手动流转任务
func (m *Machine[T]) manual(ctx *context.MachineContext, event manual.ManualEvent) (string, error) {
	ctx.Log().Infof("--> Transition manual begin")
	nextStatus := ""
	var msg error
	defer func() {
		ctx.Log().WithFields(logrus.Fields{
			"nextStatus": nextStatus,
			"error":      msg,
		}).Infof("<-- Transition manual complete")
	}()
	if event.Action == constant.EventActionForce {
		return m.doForce(ctx, event)
	}
	return m.doManual(ctx, event)
}

// doManual 手动流转任务
func (m *Machine[T]) doManual(ctx *context.MachineContext, event manual.ManualEvent) (string, error) {
	// 1.检查状态是否可操作
	if !ctx.IsManualStatus() {
		ctx.Log().Error("Manual operation forbidden: not a manual status")
		return "", errors.NewWithArgs(code.TaskManualForbidden, ctx.State())
	}

	// 2.获的执行器
	transition, err := m.machineTransition.GetManualTransition(ctx)
	if err != nil {
		return "", err
	}

	// 3.操作检查
	if err = m.MachineAudit.Audit(ctx, transition.Config(), event); err != nil {
		return "", err
	}

	// 4.状态转换前行为
	if err = transition.Before(ctx, event); err != nil {
		return "", err
	}

	// 5.状态转换
	nextStatus, msg := transition.Execute(ctx, event)
	if msg != nil {
		if err = transition.Fail(ctx, event, nextStatus, msg); err != nil {
			return "", err
		}
		return nextStatus, msg
	}
	if strings.NotBlank(nextStatus) && ctx.Status != nextStatus {
		if err = transition.Succ(ctx, event, nextStatus); err != nil {
			return "", err
		}
		return nextStatus, nil
	}
	return nextStatus, nil
}

// doForce 手动停止任务
func (m *Machine[T]) doForce(ctx *context.MachineContext, event manual.ManualEvent) (string, error) {
	if strings.IsBlank(event.Operator.UserId) {
		return "", errors.NewWithArgs(code.UnauthorizedOperatorLoss, ctx.State())
	}
	if !event.Operator.Supper {
		ctx.Log().WithFields(logrus.Fields{
			"operator": event.Operator,
		}).Error("Task force forbidden! the operator is not supper")
		return "", errors.NewWithArgs(code.UnauthorizedDisAdmin, ctx.State(), event.Operator.UserId)
	}
	if ctx.Status == event.NextStatus {
		ctx.Log().WithFields(logrus.Fields{
			"nextStatus": event.NextStatus,
		}).Error("Task force forbidden! the nextStatus is same as current status")
		return "", errors.NewWithArgs(code.TaskForceStatusUnchanged, ctx.State())
	}
	if !ctx.MachineStatus.IsValidStatus(event.NextStatus) {
		ctx.Log().WithFields(logrus.Fields{
			"nextStatus": event.NextStatus,
		}).Error("Task force forbidden! the nextStatus is invalid")
		return "", errors.NewWithArgs(code.TaskManualStatusInvalid, ctx.State(), constant.EventActionMap[constant.EventActionForce], event.NextStatus)
	}
	return event.NextStatus, nil
}
