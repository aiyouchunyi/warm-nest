// Package action @Author larry
// @Date 2025/4/8 19:13
// @Desc

package actions

import (
	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/machines/code"
	"warm-nest/pkg/tool/machines/model/constant"
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/manual"
)

type RollbackAction struct {
}

func NewRollbackAction() *RollbackAction {
	return &RollbackAction{}
}

func (a *RollbackAction) Action() string {
	return constant.EventActionRollback
}

func (a *RollbackAction) Do(_ manual.ManualConfig, ctx *context.MachineContext, _ manual.ManualEvent) (string, error) {
	preNode := ctx.PreNode()
	if preNode.Status == "" {
		ctx.Log().Errorf("Rollback task forbidden! The task status is empty")
		return "", errors.NewWithArgs(code.TaskRollbackForbidden, ctx.State())
	}
	return preNode.Status, nil
}
