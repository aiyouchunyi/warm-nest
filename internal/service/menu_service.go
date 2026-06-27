// Package service @Author larry
// @Date 2026/06/27
// @Desc 服务号自定义菜单发布服务
//
// 自定义菜单是服务号「全局静态配置」：对所有关注者一致，微信侧无法按业务身份（老人/子女）
// 下发不同 pagepath（个性化菜单 menu/addconditional 只支持性别/地区/语言/标签，无业务身份维度）。
// 故菜单统一跳小程序入口页（pages/login/index），由小程序按登录态自行分发到老人/子女页。
// 菜单跳转用小程序 AppId（config.WeChatConf().AppId），非服务号 AppId。

package service

import (
	"fmt"
	"sync"

	"warm-nest/internal/config"
	"warm-nest/internal/feign/wechat"
	wxmodel "warm-nest/internal/feign/wechat/model"
)

// 菜单跳转配置：统一入口页 + 老版本兜底 H5
const (
	menuEntryPage   = "pages/login/index"          // 菜单统一跳小程序入口（小程序内按登录态分发老人/子女）
	menuFallbackURL = "http://www.changkankan.com" // 老版本微信不支持跳小程序时的 H5 兜底
)

// MenuService 服务号菜单服务
type MenuService struct{}

var menuService *MenuService
var menuServiceOnce sync.Once

// GetMenuService 获取菜单服务单例
func GetMenuService() *MenuService {
	menuServiceOnce.Do(func() {
		menuService = &MenuService{}
	})
	return menuService
}

// PublishDefaultMenu 发布 1.0 默认底部菜单（覆盖式）：单按钮「进入小程序」跳入口页。
func (s *MenuService) PublishDefaultMenu() (*wxmodel.CreateMenuResp, error) {
	miniAppId := config.WeChatConf().AppId // 菜单跳的是小程序，用小程序 AppId
	req := wxmodel.CreateMenuReq{
		Button: []wxmodel.MenuButton{
			{Type: "miniprogram", Name: "进入小程序", AppId: miniAppId, PagePath: menuEntryPage, Url: menuFallbackURL},
		},
	}
	resp, err := wechat.Client().CreateMenu(req)
	if err != nil {
		return nil, fmt.Errorf("publish default menu: %w", err)
	}
	if resp.ErrCode != 0 {
		return nil, fmt.Errorf("publish default menu errcode=%d errmsg=%s", resp.ErrCode, resp.ErrMsg)
	}
	return resp, nil
}
