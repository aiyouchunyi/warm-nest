// Package transitionMap @Author larry
// @Date 2025/2/8 15:56
// @Desc

package impls

import (
	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/machines/code"
	"warm-nest/pkg/tool/machines/impls/transitions"
	"warm-nest/pkg/tool/machines/impls/transitions/actions"
	"warm-nest/pkg/tool/machines/model/constant"
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/manual"
	"warm-nest/pkg/utils/slices"
)

type IMachineTransition interface {
	GetMachineStatus() context.MachineStatus
	GetAutoTransition(ctx *context.MachineContext) (transitions.IAutoTransition, error)
	GetManualTransition(ctx *context.MachineContext) (transitions.IManualTransition, error)
}

// MachineTransition 状态转换器状态工厂
type MachineTransition struct {
	context.MachineStatus
	transitionMap map[string]transitions.ITransition
}

func NewDefaultTransition(transitions ...transitions.ITransition) *MachineTransition {
	return NewTransition([]string{constant.MachineSuccess}, []string{constant.MachineFailed}, transitions...)
}

func NewTransition(succStatuses []string, failStatuses []string, trans ...transitions.ITransition) *MachineTransition {
	trans = transitions.AddIfAbsent(trans, transitions.NewTransitionAction(constant.MachineWait, manual.NewManualConfig(manual.AuthRole("super_admin")), actions.NewRollbackAction()))
	trans = transitions.AddIfAbsent(trans, transitions.NewTransitionAction(constant.MachineCancel, manual.NewManualConfig(manual.AuthRole("super_admin")), actions.NewRollbackAction()))
	succStatuses = slices.AddIfAbsent(succStatuses, constant.MachineSuccess)
	failStatuses = slices.AddIfAbsent(failStatuses, constant.MachineFailed)
	autoStatues := make([]string, 0)
	manualStatuses := make([]string, 0)
	for _, transition := range trans {
		_, ok := transition.(transitions.IAutoTransition)
		if ok {
			autoStatues = append(autoStatues, transition.State())
		} else {
			manualStatuses = append(manualStatuses, transition.State())
		}
	}
	return &MachineTransition{
		transitionMap: slices.ToMap(trans, func(transition transitions.ITransition) string {
			return transition.State()
		}),
		MachineStatus: context.NewMachineStatus(succStatuses, failStatuses, autoStatues, manualStatuses),
	}
}

// GetMachineStatus 获取状态机状态配置
func (t *MachineTransition) GetMachineStatus() context.MachineStatus {
	return t.MachineStatus
}

// GetAutoTransition 获取自动状态转换器
func (t *MachineTransition) GetAutoTransition(ctx *context.MachineContext) (transitions.IAutoTransition, error) {
	transition, ok := t.transitionMap[ctx.Status]
	if !ok {
		ctx.Log().Error("Machine auto transition not found!")
		return nil, errors.NewWithArgs(code.MachineAutoTransitionNotFound, ctx.State())
	}
	autoTransition, ok := transition.(transitions.IAutoTransition)
	if !ok {
		ctx.Log().Error("Machine auto transition not found!")
		return nil, errors.NewWithArgs(code.MachineAutoTransitionNotFound, ctx.State())
	}
	return autoTransition, nil
}

// GetManualTransition 获取手动状态转换器
func (t *MachineTransition) GetManualTransition(ctx *context.MachineContext) (transitions.IManualTransition, error) {
	transition, ok := t.transitionMap[ctx.Status]
	if !ok {
		ctx.Log().Error("Machine manual transition not found!")
		return nil, errors.NewWithArgs(code.MachineManualTransitionNotFound, ctx.State())
	}
	manualTransition, ok := transition.(transitions.IManualTransition)
	if !ok {
		ctx.Log().Error("Machine manual transition not found!")
		return nil, errors.NewWithArgs(code.MachineManualTransitionNotFound, ctx.State())
	}
	return manualTransition, nil
}
