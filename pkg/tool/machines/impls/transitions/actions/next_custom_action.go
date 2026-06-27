// Package action @Author larry
// @Date 2025/4/9 11:09
// @Desc

package actions

import (
	"github.com/sirupsen/logrus"

	errors2 "warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/machines/code"
	"warm-nest/pkg/tool/machines/model/constant"
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/manual"
)

type NextCustomAction struct {
	action string
}

func NewNextCustomAction(action string) *NextCustomAction {
	return &NextCustomAction{
		action: action,
	}
}

func (a *NextCustomAction) Action() string {
	return a.action
}

func (a *NextCustomAction) Do(_ manual.ManualConfig, ctx *context.MachineContext, event manual.ManualEvent) (string, error) {
	if !ctx.MachineStatus.IsValidStatus(event.NextStatus) {
		ctx.Log().WithFields(logrus.Fields{
			"nextStatus": event.NextStatus,
		}).Errorf("Task %s forbidden! the nextStatus is invalid", a.action)
		return "", errors2.NewWithArgs(code.TaskManualStatusInvalid, ctx.State(), constant.EventActionMap[a.action], event.NextStatus)
	}
	return event.NextStatus, nil
}
