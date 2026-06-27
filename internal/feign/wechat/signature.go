// Package wechat @Author larry
// @Date 2026/06/15
// @Desc access_token 注入（对齐 binance/okex 的 DoSignature 签名 hook 范式）
//
// 微信的"鉴权"不是 hmac 签名，而是给请求带上 access_token query param。复用框架
// client 的 SignRequest 机制：Client() 注册本函数为 WithSignature，凡走 h.SignRequest()
// 的请求在发出前由框架 PreRequestHook 调用本函数注入 token；走 h.Request() 的不注入。

package wechat

import (
	"fmt"
	"net/http"
)

// injectAccessToken 框架 sign hook：取对应主体 access_token 注入到请求 query param。
// 签名类型 sign.DoSignature = func(signAccount string, request *http.Request) error；
// signAccount 指明取哪套 token（SignAccountMini 小程序 / SignAccountOfficial 服务号）。
func injectAccessToken(signAccount string, request *http.Request) error {
	token, err := GetAccessToken(signAccount)
	if err != nil {
		return fmt.Errorf("wechat inject access_token: %w", err)
	}
	q := request.URL.Query()
	q.Set("access_token", token)
	request.URL.RawQuery = q.Encode()
	return nil
}
