// Package dto @Author larry
// @Date 2025/5/26 16:12
// @Desc

package dto

import (
	"warm-nest/pkg/app/web/session"
	"warm-nest/pkg/kinds/jsons"
)

type MachineTasksReq struct {
	session.Session
	TaskIds    []string         `gorm:"comment:任务ID列表" json:"taskIds" validate:"required"`
	Action     string           `gorm:"comment:操作类型" json:"action" validate:"required"`
	NextStatus string           `gorm:"comment:下个状态" json:"nextStatus,omitempty"`
	Opinion    string           `gorm:"comment:操作意见" json:"opinion,omitempty"`
	Params     jsons.JSONObject `gorm:"comment:操作参数" json:"params,omitempty"`
}
