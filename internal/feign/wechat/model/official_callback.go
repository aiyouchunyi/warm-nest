// Package model @Author larry
// @Date 2026/06/18
// @Desc 服务号（公众平台）消息回调事件 DTO 与事件类型常量（PRD §8.2 关注/取关）
package model

import "encoding/xml"

// 服务号事件类型
const (
	OfficialEventSubscribe   = "subscribe"   // 关注
	OfficialEventUnsubscribe = "unsubscribe" // 取关
)

// OfficialEventMsg 服务号推送的事件 XML（仅取关注/取关所需字段）
//
// UnionId：仅当该服务号已绑定微信开放平台时，事件 XML 才带 <Unionid>。
// 它是跨小程序/服务号同一用户的锚点，本服务关注判定按 UnionId 关联 User（见 model.Fan）。
// 未绑开放平台 → 该字段为空，V1 无法把服务号 openid 映射到小程序用户，落库时记日志降级。
type OfficialEventMsg struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`   // 服务号原始ID
	FromUserName string   `xml:"FromUserName"` // 用户在该服务号下的 openid（非小程序 openid）
	MsgType      string   `xml:"MsgType"`      // event
	Event        string   `xml:"Event"`        // subscribe/unsubscribe
	UnionId      string   `xml:"Unionid"`      // 开放平台 unionid（绑定开放平台后才带；跨主体锚点）
}
