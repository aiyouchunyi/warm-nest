// Package wechat @Author larry
// @Date 2026/06/27
// @Desc 服务号自定义菜单创建 cgi-bin/menu/create（需服务号 access_token）
//
// 菜单是服务号全局配置，发布后覆盖旧菜单、对所有关注者生效。需服务号 access_token
// （SignRequest(SignAccountOfficial) 自动注入 query）。errcode=0 即发布成功。
package wechat

import (
	"fmt"

	"warm-nest/pkg/tool/client/utils"

	"warm-nest/internal/config"
	"warm-nest/internal/feign/wechat/model"
)

// CreateMenu 发布服务号自定义菜单（覆盖式：整份替换旧菜单）。
func (h *Hook) CreateMenu(req model.CreateMenuReq) (*model.CreateMenuResp, error) {
	if config.WeChatConf().Mock {
		// 本地联调旁路：跳过真实下发，返回成功
		return &model.CreateMenuResp{ErrCode: 0, ErrMsg: "mock ok"}, nil
	}
	resp, err := utils.ResultPtrV2[model.CreateMenuResp](
		h.SignRequest(SignAccountOfficial).
			// 微信接口可能返回 text/plain，强制按 json 解析（对齐 api_template_message 范式）
			ForceContentType("application/json; charset=utf-8").
			SetBody(req).
			SetResult(model.CreateMenuResp{}).
			Post("/cgi-bin/menu/create"),
	)
	if err != nil {
		return nil, fmt.Errorf("wechat create menu: %w", err)
	}
	return resp, nil
}
