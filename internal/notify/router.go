// Package notify @Author larry
// @Date 2026/06/22
// @Desc 场景→渠道路由（读 notify_route 表，未配/异常走默认渠道兜底）

package notify

import (
	"sync"

	"github.com/sirupsen/logrus"

	"warm-nest/internal/config"
	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
)

// DefaultChannel 兜底渠道：读 config [notify].defaultChannel；配了非法值则硬兜底到 log（防全链路炸）。
// 供 router 未命中兜底 + migrate seed 初值共用，是「默认渠道」的唯一真相源。
func DefaultChannel() string {
	ch := config.NotifyConf().DefaultChannel
	if !model.IsValidNotifyChannel(ch) {
		logrus.WithField("configured", ch).Warn("invalid notify defaultChannel in config, hard fallback to log")
		return model.NotifyChannelLog
	}
	return ch
}

// ChannelRouter 把通知场景解析成 (渠道, 模板ID, 是否外推)
type ChannelRouter struct {
	notifyRouteMapper *mapper.NotifyRouteMapper
}

var (
	router     *ChannelRouter
	routerOnce sync.Once
)

// defaultRouter 路由单例
func defaultRouter() *ChannelRouter {
	routerOnce.Do(func() {
		router = &ChannelRouter{notifyRouteMapper: mapper.GetNotifyRouteMapper()}
	})
	return router
}

// Resolve 解析场景的下发渠道、模板ID 与跳转页。
//   - 表里有该场景配置：按配置返回（enabled=false 表示该场景不外推）；
//   - 表里没有（漏 seed / 新场景）：走默认渠道兜底、enabled=true、模板ID/跳转页空（notifier 会 soft-skip / 兜底）。
//
// 查表本身出错时也降级到默认渠道（记日志），不因配置读失败而中断业务推送链路。
func (r *ChannelRouter) Resolve(scene string) (channel, templateId, pagePath string, enabled bool, err error) {
	route, err := r.notifyRouteMapper.GetByScene(scene)
	if err != nil {
		logrus.WithError(err).WithField("scene", scene).Warn("notify route query failed, fallback to default channel")
		return DefaultChannel(), "", "", true, nil
	}
	if route == nil {
		logrus.WithField("scene", scene).Warn("notify route not configured, fallback to default channel")
		return DefaultChannel(), "", "", true, nil
	}
	return route.Channel, route.TemplateId, route.PagePath, route.Enabled, nil
}
