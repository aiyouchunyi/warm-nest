// Package wechat @Author larry
// @Date 2026/06/22
// @Desc 服务号模板消息下发 cgi-bin/message/template/send（PRD §5.2 主推渠道）
//
// 与小程序订阅消息（subscribe/send）是两套独立体系：
//   - 主体：服务号 access_token（SignAccountOfficial），touser=服务号 openid（Fan.OfficialOpenId）；
//   - 投递前提：用户已关注服务号（关注态由 Fan 表维护），未关注微信侧直接拒发；
//   - data 结构：{key:{value,color}}，可带跳转 url 或 miniprogram；无 subscribe 的 miniprogram_state。
//
// 常见 errcode：0 成功；40003 touser 非法/未关注；43004 需先关注；40001 token 失效。
package wechat

import (
	"fmt"

	"warm-nest/pkg/tool/client/utils"

	"warm-nest/internal/config"
	"warm-nest/internal/feign/wechat/model"
)

// SendTemplateMessage 下发服务号模板消息，返回 errcode 供上层按未关注/失效分别处理
// （需服务号 access_token，走 SignRequest(SignAccountOfficial) 自动注入）
func (h *Hook) SendTemplateMessage(req model.SendTemplateMsgReq) (*model.SendTemplateMsgResp, error) {
	if config.WeChatConf().Mock {
		// 本地联调旁路：跳过真实下发，返回成功
		return &model.SendTemplateMsgResp{ErrCode: model.TmplMsgOk, ErrMsg: "mock ok"}, nil
	}
	resp, err := utils.ResultPtrV2[model.SendTemplateMsgResp](
		h.SignRequest(SignAccountOfficial).
			// 微信接口可能返回 text/plain，强制按 json 解析
			ForceContentType("application/json; charset=utf-8").
			SetBody(req).
			SetResult(model.SendTemplateMsgResp{}).
			Post("/cgi-bin/message/template/send"),
	)
	if err != nil {
		return nil, fmt.Errorf("template message send: %w", err)
	}
	return resp, nil
}
