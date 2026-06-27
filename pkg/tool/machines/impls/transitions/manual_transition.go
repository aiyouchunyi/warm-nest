// Package transitions @Author larry
// @Date 2025/3/7 10:47
// @Desc

package transitions

import (
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/manual"
)

// IManualTransition 状态转换器
type IManualTransition interface {
	ITransition
	Config() manual.ManualConfig                                                                    // 状态转换器配置
	Before(ctx *context.MachineContext, event manual.ManualEvent) error                             // 操作前置处理
	Execute(ctx *context.MachineContext, event manual.ManualEvent) (string, error)                  // 操作执行
	Succ(ctx *context.MachineContext, event manual.ManualEvent, nextStatus string) error            // 操作成功后置处理
	Fail(ctx *context.MachineContext, event manual.ManualEvent, nextStatus string, msg error) error // 操作失败后置处理
}

type DefaultManualTransition struct {
}

// Config 状态转换器配置
func (t *DefaultManualTransition) Config() manual.ManualConfig {
	return manual.ManualConfig{}
}

// Before 操作前置处理
func (t *DefaultManualTransition) Before(_ *context.MachineContext, _ manual.ManualEvent) error {
	return nil
}

// Succ 操作成功后置处理
func (t *DefaultManualTransition) Succ(_ *context.MachineContext, _ manual.ManualEvent, _ string) error {
	return nil
}

// Fail 操作失败后置处理
func (t *DefaultManualTransition) Fail(_ *context.MachineContext, _ manual.ManualEvent, _ string, _ error) error {
	return nil
}
