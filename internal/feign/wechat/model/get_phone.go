// Package model @Author larry
// @Date 2026/06/21
// @Desc 小程序手机号快速验证 getuserphonenumber 相关 DTO
package model

// GetPhoneNumberReq 换手机号请求体
type GetPhoneNumberReq struct {
	Code string `json:"code"` // getPhoneNumber 回调的动态令牌
}

// PhoneInfoData 微信返回的 phone_info 子结构
type PhoneInfoData struct {
	PhoneNumber     string `json:"phoneNumber"`     // 带区号完整号码
	PurePhoneNumber string `json:"purePhoneNumber"` // 无区号手机号（国内场景用这个）
	CountryCode     string `json:"countryCode"`     // 区号
}

// GetPhoneNumberResp 换手机号返回
type GetPhoneNumberResp struct {
	ErrCode   int           `json:"errcode"`
	ErrMsg    string        `json:"errmsg"`
	PhoneInfo PhoneInfoData `json:"phone_info"`
}
