// Package action @Author larry
// @Date 2025/4/9 16:28
// @Desc

package actions

import (
	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/machines/code"
	"warm-nest/pkg/tool/machines/model/constant"
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/manual"
	"warm-nest/pkg/utils/slices"
)

type DismissAction struct {
	*NextAction
}

func NewDismissAction(preState string) *DismissAction {
	return &DismissAction{
		NewNextAction(constant.EventActionDismiss, preState),
	}
}

func (a *DismissAction) Do(cfg manual.ManualConfig, ctx *context.MachineContext, event manual.ManualEvent) (string, error) {
	if !slices.Contain(ctx.HistoryStatuses(), a.nextState) {
		ctx.Log().Errorf("Dismiss task forbidden! The task status is invalid")
		return "", errors.NewWithArgs(code.TaskManualStatusInvalid, "驳回", ctx.State(), a.nextState)
	}
	return a.NextAction.Do(cfg, ctx, event)
}
