// Package mapper @Author larry
// @Date 2026/06/22
// @Desc 通知路由配置映射器（PRD §5.2 按场景配渠道）

package mapper

import (
	"fmt"
	"sync"

	"warm-nest/pkg/tool/database/mysqls"

	"warm-nest/internal/model"
)

// NotifyRouteMapper 通知路由配置映射器
type NotifyRouteMapper struct {
	mysqls.MysqlMapper[model.NotifyRoute]
}

var notifyRouteMapper *NotifyRouteMapper
var notifyRouteOnce sync.Once

// GetNotifyRouteMapper 获取通知路由配置映射器单例
func GetNotifyRouteMapper() *NotifyRouteMapper {
	notifyRouteOnce.Do(func() {
		notifyRouteMapper = &NotifyRouteMapper{
			MysqlMapper: mysqls.NewMapper[model.NotifyRoute](),
		}
	})
	return notifyRouteMapper
}

// GetByScene 按场景查路由配置（找不到返回 nil,nil，由上层走默认渠道兜底）
func (m *NotifyRouteMapper) GetByScene(scene string) (*model.NotifyRoute, error) {
	var r model.NotifyRoute
	err := m.Model().Where("scene = ?", scene).First(&r).Error
	if err != nil {
		if mysqls.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get notify route by scene %s: %w", scene, err)
	}
	return &r, nil
}

// ListAll 列出全部路由配置（管理后台用，按场景升序）
func (m *NotifyRouteMapper) ListAll() ([]model.NotifyRoute, error) {
	var list []model.NotifyRoute
	if err := m.Model().Order("scene ASC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list all notify route: %w", err)
	}
	return list, nil
}

// UpdateRouteByScene 按场景更新渠道/模板ID/跳转页/启用态（管理后台改配置用，只更新这四列）
func (m *NotifyRouteMapper) UpdateRouteByScene(scene, channel, templateId, pagePath string, enabled bool) error {
	err := m.Model().Where("scene = ?", scene).Updates(map[string]interface{}{
		"channel":     channel,
		"template_id": templateId,
		"page_path":   pagePath,
		"enabled":     enabled,
	}).Error
	if err != nil {
		return fmt.Errorf("update notify route by scene %s: %w", scene, err)
	}
	return nil
}
