// Package wechat @Author larry
// @Date 2026/06/25
// @Desc 服务号「获取用户基本信息」cgi-bin/user/info（用于回调缺 unionid 时补偿）
//
// 关注/取关事件推送的 XML 不一定带 <Unionid>（微信常态，与是否绑定开放平台无强关系），
// 但本接口对已关注用户必返 unionid（服务号已绑定开放平台同主体时）。故回调拿到 openid 后
// 调本接口补 unionid，再落 Fan 表。需服务号 access_token（SignRequest(SignAccountOfficial) 自动注入 query）。
package wechat

import (
	"fmt"

	"warm-nest/pkg/tool/client/utils"

	"warm-nest/internal/config"
	"warm-nest/internal/feign/wechat/model"
)

// GetOfficialUserInfo 按服务号 openid 查用户基本信息（关注态 + unionid）。
func (h *Hook) GetOfficialUserInfo(officialOpenId string) (*model.OfficialUserInfoResp, error) {
	if config.WeChatConf().Mock {
		// 本地联调旁路：用 openid 拼一个稳定假 unionid，便于复现同一用户
		return &model.OfficialUserInfoResp{Subscribe: 1, OpenId: officialOpenId, UnionId: "mock_union_" + officialOpenId}, nil
	}
	resp, err := utils.ResultPtrV2[model.OfficialUserInfoResp](
		h.SignRequest(SignAccountOfficial).
			// 微信接口可能返回 text/plain，强制按 json 解析（对齐 api_code2session / api_template_message 范式）
			ForceContentType("application/json; charset=utf-8").
			SetQueryParam("openid", officialOpenId).
			SetQueryParam("lang", "zh_CN").
			SetResult(model.OfficialUserInfoResp{}).
			Get("/cgi-bin/user/info"),
	)
	if err != nil {
		return nil, fmt.Errorf("wechat get official user info: %w", err)
	}
	if resp.ErrCode != 0 {
		return nil, fmt.Errorf("wechat get official user info errcode=%d errmsg=%s", resp.ErrCode, resp.ErrMsg)
	}
	return resp, nil
}
