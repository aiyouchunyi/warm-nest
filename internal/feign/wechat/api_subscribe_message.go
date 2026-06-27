// Package wechat @Author larry
// @Date 2026/06/15
// @Desc 订阅消息下发 cgi-bin/message/subscribe/send
//
// 一次性订阅：用户授权一次后端发一次。授权次数由微信侧管理、后端无状态直发，
// 不建授权记录表。常见 errcode：0 成功；43101/41028 未授权（让前端重新引导授权）；
// 40001 token 失效（刷新重试）；9001001 下发上限。

package wechat

import (
	"fmt"

	"warm-nest/pkg/tool/client/utils"

	"warm-nest/internal/config"
	"warm-nest/internal/feign/wechat/model"
)

// SendSubscribeMessage 下发订阅消息，返回 errcode 供上层按未授权/失效分别处理（需 access_token，走 SignRequest 自动注入）
func (h *Hook) SendSubscribeMessage(req model.SendSubscribeMsgReq) (*model.SendSubscribeMsgResp, error) {
	if config.WeChatConf().Mock {
		// 本地联调旁路：跳过真实下发，返回成功
		return &model.SendSubscribeMsgResp{ErrCode: model.SubMsgOk, ErrMsg: "mock ok"}, nil
	}
	resp, err := utils.ResultPtrV2[model.SendSubscribeMsgResp](
		h.SignRequest(SignAccountMini).
			// 微信接口可能返回 text/plain，强制按 json 解析
			ForceContentType("application/json; charset=utf-8").
			SetBody(req).
			SetResult(model.SendSubscribeMsgResp{}).
			Post("/cgi-bin/message/subscribe/send"),
	)
	if err != nil {
		return nil, fmt.Errorf("subscribe message send: %w", err)
	}
	return resp, nil
}
