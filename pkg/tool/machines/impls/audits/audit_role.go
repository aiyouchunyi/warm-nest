// Package auth @Author larry
// @Date 2025/5/23 11:17
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

// AuditRole 角色权限校验
func AuditRole(ctx *context.MachineContext, cfg manual.ManualConfig, event manual.ManualEvent) error {
	authRoles, ok := cfg.GetAuth(constant.AuthKindRole, event.Action)
	if !ok || len(authRoles) == 0 {
		return nil
	}
	if slices.HasIntersection(authRoles, event.Operator.RoleIds) {
		return nil
	}
	return errors.NewWithArgs(code.UnauthorizedRole, ctx.State(), event.Operator.UserId, event.Operator.RoleIds, authRoles)
}
