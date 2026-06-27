// Package dto @Author larry
// @Date 2025/5/26 16:13
// @Desc

package dto

import (
	"warm-nest/pkg/app/web/session"
	"warm-nest/pkg/kinds/jsons"
)

type MachineTaskReq struct {
	session.Session
	TaskId     string           `gorm:"comment:任务ID" json:"taskId" validate:"required"`
	Action     string           `gorm:"comment:操作类型" json:"action"`
	NextStatus string           `gorm:"comment:下个状态" json:"nextStatus,omitempty"`
	Opinion    string           `gorm:"comment:操作意见" json:"opinion,omitempty"`
	Params     jsons.JSONObject `gorm:"comment:操作参数" json:"params,omitempty"`
}
