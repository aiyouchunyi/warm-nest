// Package dto @Author larry
// File machine_task.go
// @Date 2024/5/8 20:52:00
// @Desc 状态机信息
package dto

import (
	"warm-nest/pkg/kinds/jsons"
	"warm-nest/pkg/tool/machines/model/task/info"
)

type MachineTaskDTO struct {
	Status      string                       `gorm:"comment:状态" json:"status"`
	StatusInfos jsons.Array[info.StatusInfo] `gorm:"comment:状态信息路径[状态值]" json:"statusInfos,omitempty"`
}
