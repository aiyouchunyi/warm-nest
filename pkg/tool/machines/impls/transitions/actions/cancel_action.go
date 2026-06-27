// Package action @Author larry
// @Date 2025/4/8 19:21
// @Desc

package actions

import (
	"warm-nest/pkg/tool/machines/model/constant"
)

type CancelAction struct {
	*NextAction
}

func NewCancelAction(nextStatus string) *CancelAction {
	return &CancelAction{
		NextAction: NewNextAction(constant.EventActionCancel, nextStatus),
	}
}
