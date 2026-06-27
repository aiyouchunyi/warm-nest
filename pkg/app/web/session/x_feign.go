// Package model @Author larry
// @Date 2025/10/17 10:03
// @Desc

package session

type XFeign struct {
	ClientIp      string `json:"clientIp,omitempty" form:"clientIp"`
	Language      string `json:"language,omitempty" form:"language"`
	UserId        string `json:"userId,omitempty" form:"userId"`
	TraceId       string `json:"traceId,omitempty" form:"traceId"`
	TradePlatform string `json:"tradePlatform,omitempty" form:"tradePlatform"`
	UserGroup     string `json:"userGroup,omitempty" form:"userGroup"`
}
