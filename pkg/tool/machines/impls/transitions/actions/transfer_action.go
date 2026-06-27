// Package action @Author larry
// @Date 2025/4/9 16:28
// @Desc

package actions

import (
	"warm-nest/pkg/tool/machines/model/constant"
)

type TransferAction struct {
	*NextAction
}

func NewTransferAction(nextState string) *TransferAction {
	return &TransferAction{
		NewNextAction(constant.EventActionTransfer, nextState),
	}
}
