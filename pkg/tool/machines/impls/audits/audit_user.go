// Package auth @Author larry
// @Date 2025/5/23 11:08
// @Desc

package audits

import (
	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/machines/code"
	"warm-nest/pkg/tool/machines/model/constant"
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/manual"
	"warm-nest/pkg/utils/slices"
)

// AuditUser 用户权限校验
func AuditUser(ctx *context.MachineContext, cfg manual.ManualConfig, event manual.ManualEvent) error {
	authUsers, ok := cfg.GetAuth(constant.AuthKindUser, event.Action)
	if !ok || len(authUsers) == 0 {
		return nil
	}
	if slices.Contain(authUsers, event.Operator.UserId) {
		return nil
	}
	return errors.NewWithArgs(code.UnauthorizedUser, ctx.State(), event.Operator.UserId, authUsers)
}
