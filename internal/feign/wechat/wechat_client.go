// Package wechat @Author larry
// @Date 2026/06/15
// @Desc 微信小程序 API 客户端（外部 feign）
//
// 微信返回非公司标准响应（{errcode,errmsg,...}），故用 ResultPtrV2 直接反序列化、
// 不走公司 RespCode 校验；errcode 由各 api 方法自行判定。
//
// feign 客户端范式：
//   - api 方法挂为 *Hook 成员函数（一方法一文件）；
//   - 鉴权收口到框架 SignRequest 机制：Client() 注册 WithSignature(injectAccessToken)，
//     需 access_token 的接口用 h.SignRequest(account)（触发 PreRequestHook 注入对应 token），
//     免 token 接口（code2session 用 appid/secret）用 h.Request()（SignAccount 为空，框架跳过注入）。
//   - 双主体：小程序与服务号是两个独立 access_token 主体（不同 appid/secret）。signAccount
//     既是"是否注入"的开关、也指明注入哪套 token：SignAccountMini=小程序（码/手机号/订阅消息），
//     SignAccountOfficial=服务号（模板消息）。token 缓存按 account 分桶（见 token_manager）。

package wechat

import (
	"warm-nest/pkg/tool/client"
	"warm-nest/pkg/tool/client/config"
	"warm-nest/pkg/tool/client/context"

	"warm-nest/internal/feign/wechat/model"
)

const ServiceName = "wechat"

// 签名账户：既触发 access_token 注入 hook，也指明注入哪套 token（小程序/服务号两主体）
const (
	SignAccountMini     = "mini"     // 小程序主体（AppId+Secret）：小程序码、手机号、订阅消息
	SignAccountOfficial = "official" // 服务号主体（OfficialAppId+OfficialSecret）：模板消息
)

// Hook 微信 API 客户端
type Hook struct {
	*client.HttpSession
}

func init() {
	client.RegisterDefaultDomain(config.DomainConfig{
		Name:   ServiceName,
		Domain: "https://api.weixin.qq.com",
	})
}

// Client 创建微信客户端（每次新建 Hook）。注册 access_token 注入为签名 hook：
// 走 h.SignRequest(SignAccountDefault) 的请求在发出前自动带上 access_token。
func Client(options ...context.Option) *Hook {
	// 用微信专属 RespCode 注册（Success 恒 true），而非 RegisterDefault（强制公司信封 result:true）。
	// 微信返回 {errcode/access_token/...} 无 result 字段，走 RegisterDefault 会被框架钩子误判失败
	// （日志渲染成 `-`）；errcode 由各 api 方法自行判定。见 resp.go 头注释。
	client.Register[model.WeChatRespCode](ServiceName, config.WithSignature(injectAccessToken))
	return &Hook{
		HttpSession: client.NewSession(ServiceName, options...),
	}
}
