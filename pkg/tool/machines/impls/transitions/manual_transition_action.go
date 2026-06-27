// Package transitions @Author larry
// @Date 2025/4/9 10:51
// @Desc

package transitions

import (
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/machines/code"
	"warm-nest/pkg/tool/machines/impls/transitions/actions"
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/manual"
	"warm-nest/pkg/utils/maps"
)

type ManualTransitionAction struct {
	DefaultManualTransition
	state        string
	manualConfig manual.ManualConfig
	acts         map[string]actions.Action
}

func NewDefaultTransitionAction(state string, actions ...actions.Action) *ManualTransitionAction {
	return NewTransitionAction(state, manual.NewManualConfig(), actions...)
}

func NewTransitionAction(state string, manualConfig manual.ManualConfig, acts ...actions.Action) *ManualTransitionAction {
	transitionAction := &ManualTransitionAction{
		state:        state,
		manualConfig: manualConfig,
		acts:         make(map[string]actions.Action),
	}
	for _, act := range acts {
		transitionAction.acts[act.Action()] = act
	}
	transitionAction.manualConfig.AddAction(maps.Keys(transitionAction.acts)...)
	return transitionAction
}

func (t *ManualTransitionAction) State() string {
	return t.state
}

func (t *ManualTransitionAction) Config() manual.ManualConfig {
	return t.manualConfig
}

func (t *ManualTransitionAction) Execute(ctx *context.MachineContext, event manual.ManualEvent) (string, error) {
	act, ok := t.acts[event.Action]
	if !ok {
		ctx.Log().WithFields(logrus.Fields{
			"action": event.Action,
		}).Errorf("action not found excutor!")
		return "", errors.NewWithArgs(code.TaskActionNotFound, ctx.State(), event.Action)
	}
	return act.Do(t.Config(), ctx, event)
}
