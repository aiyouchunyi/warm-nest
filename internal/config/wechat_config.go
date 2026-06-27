// Package config @Author larry
// @Date 2026/06/15
// @Desc 微信小程序配置
//
// secret 加载入口说明（评审修正）：configs.MustLoad 只读 toml+env，不拉 AWS。
// 若 AppSecret 走 AWS Secrets Manager → 改用 configs.MustLoadSecret；
// 若由部署平台注入环境变量（${WECHAT_SECRET} 占位）→ 用 MustLoad。按真实注入方式二选一。

package config

import (
	"sync"

	"warm-nest/pkg/app/configs"
)

// WeChatConfig 微信小程序配置（app.toml [wechat] 段）
type WeChatConfig struct {
	AppId  string `toml:"appId"`  // 小程序 AppID
	Secret string `toml:"secret"` // AppSecret（部署平台从 AWS 注入到 env）
	// 注：各类消息的模板ID 不再在此配置——已下沉到 notify_route 表按场景随渠道维护（PRD §5.2），
	// 由运营在 /warm-nest/admin/v1/notify-route 管理接口按真实平台模板ID填写。
	MiniprogramState string `toml:"miniprogramState" default:"formal"` // formal|trial|develop
	TokenRefreshSec  int64  `toml:"tokenRefreshSec" default:"5400"`    // access_token 刷新间隔秒（< 7200 留余量）
	// —— 服务号（公众平台）配置：与小程序不同主体，用于 §8.2 关注/取关事件回调验签 + 模板消息下发 ——
	// 服务号关注事件推 XML 到 OfficialCallbackToken 配置的回调 URL，按本 Token 做 sha1 签名校验。
	OfficialAppId         string `toml:"officialAppId"`         // 服务号 AppID（与小程序 AppId 不同）
	OfficialSecret        string `toml:"officialSecret"`        // 服务号 AppSecret（拉服务号 access_token，发模板消息用；template 渠道必填）
	OfficialCallbackToken string `toml:"officialCallbackToken"` // 服务号消息推送回调验签 Token（公众平台后台配置）
	// Mock 本地联调旁路：true 时 feign 不真调微信，code2session 返回固定假 openid、推送跳过。
	// 合法环境分支（同框架 IsLocal），生产 toml 不开即等于无。默认 false。
	Mock bool `toml:"mock" default:"false"`
}

var weChatConf WeChatConfig
var weChatConfOnce sync.Once

// WeChatConf 获取微信配置单例
func WeChatConf() WeChatConfig {
	weChatConfOnce.Do(func() {
		weChatConf = configs.MustLoad[WeChatConfig]("wechat")
	})
	return weChatConf
}
