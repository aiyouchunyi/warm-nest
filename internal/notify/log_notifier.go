// Package notify @Author larry
// @Date 2026/06/22
// @Desc 日志渠道策略：只打日志、不真发（PRD §5.2 测试/联调用）
//
// 用途：无任何微信配置（小程序/服务号 secret、模板ID 都没有）时，把通知路由的 channel
// 配成 "log"，即可跑通完整业务链路——内消息流照常落库（通道①），外推这一步只在日志里
// 打出「本该推什么给谁」，不查 openid、不调微信。便于在干净环境做端到端业务测试。

package notify

import (
	"github.com/sirupsen/logrus"

	"warm-nest/internal/model"
)

// logNotifier 日志渠道（只记不发）
type logNotifier struct{}

func newLogNotifier() *logNotifier { return &logNotifier{} }

// Channel 渠道标识
func (n *logNotifier) Channel() string { return model.NotifyChannelLog }

// Push 只打一条结构化日志，永不报错（测试链路不应因外推失败而中断）
func (n *logNotifier) Push(ctx PushContext) error {
	logrus.WithFields(logrus.Fields{
		"channel":    n.Channel(),
		"receiverId": ctx.ReceiverUserId,
		"templateId": ctx.TemplateId,
		"params":     ctx.Params,
	}).Info("notify[log]: would push notification (log-only channel, not actually sent)")
	return nil
}
