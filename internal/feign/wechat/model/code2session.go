// Package model @Author larry
// @Date 2026/06/15
// @Desc 小程序登录凭证校验 code2session 相关 DTO
package model

// Code2SessionResp 微信 jscode2session 返回
type Code2SessionResp struct {
	OpenId     string `json:"openid"`
	UnionId    string `json:"unionid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}
