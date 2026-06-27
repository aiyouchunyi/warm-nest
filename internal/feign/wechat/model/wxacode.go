// Package model @Author larry
// @Date 2026/06/15
// @Desc 生成无限制小程序码 wxa/getwxacodeunlimit 相关 DTO
package model

// GetWxaCodeReq 小程序码请求
type GetWxaCodeReq struct {
	Scene      string `json:"scene"`                 // 参数（如 inviteCode），≤32 可见字符
	Page       string `json:"page,omitempty"`        // 跳转页，不填默认主页
	CheckPath  bool   `json:"check_path"`            // 是否校验 page 存在（开发期 false）
	EnvVersion string `json:"env_version,omitempty"` // release|trial|develop
	Width      int    `json:"width,omitempty"`       // 宽度 px
}
