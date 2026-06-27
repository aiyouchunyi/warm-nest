// Package action @Author larry
// @Date 2025/5/26 13:38
// @Desc

package actions

import (
	"warm-nest/pkg/tool/machines/model/constant"
)

type RejectAction struct {
	*NextAction
}

func NewRejectAction(nextStatus string) *RejectAction {
	return &RejectAction{
		NextAction: NewNextAction(constant.EventActionReject, nextStatus),
	}
}
