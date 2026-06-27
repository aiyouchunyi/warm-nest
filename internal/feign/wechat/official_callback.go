// Package wechat @Author larry
// @Date 2026/06/18
// @Desc 服务号（公众平台）消息回调验签与事件解析（PRD §8.2 关注/取关）
//
// 与本包其它文件不同：这里不是"我们调微信"，而是"微信服务器回调我们"。
// 微信公众平台把用户关注/取关事件以 XML POST 到后台配置的回调 URL，并在
// 配置时先发一个 GET（带 echostr）做 URL 有效性校验。两步都按下述 sha1 验签：
//   将 token、timestamp、nonce 三者字典序排序后拼接 → sha1 → 与 signature 比对。
// token 由公众平台后台与我方约定（config.OfficialCallbackToken），非 access_token。

package wechat

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"sort"
	"strings"

	"warm-nest/internal/feign/wechat/model"
)

// VerifyOfficialSignature 校验服务号回调签名（GET URL 校验与 POST 事件共用）。
// token 取自 config.OfficialCallbackToken。
func VerifyOfficialSignature(token, signature, timestamp, nonce string) bool {
	if token == "" || signature == "" {
		return false
	}
	arr := []string{token, timestamp, nonce}
	sort.Strings(arr)
	h := sha1.New()
	h.Write([]byte(strings.Join(arr, "")))
	return hex.EncodeToString(h.Sum(nil)) == signature
}

// ParseOfficialEvent 解析服务号事件 XML body
func ParseOfficialEvent(body []byte) (*model.OfficialEventMsg, error) {
	var msg model.OfficialEventMsg
	if err := xml.Unmarshal(body, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
