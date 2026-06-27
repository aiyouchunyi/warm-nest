// Package database @Author larry
// File db_model.go
// @Date 2024/5/22 14:12:00
// @Desc 数据库系统字段
package database

import (
	"gorm.io/gorm"
)

// Identify 数据库唯一标识
type Identify struct {
	ID string `gorm:"type:bigint unsigned;primary;autoIncrement;comment:唯一标识" json:"id"  bson:"_id,omitempty"`
}

// SystemField 系统字段
type SystemField struct {
	CreatedId string         `gorm:"size:64;comment:创建者" json:"createdId,omitempty" bson:"createdId,omitempty"`
	CreatedAt int64          `gorm:"index:i_create;autoCreateTime:milli;comment:创建时间" json:"createdAt" bson:"createdAt"`
	UpdatedId string         `gorm:"size:64;comment:更新者" json:"updatedId,omitempty" bson:"updatedId,omitempty"`
	UpdatedAt int64          `gorm:"index:i_update;autoUpdateTime:milli;comment:更新时间" json:"updatedAt" bson:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"comment:删除时间" json:"-"  bson:"-"`
}

type Model interface {
	TableName() string
}
