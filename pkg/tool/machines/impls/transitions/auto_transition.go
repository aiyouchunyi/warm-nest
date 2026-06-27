// Package transitions @Author larry
// @Date 2024/4/29 17:16:00
// @Desc 任务节点状态工厂
package transitions

import (
	"warm-nest/pkg/tool/machines/model/context"
)

// IAutoTransition 状态转换器
type IAutoTransition interface {
	ITransition
	Before(ctx *context.MachineContext) error
	Execute(ctx *context.MachineContext) (string, error)
	Succ(ctx *context.MachineContext, nextStatus string) error
	Fail(ctx *context.MachineContext, nextStatus string, msg error) error
}

type DefaultAutoTransition struct {
}

func (t *DefaultAutoTransition) Before(_ *context.MachineContext) error {
	return nil
}

func (t *DefaultAutoTransition) Succ(_ *context.MachineContext, _ string) error {
	return nil
}

func (t *DefaultAutoTransition) Fail(_ *context.MachineContext, _ string, _ error) error {
	return nil
}
