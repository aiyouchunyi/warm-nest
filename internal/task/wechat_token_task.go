// Package task @Author larry
// @Date 2026/06/15
// @Desc 微信 access_token 定时刷新（内存缓存，单实例）

package task

import (
	"sync"

	"github.com/sirupsen/logrus"

	"warm-nest/internal/config"
	"warm-nest/internal/feign/wechat"
)

// WeChatTokenTask 微信 token 刷新任务
type WeChatTokenTask struct{}

var weChatTokenTask *WeChatTokenTask
var weChatTokenTaskOnce sync.Once

// GetWeChatTokenTask 获取 token 刷新任务单例
func GetWeChatTokenTask() *WeChatTokenTask {
	weChatTokenTaskOnce.Do(func() {
		weChatTokenTask = &WeChatTokenTask{}
	})
	return weChatTokenTask
}

// Refresh 刷新 access_token（小程序必刷；服务号仅在配置了 officialSecret 时刷，未配置跳过不报错）
func (t *WeChatTokenTask) Refresh(args ...any) error {
	if err := wechat.Refresh(wechat.SignAccountMini); err != nil {
		return err
	}
	// 服务号 token 仅 template 渠道用；未配 officialSecret 时跳过，避免无谓失败刷屏
	if config.WeChatConf().OfficialSecret == "" {
		return nil
	}
	if err := wechat.Refresh(wechat.SignAccountOfficial); err != nil {
		logrus.WithError(err).Warn("refresh official access_token failed")
	}
	return nil
}
