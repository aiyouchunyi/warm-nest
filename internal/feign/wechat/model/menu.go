// Package model @Author larry
// @Date 2026/06/27
// @Desc 服务号自定义菜单 cgi-bin/menu/create 相关 DTO
//
// 自定义菜单是服务号「全局静态配置」：对所有关注者一致，无法按业务身份（老人/子女）下发不同内容。
// 故菜单统一跳小程序入口页（pages/login/index），由小程序按登录态自行分发到老人/子女页。
// 菜单接口不卡小程序发布状态（与模板消息 40165 不同）：未发布时普通用户点提示「开发中」，
// 体验/开发成员点可直接打开体验版。
package model

// MenuButton 自定义菜单按钮（一级或带 sub_button 的父级）
type MenuButton struct {
	Type     string `json:"type,omitempty"`     // miniprogram / click / view 等；父级菜单不填
	Name     string `json:"name"`               // 按钮文案
	AppId    string `json:"appid,omitempty"`    // type=miniprogram：跳转小程序 appid（须与服务号同主体绑定）
	PagePath string `json:"pagepath,omitempty"` // type=miniprogram：小程序页（无前置斜杠、无后缀）
	Url      string `json:"url,omitempty"`      // type=miniprogram 的老版本兜底 H5 / type=view 的网页
}

// CreateMenuReq 创建自定义菜单请求（button 至多 3 个一级）
type CreateMenuReq struct {
	Button []MenuButton `json:"button"`
}

// CreateMenuResp 创建菜单返回
type CreateMenuResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}
