// Package action @Author larry
// @Date 2025/4/9 11:09
// @Desc

package actions

import (
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/manual"
)

type NextAction struct {
	action    string
	nextState string
}

func NewNextAction(action string, nextState string) *NextAction {
	return &NextAction{
		action:    action,
		nextState: nextState,
	}
}

func (a *NextAction) Action() string {
	return a.action
}

func (a *NextAction) Do(_ manual.ManualConfig, _ *context.MachineContext, _ manual.ManualEvent) (string, error) {
	return a.nextState, nil
}
