// Package model @Author larry
// @Date 2026/06/24
// @Desc 微信响应码适配（框架 client 的 OnAfterResponse 校验需要一个 IRespCode 实现）
//
// 为什么需要：框架 client.Register[R] 会注册 ParseErr[R] 作为 OnAfterResponse 钩子，对**每个**
// 响应（含 HTTP 200）做 R.Success() 校验，不成功就返回 CodeError。默认的 model.RespCode.Success()
// 读公司信封字段 `result:true`，而微信返回的是 {access_token/errcode/...}，根本没有 `result` 字段
// → 任何真实微信 200 响应都会被判失败、返回空 CodeError（日志里渲染成 `-`）。
//
// 这正是之前 mock=false 后 WeChatTokenRefresh 报 "fetch wechat access_token (mini): -" 的根因：
// 微信明明返回了合法 token，却被框架信封校验拦下。mock=true 时旁路了真实调用故一直没暴露。
//
// 修法（feign 错误码注册范式）：Client() 用 client.Register[WeChatRespCode] 替代
// RegisterDefault。微信不使用公司信封，各 api 方法已各自按 errcode 判定（token_manager 查 ErrCode!=0、
// template/subscribe notifier 区分 40003/43004 soft-skip），故框架层 Success() 恒 true，把 errcode
// 处置权完全交回各方法，框架钩子不再误拦。
package model

import "strconv"

// WeChatRespCode 框架响应钩子用的微信响应码占位。
// 微信不走公司信封：成功/失败由各 api 方法读 errcode 自行判定，框架层一律放行（Success 恒 true）。
type WeChatRespCode struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// Success 恒 true：框架层不拦微信响应，errcode 处置交回各 api 方法（见本文件头注释）。
func (WeChatRespCode) Success() bool { return true }

// GetCode 返回微信 errcode（框架记录用；本服务各方法自行判 errcode，不依赖此值）。
func (r WeChatRespCode) GetCode() string {
	if r.ErrCode == 0 {
		return ""
	}
	return strconv.Itoa(r.ErrCode)
}

// GetMessage 返回微信 errmsg。
func (r WeChatRespCode) GetMessage() string { return r.ErrMsg }
