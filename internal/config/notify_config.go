// Package config @Author larry
// @Date 2026/06/22
// @Desc 通知配置（app.toml [notify] 段，PRD §5.2 按场景配渠道）

package config

import (
	"sync"

	"warm-nest/pkg/app/configs"
)

// NotifyConfig 通知配置
type NotifyConfig struct {
	// DefaultChannel 未配置场景的兜底渠道 + migrate seed 初值。
	// 可选 subscribe(小程序订阅)/template(服务号模板)/log(只打日志不真发)/sms(未实现)。
	// 默认 log：无微信配置即可零配置跑通完整业务链路；服务号就绪后改 template（PRD §5.2 主推）。
	DefaultChannel string `toml:"defaultChannel" default:"log"`
}

var notifyConf NotifyConfig
var notifyConfOnce sync.Once

// NotifyConf 获取通知配置单例
func NotifyConf() NotifyConfig {
	notifyConfOnce.Do(func() {
		notifyConf = configs.MustLoad[NotifyConfig]("notify")
	})
	return notifyConf
}
