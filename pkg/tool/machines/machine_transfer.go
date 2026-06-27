// Package machines @Author larry
// @Date 2025/5/21 17:13
// @Desc

package machines

import (
	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/machines/code"
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/task/info"
)

// Transfer 执行状态流转，包含完整生命周期钩子
// 执行顺序：Before → DoNext → Sync → Target → Fail → Succ → Fin
func (m *Machine[T]) Transfer(ctx *context.MachineContext, nextStatus string, msg error, manuals ...info.ManualInfo) error {
	if err := m.BeforeAction(ctx, nextStatus, msg); err != nil {
		return err
	}

	change := ctx.DoNext(nextStatus, errors.ErrMsg(msg), manuals...)
	if err := m.Sync(ctx); err != nil {
		return err
	}

	if !change {
		ctx.Log().Info("Task status unchanged, operation complete")
		return nil
	}

	if err := m.TargetAction(ctx); err != nil {
		return err
	}

	if err := m.FailAction(ctx); err != nil {
		return err
	}

	if err := m.SuccAction(ctx); err != nil {
		return err
	}

	if err := m.FinAction(ctx); err != nil {
		return err
	}
	return nil
}

// BeforeAction 前置行为
func (m *Machine[T]) BeforeAction(ctx *context.MachineContext, nextStatus string, msg error) error {
	err := m.MachineAction.BeforeAction(ctx, nextStatus, msg)
	if err != nil {
		ctx.Log().WithError(err).Errorf("Action before failed")
		return errors.NewWithArgs(code.MachineBeforeActionException, ctx.State(), err)
	}
	return nil
}

// TargetAction 目标行为
func (m *Machine[T]) TargetAction(ctx *context.MachineContext) error {
	err := m.MachineAction.TargetAction(ctx)
	if err != nil {
		ctx.Log().WithError(err).Errorf("Action target failed")
		return errors.NewWithArgs(code.MachineTargetActionException, ctx.State(), err)
	}
	return nil
}

// FailAction 失败行为
func (m *Machine[T]) FailAction(ctx *context.MachineContext) error {
	if !ctx.IsFailStatus() {
		return nil
	}
	err := m.MachineAction.FailAction(ctx)
	if err != nil {
		ctx.Log().WithError(err).Errorf("Action fail failed")
		return errors.NewWithArgs(code.MachineFailActionException, ctx.State(), err)
	}
	return nil
}

// SuccAction 成功行为
func (m *Machine[T]) SuccAction(ctx *context.MachineContext) error {
	if !ctx.IsSuccessStatus() {
		return nil
	}
	err := m.MachineAction.SuccAction(ctx)
	if err != nil {
		ctx.Log().WithError(err).Errorf("Action succ failed")
		return errors.NewWithArgs(code.MachineSuccActionException, ctx.State(), err)
	}
	return nil
}

// FinAction 完成行为
func (m *Machine[T]) FinAction(ctx *context.MachineContext) error {
	if !ctx.IsFinalStatus() {
		return nil
	}
	err := m.MachineAction.FinAction(ctx)
	if err != nil {
		ctx.Log().WithError(err).Errorf("Action fin failed")
		return errors.NewWithArgs(code.MachineFinActionException, ctx.State(), err)
	}
	return nil
}
