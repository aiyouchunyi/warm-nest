// Package api @Author larry
// @Date 2026/06/22
// @Desc 通知路由配置管理接口（后台：查全部 / 改某场景渠道，PRD §5.2）

package api

import (
	"sync"

	"warm-nest/pkg/app/web/session"

	"warm-nest/internal/service"
)

// ApiNotifyRoute 通知路由配置管理接口
type ApiNotifyRoute struct{}

var apiNotifyRoute *ApiNotifyRoute
var apiNotifyRouteOnce sync.Once

// GetApiNotifyRoute 获取通知路由配置管理接口单例
func GetApiNotifyRoute() *ApiNotifyRoute {
	apiNotifyRouteOnce.Do(func() {
		apiNotifyRoute = &ApiNotifyRoute{}
	})
	return apiNotifyRoute
}

// ListNotifyRouteReq 查全部路由配置请求（无入参）
type ListNotifyRouteReq struct {
	session.Session
}

// List 查全部通知路由配置（管理后台展示场景→渠道映射）
func (a *ApiNotifyRoute) List(req ListNotifyRouteReq) (interface{}, error) {
	return service.GetNotifyRouteService().ListRoutes()
}

// SaveNotifyRouteReq 改某场景渠道配置请求
type SaveNotifyRouteReq struct {
	session.Session
	Scene      string `json:"scene" validate:"required"`   // 通知场景（内置枚举，须已 seed）
	Channel    string `json:"channel" validate:"required"` // 下发渠道[subscribe/template/sms/log]
	TemplateId string `json:"templateId"`                  // 该渠道模板ID（log 渠道可空）
	PagePath   string `json:"pagePath"`                    // 模板消息点击跳转的小程序页（template 渠道用，空则代码兜底）
	Enabled    bool   `json:"enabled"`                     // 是否外推
}

// Save 改某场景的渠道/模板ID/跳转页/启用态（运营改配置，不新建场景）
func (a *ApiNotifyRoute) Save(req SaveNotifyRouteReq) (interface{}, error) {
	if err := service.GetNotifyRouteService().SaveRoute(req.Scene, req.Channel, req.TemplateId, req.PagePath, req.Enabled); err != nil {
		return nil, err
	}
	return map[string]bool{"ok": true}, nil
}
