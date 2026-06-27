// Package mapper @Author larry
// @Date 2026/06/15
// @Desc 守护关系映射器

package mapper

import (
	"fmt"
	"sync"

	"warm-nest/pkg/tool/database/mysqls"

	"warm-nest/internal/model"
)

// GuardianshipMapper 守护关系映射器
type GuardianshipMapper struct {
	mysqls.MysqlMapper[model.Guardianship]
}

var guardianshipMapper *GuardianshipMapper
var guardianshipOnce sync.Once

// GetGuardianshipMapper 获取守护关系映射器单例
func GetGuardianshipMapper() *GuardianshipMapper {
	guardianshipOnce.Do(func() {
		guardianshipMapper = &GuardianshipMapper{
			MysqlMapper: mysqls.NewMapper[model.Guardianship](),
		}
	})
	return guardianshipMapper
}

// ListByGuardian 查某子女守护的所有 ACTIVE 关系（子女视角：我守护谁）
func (m *GuardianshipMapper) ListByGuardian(guardianUserId string) ([]model.Guardianship, error) {
	var list []model.Guardianship
	err := m.Model().
		Where("guardian_user_id = ? AND status = ?", guardianUserId, model.GuardianshipStatusActive).
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list guardianship by guardian %s: %w", guardianUserId, err)
	}
	return list, nil
}

// ListByElder 查守护某老人的所有 ACTIVE 关系（老人视角：谁守护我）
func (m *GuardianshipMapper) ListByElder(elderUserId string) ([]model.Guardianship, error) {
	var list []model.Guardianship
	err := m.Model().
		Where("elder_user_id = ? AND status = ?", elderUserId, model.GuardianshipStatusActive).
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list guardianship by elder %s: %w", elderUserId, err)
	}
	return list, nil
}

// ListAllActive 列全部 ACTIVE 守护关系（预热提醒等全量扫描用；1 期一对一、量小，不分页）
func (m *GuardianshipMapper) ListAllActive() ([]model.Guardianship, error) {
	var list []model.Guardianship
	err := m.Model().
		Where("status = ?", model.GuardianshipStatusActive).
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list all active guardianship: %w", err)
	}
	return list, nil
}

// GetActive 查某 (子女,老人) 对的 ACTIVE 关系（绑定防重用，找不到返回 nil,nil）
func (m *GuardianshipMapper) GetActive(guardianUserId, elderUserId string) (*model.Guardianship, error) {
	var g model.Guardianship
	err := m.Model().
		Where("guardian_user_id = ? AND elder_user_id = ? AND status = ?",
			guardianUserId, elderUserId, model.GuardianshipStatusActive).
		First(&g).Error
	if err != nil {
		if mysqls.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get active guardianship %s->%s: %w", guardianUserId, elderUserId, err)
	}
	return &g, nil
}

// UpdateRelation 按 guardianshipId 更新关系称呼（问题1：relation 挂守护边，子女端可改）
func (m *GuardianshipMapper) UpdateRelation(guardianshipId, relation string) error {
	err := m.Model().
		Where("guardianship_id = ?", guardianshipId).
		Update("relation", relation).Error
	if err != nil {
		return fmt.Errorf("update guardianship relation %s: %w", guardianshipId, err)
	}
	return nil
}
