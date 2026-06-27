// Package model @Author larry
// @Date 2026/06/22
// @Desc 服务号模板消息下发 cgi-bin/message/template/send 相关 DTO 与响应码
package model

// 服务号模板消息常见错误码
const (
	TmplMsgOk            = 0     // 成功
	TmplMsgTokenInvalid  = 40001 // access_token 失效
	TmplMsgInvalidToUser = 40003 // touser openid 非法（含未关注）
	TmplMsgNeedSubscribe = 43004 // 需接收者先关注服务号
)

// TmplDataField 模板单字段值（服务号模板支持 color，订阅消息不支持）
type TmplDataField struct {
	Value string `json:"value"`
	Color string `json:"color,omitempty"`
}

// TmplMiniprogram 模板消息跳转小程序（可选）
type TmplMiniprogram struct {
	AppId    string `json:"appid"`              // 跳转的小程序 appid（须与服务号同主体绑定）
	PagePath string `json:"pagepath,omitempty"` // 跳转页
}

// SendTemplateMsgReq 服务号模板消息下发请求
type SendTemplateMsgReq struct {
	ToUser      string                   `json:"touser"`                // 接收者服务号 openid
	TemplateId  string                   `json:"template_id"`           // 服务号模板ID
	Url         string                   `json:"url,omitempty"`         // 点击跳转 H5（与 miniprogram 二选一）
	Miniprogram *TmplMiniprogram         `json:"miniprogram,omitempty"` // 点击跳转小程序（与 url 二选一）
	Data        map[string]TmplDataField `json:"data"`                  // 模板字段
}

// SendTemplateMsgResp 下发返回
type SendTemplateMsgResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	MsgId   int64  `json:"msgid"`
}
