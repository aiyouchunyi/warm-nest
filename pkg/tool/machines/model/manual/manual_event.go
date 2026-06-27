// Package manual @Author larry
// @Date 2025/3/7 10:50
// @Desc

package manual

import (
	"warm-nest/pkg/app/web/session"
	"warm-nest/pkg/kinds/jsons"
	"warm-nest/pkg/tool/machines/model/dto"
)

type ManualEvent struct {
	Operator   session.User     `gorm:"comment:操作人信息" json:"operator"`
	Action     string           `gorm:"comment:操作" json:"action"`
	Opinion    string           `gorm:"comment:操作意见" json:"opinion,omitempty"`
	NextStatus string           `gorm:"comment:下个状态" json:"nextStatus,omitempty"`
	Params     jsons.JSONObject `gorm:"comment:操作参数" json:"params,omitempty"`
}

func NewEvent(req dto.MachineTaskReq) ManualEvent {
	return ManualEvent{
		Operator:   req.User,
		Action:     req.Action,
		Opinion:    req.Opinion,
		NextStatus: req.NextStatus,
		Params:     req.Params,
	}
}
