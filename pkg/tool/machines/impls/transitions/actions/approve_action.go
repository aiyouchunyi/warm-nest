// Package action @Author larry
// @Date 2025/5/26 13:38
// @Desc

package actions

import (
	"warm-nest/pkg/tool/machines/model/constant"
)

type ApproveAction struct {
	*NextAction
}

func NewApproveAction(nextStatus string) *ApproveAction {
	return &ApproveAction{
		NextAction: NewNextAction(constant.EventActionApprove, nextStatus),
	}
}
