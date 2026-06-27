// Package database @Author larry
// File db_model_dto.go
// @Date 2024/5/22 14:12:00
// @Desc 数据库系统字段DTO
package database

// SystemFieldDTO 系统字段DTO
type SystemFieldDTO struct {
	ID          string `gorm:"comment:唯一标识" json:"id"`
	CreatedId   string `gorm:"comment:创建者" json:"createdId,omitempty"`
	CreatedName string `gorm:"comment:创建者名字" json:"createdName,omitempty"`
	CreatedAt   int64  `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedId   string `gorm:"comment:更新者" json:"updatedId,omitempty"`
	UpdatedName string `gorm:"comment:更新者名字" json:"updatedName,omitempty"`
	UpdatedAt   int64  `gorm:"comment:更新时间" json:"updatedAt"`
}
