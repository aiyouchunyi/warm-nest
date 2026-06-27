// Package model @Author larry
// @Date 2026/06/15
// @Desc access_token 拉取 cgi-bin/token 相关 DTO
package model

// AccessTokenResp 微信 cgi-bin/token 返回
type AccessTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}
