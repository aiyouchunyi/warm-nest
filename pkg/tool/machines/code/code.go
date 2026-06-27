// Package code @Author larry
// File code.go
// @Date 2024/5/7 15:37:00
// @Desc 状态机错误码
package code

import (
	"warm-nest/pkg/app/errors"
)

var (
	MachineAutoTransitionNotFound   = errors.NewCode("00000400", "%s 自动执行器未找到")
	MachineManualTransitionNotFound = errors.NewCode("00000401", "%s 手动执行器未找到")

	MachineBeforeActionException = errors.NewCode("00000410", "%s 前置行为异常:{%v}")
	MachineTargetActionException = errors.NewCode("00000411", "%s 目标行为异常:{%v}")
	MachineFailActionException   = errors.NewCode("00000412", "%s 失败行为异常:{%v}")
	MachineSuccActionException   = errors.NewCode("00000413", "%s 成功行为异常:{%v}")
	MachineFinActionException    = errors.NewCode("00000414", "%s 完成行为异常:{%v}")

	TaskNotFound       = errors.NewCode("00000430", "%s 查询异常:{%v}")
	TaskAlreadyLocked  = errors.NewCode("00000431", "%s 已锁定")
	TaskLockException  = errors.NewCode("00000432", "%s 锁定异常:{%v}")
	TaskSyncException  = errors.NewCode("00000433", "%s 同步异常:{%v}")
	TaskExecuteWaitMax = errors.NewCode("00000434", "%s 重复执行已达上限 %d")

	TaskManualException = errors.NewCode("00000460", "%s 操作异常")

	TaskManualForbidden      = errors.NewCode("00000461", "%s 禁止操作")
	TaskActionForbidden      = errors.NewCode("00000461", "%s 动作 %s 禁止操作, 期望动作:%v")
	TaskActionNotFound       = errors.NewCode("00000462", "%s 动作 %s 执行器未找到")
	TaskRollbackForbidden    = errors.NewCode("00000463", "%s 禁止回退")
	TaskForceStatusUnchanged = errors.NewCode("00000471", "%s 强制状态未发生变化")
	TaskManualStatusInvalid  = errors.NewCode("00000473", "%s %s状态 %s 无效")

	UnauthorizedOperatorLoss  = errors.NewCode("00000480", "%s 操作用户缺失")
	UnauthorizedUser          = errors.NewCode("00000481", "%s %s 操作未授权!允许操作人 %v")
	UnauthorizedDisAdmin      = errors.NewCode("00000482", "%s %s 操作未授权!只允许管理用户操作")
	UnauthorizedRole          = errors.NewCode("00000483", "%s %s-%v 操作未授权!允许操作角色 %v")
	UnauthorizedExprUNSupport = errors.NewCode("00000484", "%s %s 操作未授权!不满足权限规则 %v")
	UnauthorizedDisCreator    = errors.NewCode("00000485", "%s %s 操作未授权!不符合非创建人规则 %s")
	UnauthorizedCreator       = errors.NewCode("00000486", "%s %s 操作未授权!不符合创建人规则 %s")
	UnauthorizedDisPrevious   = errors.NewCode("00000487", "%s %s操作未授权!不符合非前置操作人规则 %v")
	UnauthorizedCheckErr      = errors.NewCode("00000488", "%s %s 操作授权检查失败:{%v}")
)
