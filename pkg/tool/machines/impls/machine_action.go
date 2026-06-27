// Package action @Author Larry
// @Date 2024/10/24 09:15
// @Desc

package impls

import (
	"warm-nest/pkg/tool/machines/model/context"
)

type IMachineAction interface {
	BeforeAction(ctx *context.MachineContext, nextStatus string, msg error) error // 前置行为
	TargetAction(ctx *context.MachineContext) error                               // 目标状态行为
	FailAction(ctx *context.MachineContext) error                                 // 失败行为
	SuccAction(ctx *context.MachineContext) error                                 // 成功行为
	FinAction(ctx *context.MachineContext) error                                  // 完成行为
}

type MachineAction struct{}

func NewAction() IMachineAction {
	return &MachineAction{}
}

func (m *MachineAction) BeforeAction(_ *context.MachineContext, _ string, _ error) error {
	return nil
}

// TargetAction 目标状态行为
func (m *MachineAction) TargetAction(_ *context.MachineContext) error {
	return nil
}

// FailAction 失败行为
func (m *MachineAction) FailAction(_ *context.MachineContext) error {
	return nil
}

// SuccAction 成功行为
func (m *MachineAction) SuccAction(_ *context.MachineContext) error {
	return nil
}

// FinAction 完成行为
func (m *MachineAction) FinAction(_ *context.MachineContext) error {
	return nil
}
