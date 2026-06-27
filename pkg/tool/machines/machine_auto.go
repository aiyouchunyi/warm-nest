// Package machines @Author larry
// @Date 2025/2/8 13:56
// @Desc

package machines

import (
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/utils/strings"
)

// Run 执行状态机，从当前状态开始自动流转
// 自动流转至停止状态
func (m *Machine[T]) Run(ctx *context.MachineContext) error {
	if ctx.IsHaltStatus() {
		ctx.Log().Warnf("Task status is halt, no need to run machine")
		return nil
	}

	ctx.Log().Infof("==> Machine begin")
	if err := m.Lock(ctx); err != nil {
		return err
	}

	defer func() {
		m.Unlock(ctx)
		ctx.Log().Infof("<== Machine complete")
	}()

	for !ctx.IsHaltStatus() {
		// 1.执行当前状态
		ctx.Log().Infof("--> Transition begin")

		nextStatus, msg := m.execute(ctx)

		ctx.Log().WithFields(logrus.Fields{
			"nextStatus": nextStatus,
			"msg":        msg,
		}).Infof("<-- Transition complete")

		// 2.状态转换后行为
		if err := m.Transfer(ctx, nextStatus, msg); err != nil {
			ctx.Log().WithError(err).Errorf("Task transfer failed, machine stop")
			return err
		}

		// 3.状态未变化，退出循环
		if strings.IsBlank(nextStatus) || !ctx.Change {
			ctx.Log().Info("Task status unchanged, machine stop")
			break
		}
	}
	return nil
}

// execute 执行
func (m *Machine[T]) execute(ctx *context.MachineContext) (string, error) {
	transition, err := m.machineTransition.GetAutoTransition(ctx)
	if err != nil {
		return "", err
	}
	if err = transition.Before(ctx); err != nil {
		return "", err
	}
	nextStatus, msg := transition.Execute(ctx)

	// 存在错误信息，即状态转换失败，执行失败行为
	if msg != nil {
		if err = transition.Fail(ctx, nextStatus, msg); err != nil {
			return "", err
		}
		return nextStatus, msg
	}

	// 不存在错误信息且状态发生变化，执行成功行为
	if strings.NotBlank(nextStatus) && ctx.Status != nextStatus {
		if err = transition.Succ(ctx, nextStatus); err != nil {
			return "", err
		}
		return nextStatus, nil
	}
	return nextStatus, nil
}
