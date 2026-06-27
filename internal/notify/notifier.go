// Package notify @Author larry
// @Date 2026/06/22
// @Desc 通知外推：工厂 + 策略 + 场景路由（PRD §5.2 上层认场景、底层换渠道）
//
// 分层：业务（message_service）只调 Dispatch(scene, receiver, params)，不认渠道。
//
//	Dispatch → ChannelRouter 按 notify_route 表把 scene 解析成 (channel, templateId)
//	         → registry 按 channel 取 Notifier 策略 → Push 落底层 API。
//
// 新增渠道（如短信）= 加一个 Notifier 实现 + 在 buildRegistry 注册一行，Dispatch/router/上层不动（开闭）。
//
// 仅负责「外部渠道外推」这一通道；小程序内消息流（必达兜底）在 message_service 通道①，与本包无关。
package notify

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/kinds/jsons"
)

// PushContext 单次外推上下文（接收方 + 该渠道模板ID + 渲染参数）
type PushContext struct {
	Scene          string           // 业务场景（notifier 按场景把语义 params 翻译成各自渠道的模板字段）
	ReceiverUserId string           // 接收方业务用户ID（由 notifier 自行解析成各渠道所需的 openid/手机号）
	TemplateId     string           // 该渠道对应的模板ID（空则 notifier no-op，不外推）
	PagePath       string           // 服务号模板消息点击跳转的小程序页（template 渠道用，空则按场景兜底）
	Params         jsons.JSONObject // 语义渲染参数（key 为业务语义名，由 notifier 翻译成平台模板字段）
}

// Notifier 通知渠道策略：一个渠道一个实现，自报 Channel() 名供注册表索引
type Notifier interface {
	Channel() string            // 渠道标识（对应 model.NotifyChannel*）
	Push(ctx PushContext) error // 外推一条；soft-skip（缺 openid/模板）记日志返 nil，真失败返 err
}

var (
	registry     map[string]Notifier
	registryOnce sync.Once
)

// buildRegistry 装配所有渠道策略（新增渠道在此注册一行）
func buildRegistry() map[string]Notifier {
	reg := make(map[string]Notifier)
	for _, n := range []Notifier{
		newSubscribeNotifier(), // 小程序订阅消息
		newTemplateNotifier(),  // 服务号模板消息
		newLogNotifier(),       // 只打日志不真发（零微信配置跑通业务链路）
		// 扩展点：newSmsNotifier() —— 实现 Notifier 后在此加一行即可
	} {
		reg[n.Channel()] = n
	}
	return reg
}

// getNotifier 按渠道取策略
func getNotifier(channel string) (Notifier, bool) {
	registryOnce.Do(func() { registry = buildRegistry() })
	n, ok := registry[channel]
	return n, ok
}

// Dispatch 按场景外推（best-effort 入口）：解析渠道 → 取策略 → Push。
// 返回 error 供调用方记日志；内消息流由调用方先行落库，本调用失败不影响必达。
func Dispatch(scene, receiverUserId string, params jsons.JSONObject) error {
	channel, templateId, pagePath, enabled, err := defaultRouter().Resolve(scene)
	if err != nil {
		return fmt.Errorf("notify resolve scene %s: %w", scene, err)
	}
	if !enabled {
		// 该场景配置为不外推（仅内消息流）
		return nil
	}
	n, ok := getNotifier(channel)
	if !ok {
		logrus.WithFields(logrus.Fields{"scene": scene, "channel": channel}).
			Warn("notify dispatch: no notifier registered for channel")
		return fmt.Errorf("no notifier for channel %s", channel)
	}
	return n.Push(PushContext{Scene: scene, ReceiverUserId: receiverUserId, TemplateId: templateId, PagePath: pagePath, Params: params})
}
