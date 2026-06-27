// Package wechat @Author larry
// @Date 2026/06/15
// @Desc 小程序登录凭证校验 code2session

package wechat

import (
	"fmt"

	"warm-nest/pkg/tool/client/utils"

	"warm-nest/internal/config"
	"warm-nest/internal/feign/wechat/model"
)

// Code2Session 用小程序登录 code 换取 openid/unionid/session_key（免 access_token，用 appid/secret）
func (h *Hook) Code2Session(code string) (*model.Code2SessionResp, error) {
	conf := config.WeChatConf()
	if conf.Mock {
		// 本地联调旁路：用 code 拼一个稳定的假 openid，便于复现同一用户
		return &model.Code2SessionResp{OpenId: "mock_openid_" + code, SessionKey: "mock_session_key"}, nil
	}
	resp, err := utils.ResultPtrV2[model.Code2SessionResp](
		h.Request().
			// 微信 jscode2session 返回 Content-Type=text/plain（非 application/json），
			// resty 默认只对 json/xml 自动反序列化 → 不强制则 resp 各字段全零值（openid 空被误判失败）。
			// 外部 API 返回 text/plain 时强制按 json 解析。
			ForceContentType("application/json; charset=utf-8").
			SetQueryParam("appid", conf.AppId).
			SetQueryParam("secret", conf.Secret).
			SetQueryParam("js_code", code).
			SetQueryParam("grant_type", "authorization_code").
			SetResult(model.Code2SessionResp{}).
			Get("/sns/jscode2session"),
	)
	if err != nil {
		return nil, fmt.Errorf("wechat code2session: %w", err)
	}
	if resp.ErrCode != 0 || resp.OpenId == "" {
		return nil, fmt.Errorf("wechat code2session errcode=%d errmsg=%s", resp.ErrCode, resp.ErrMsg)
	}
	return resp, nil
}
