// Package model @Author larry
// @Date 2026/06/15
// @Desc 订阅消息下发 cgi-bin/message/subscribe/send 相关 DTO 与响应码
package model

// 微信订阅消息常见错误码
const (
	SubMsgOk            = 0       // 成功
	SubMsgTokenInvalid  = 40001   // access_token 失效
	SubMsgNotAuthorized = 43101   // 用户未授权/拒收该模板
	SubMsgRejected      = 41028   // 用户未授权（form_id 等）
	SubMsgSendLimit     = 9001001 // 下发次数上限
)

// MsgDataField 模板单字段值
type MsgDataField struct {
	Value string `json:"value"`
}

// SendSubscribeMsgReq 订阅消息下发请求
type SendSubscribeMsgReq struct {
	ToUser           string                  `json:"touser"`                      // 接收者 openid
	TemplateId       string                  `json:"template_id"`                 // 模板ID
	Page             string                  `json:"page,omitempty"`              // 点击跳转页
	Data             map[string]MsgDataField `json:"data"`                        // 模板字段（thingN/timeN/...）
	MiniprogramState string                  `json:"miniprogram_state,omitempty"` // formal|trial|develop
	Lang             string                  `json:"lang,omitempty"`              // zh_CN
}

// SendSubscribeMsgResp 下发返回
type SendSubscribeMsgResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}
