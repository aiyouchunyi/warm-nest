// Package mapper @Author larry
// @Date 2026/06/18
// @Desc 服务号粉丝映射器

package mapper

import (
	"fmt"
	"sync"

	"warm-nest/pkg/tool/database/mysqls"

	"warm-nest/internal/model"
)

// FanMapper 服务号粉丝映射器
type FanMapper struct {
	mysqls.MysqlMapper[model.Fan]
}

var fanMapper *FanMapper
var fanOnce sync.Once

// GetFanMapper 获取服务号粉丝映射器单例
func GetFanMapper() *FanMapper {
	fanOnce.Do(func() {
		fanMapper = &FanMapper{
			MysqlMapper: mysqls.NewMapper[model.Fan](),
		}
	})
	return fanMapper
}

// GetByUnionId 按 unionId 查粉丝记录（找不到返回 nil,nil）
func (m *FanMapper) GetByUnionId(unionId string) (*model.Fan, error) {
	var fan model.Fan
	err := m.Model().Where("union_id = ?", unionId).First(&fan).Error
	if err != nil {
		if mysqls.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get fan by union %s: %w", unionId, err)
	}
	return &fan, nil
}

// Create 新建粉丝记录（首次关注，service 在 GetByUnionId 查空后调用）
func (m *FanMapper) Create(fan *model.Fan) error {
	if err := m.MysqlMapper.Create(fan); err != nil {
		return fmt.Errorf("create fan %s: %w", fan.UnionId, err)
	}
	return nil
}

// Save 更新已有粉丝记录（model 须带主键，由 GetByUnionId 取回后改字段再存）
func (m *FanMapper) Save(fan *model.Fan) error {
	if err := m.MysqlMapper.Save(fan); err != nil {
		return fmt.Errorf("save fan %s: %w", fan.UnionId, err)
	}
	return nil
}
