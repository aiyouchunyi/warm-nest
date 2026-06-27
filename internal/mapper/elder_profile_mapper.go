// Package mapper @Author larry
// @Date 2026/06/15
// @Desc 被守护人档案映射器

package mapper

import (
	"fmt"
	"sync"

	"warm-nest/pkg/tool/database/mysqls"

	"warm-nest/internal/model"
)

// ElderProfileMapper 被守护人档案映射器
type ElderProfileMapper struct {
	mysqls.MysqlMapper[model.ElderProfile]
}

var elderProfileMapper *ElderProfileMapper
var elderProfileOnce sync.Once

// GetElderProfileMapper 获取被守护人档案映射器单例
func GetElderProfileMapper() *ElderProfileMapper {
	elderProfileOnce.Do(func() {
		elderProfileMapper = &ElderProfileMapper{
			MysqlMapper: mysqls.NewMapper[model.ElderProfile](),
		}
	})
	return elderProfileMapper
}

// ListByRemindTime 查约定提醒时间为 remindTime(HH:mm) 的全部被守护人（未打卡提醒任务用）
func (m *ElderProfileMapper) ListByRemindTime(remindTime string) ([]model.ElderProfile, error) {
	var list []model.ElderProfile
	err := m.Model().Where("remind_time = ?", remindTime).Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list elder profile by remindTime %s: %w", remindTime, err)
	}
	return list, nil
}

// ListWithRemindTime 查全部已设提醒时间的被守护人档案（未打卡提醒轮询任务用，按时刻窗口在内存里筛）
func (m *ElderProfileMapper) ListWithRemindTime() ([]model.ElderProfile, error) {
	var list []model.ElderProfile
	err := m.Model().Where("remind_time <> ''").Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list elder profile with remind time: %w", err)
	}
	return list, nil
}

// GetByUserId 按被守护人用户ID查档案（1:1，找不到返回 nil,nil）
func (m *ElderProfileMapper) GetByUserId(userId string) (*model.ElderProfile, error) {
	var profile model.ElderProfile
	err := m.Model().Where("user_id = ?", userId).First(&profile).Error
	if err != nil {
		if mysqls.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get elder profile by userId %s: %w", userId, err)
	}
	return &profile, nil
}
