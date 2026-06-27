// Package model @Author larry
// @Date 2026/06/25
// @Desc 服务号「获取用户基本信息」cgi-bin/user/info 相关 DTO
package model

// OfficialUserInfoResp 服务号 cgi-bin/user/info 返回（仅取关注态 + unionid）。
// 关注事件 XML 不一定带 unionid（微信常态），但本接口对已关注用户必返 unionid（服务号已绑定开放平台时），
// 故用作回调缺 unionid 时的补偿来源。subscribe=1 关注，=0 未关注（未关注时仅返回 openid）。
type OfficialUserInfoResp struct {
	Subscribe int    `json:"subscribe"`
	OpenId    string `json:"openid"`
	UnionId   string `json:"unionid"`
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
}
