// Package service @Author larry
// @Date 2026/06/22
// @Desc 通知路由配置服务（管理后台：查全部 / 改某场景渠道与模板ID，PRD §5.2）

package service

import (
	"fmt"
	"sync"

	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
)

// NotifyRouteService 通知路由配置服务
type NotifyRouteService struct {
	notifyRouteMapper *mapper.NotifyRouteMapper
}

var notifyRouteService *NotifyRouteService
var notifyRouteServiceOnce sync.Once

// GetNotifyRouteService 获取通知路由配置服务单例
func GetNotifyRouteService() *NotifyRouteService {
	notifyRouteServiceOnce.Do(func() {
		notifyRouteService = &NotifyRouteService{
			notifyRouteMapper: mapper.GetNotifyRouteMapper(),
		}
	})
	return notifyRouteService
}

// ListRoutes 查全部路由配置（管理后台展示）
func (s *NotifyRouteService) ListRoutes() ([]model.NotifyRoute, error) {
	return s.notifyRouteMapper.ListAll()
}

// SaveRoute 改某场景的渠道/模板ID/跳转页/启用态。校验场景存在（须先 seed）+ 渠道合法，
// 不接受前端新建场景（场景是代码内置枚举，由 migrate seed）。
func (s *NotifyRouteService) SaveRoute(scene, channel, templateId, pagePath string, enabled bool) error {
	// 渠道合法性复用 model.IsValidNotifyChannel（与 config 兜底校验同一白名单；sms 未实现 Notifier 不可配）
	if !model.IsValidNotifyChannel(channel) {
		return fmt.Errorf("invalid notify channel: %s", channel)
	}
	existing, err := s.notifyRouteMapper.GetByScene(scene)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("notify scene not found (not a built-in scene): %s", scene)
	}
	return s.notifyRouteMapper.UpdateRouteByScene(scene, channel, templateId, pagePath, enabled)
}
