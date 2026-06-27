// Package wechat @Author larry
// @Date 2026/06/21
// @Desc 小程序手机号快速验证 getuserphonenumber（问题5：子女授权手机号）
//
// 前端 <button open-type="getPhoneNumber"> 授权后拿到 code（5min 有效、一次性），
// 后端用 code 换真实手机号（需 access_token，走 SignRequest 自动注入）。
// 不存 code、不建授权记录——换到手机号即写入对应 User.Phone。
package wechat

import (
	"fmt"

	"warm-nest/pkg/tool/client/utils"

	"warm-nest/internal/config"
	"warm-nest/internal/feign/wechat/model"
)

// GetPhoneNumber 用 getPhoneNumber 的 code 换真实手机号（需 access_token，走 SignRequest 自动注入）。
// 返回纯手机号（无区号）；errcode != 0 或空号视为失败。⚠️ 手机号敏感，调用方落库时勿打日志。
func (h *Hook) GetPhoneNumber(code string) (string, error) {
	if config.WeChatConf().Mock {
		// 本地联调旁路：用 code 拼一个稳定假号，便于复现
		return "13800000000", nil
	}
	resp, err := utils.ResultPtrV2[model.GetPhoneNumberResp](
		h.SignRequest(SignAccountMini).
			// 微信接口可能返回 text/plain，强制按 json 解析
			ForceContentType("application/json; charset=utf-8").
			SetBody(model.GetPhoneNumberReq{Code: code}).
			SetResult(model.GetPhoneNumberResp{}).
			Post("/wxa/business/getuserphonenumber"),
	)
	if err != nil {
		return "", fmt.Errorf("wechat get phone number: %w", err)
	}
	if resp.ErrCode != 0 || resp.PhoneInfo.PurePhoneNumber == "" {
		return "", fmt.Errorf("wechat get phone number errcode=%d errmsg=%s", resp.ErrCode, resp.ErrMsg)
	}
	return resp.PhoneInfo.PurePhoneNumber, nil
}
