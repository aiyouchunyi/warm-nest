// Package audits @Author larry
// @Date 2025/12/12 09:14
// @Desc

package audits

import (
	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/machines/code"
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/manual"
	"warm-nest/pkg/utils/maps"
)

// AuditAction 检查事件动作
func AuditAction(ctx *context.MachineContext, cfg manual.ManualConfig, event manual.ManualEvent) error {
	if !(cfg.AllowAction(event.Action)) {
		ctx.Log().Errorf("The manual action[%s] is not config Action [%v]", event.Action, maps.Keys(cfg.ActionAuths))
		return errors.NewWithArgs(code.TaskActionForbidden, ctx.State(), event.Action, maps.Keys(cfg.ActionAuths))
	}
	return nil
}
