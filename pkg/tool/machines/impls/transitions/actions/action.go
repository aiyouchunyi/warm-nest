// Package action @Author larry
// @Date 2025/4/8 19:13
// @Desc

package actions

import (
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/manual"
)

type Action interface {
	Action() string
	Do(config manual.ManualConfig, ctx *context.MachineContext, event manual.ManualEvent) (string, error)
}
