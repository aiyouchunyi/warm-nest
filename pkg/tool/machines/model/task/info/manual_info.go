// Package info @Author larry
// @Date 2025/2/25 11:39
// @Desc

package info

import (
	"warm-nest/pkg/kinds/jsons"
)

type ManualInfo struct {
	Operator string           `gorm:"comment:操作人" json:"operator"`
	Time     int64            `gorm:"comment:操作时间" json:"time"`
	Action   string           `gorm:"comment:操作" json:"action"`
	Opinion  string           `gorm:"comment:操作意见" json:"opinion,omitempty"`
	Params   jsons.JSONObject `gorm:"comment:操作参数" json:"params,omitempty"`
}
