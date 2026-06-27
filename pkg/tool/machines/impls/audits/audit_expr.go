// Package auth @Author larry
// @Date 2025/5/23 11:12
// @Desc

package audits

import (
	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/machines/code"
	"warm-nest/pkg/tool/machines/model/constant"
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/manual"
	"warm-nest/pkg/tool/machines/model/task/info"
	"warm-nest/pkg/utils/reflects"
	"warm-nest/pkg/utils/slices"
)

// AuditExpr 表达式权限校验
func AuditExpr(ctx *context.MachineContext, cfg manual.ManualConfig, event manual.ManualEvent) error {
	authExprs, ok := cfg.GetAuth(constant.AuthKindExpr, event.Action)
	if !ok || len(authExprs) == 0 {
		return nil
	}
	for _, expr := range authExprs {
		switch expr {
		case constant.AuditNotCreator:
			createdId, err := reflects.GetModelField(ctx.Task, "CreatedId")
			if err != nil {
				ctx.Log().WithError(err).Errorf("Get Model Created Id failed!")
				return errors.NewWithArgs(code.UnauthorizedCheckErr, ctx.State(), err)
			}
			if createdId != event.Operator.UserId {
				return nil
			}
			return errors.NewWithArgs(code.UnauthorizedDisCreator, ctx.State(), event.Operator.UserId, createdId)
		case constant.AuditCreator:
			createdId, err := reflects.GetModelField(ctx.Task, "CreatedId")
			if err != nil {
				ctx.Log().WithError(err).Errorf("Get Model Created Id failed!")
				return errors.NewWithArgs(code.UnauthorizedCheckErr, ctx.State(), err)
			}
			if createdId == event.Operator.UserId {
				return nil
			}
			return errors.NewWithArgs(code.UnauthorizedCreator, ctx.State(), event.Operator.UserId, createdId)
		case constant.AuditNotPrevious:
			preNode := ctx.MachineTask.PreNode()
			operators := slices.Map(preNode.Manuals, func(info info.ManualInfo) string {
				return info.Operator
			})
			if !slices.Contains(operators, event.Operator.UserId) {
				return nil
			}
			return errors.NewWithArgs(code.UnauthorizedDisPrevious, ctx.State(), event.Operator.UserId, operators)
		default:
			return errors.NewWithArgs(code.UnauthorizedExprUNSupport, ctx.State(), event.Operator.UserId, expr)
		}
	}
	return nil
}
