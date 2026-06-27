// Package wechat @Author larry
// @Date 2026/06/15
// @Desc 生成无限制小程序码 wxa/getwxacodeunlimit
//
// 成功返回二进制图片 bytes；失败返回 JSON {errcode,errmsg}。按 Content-Type 区分。

package wechat

import (
	"encoding/json"
	"fmt"
	"strings"

	"warm-nest/internal/config"
	"warm-nest/internal/feign/wechat/model"
)

// GetUnlimitedWxaCode 生成小程序码，返回图片二进制（需 access_token，走 SignRequest 自动注入）
func (h *Hook) GetUnlimitedWxaCode(req model.GetWxaCodeReq) ([]byte, error) {
	if config.WeChatConf().Mock {
		// 本地联调旁路：返回最小合法 PNG 占位（1x1 透明），不真调微信
		return mockWxaCodePNG(), nil
	}
	resp, err := h.SignRequest(SignAccountMini).
		SetBody(req).
		Post("/wxa/getwxacodeunlimit")
	if err != nil {
		return nil, fmt.Errorf("wxacode send: %w", err)
	}
	// 微信失败时返回 JSON 错误体（Content-Type application/json），成功返回 image/*
	if strings.Contains(resp.Header().Get("Content-Type"), "application/json") {
		var e struct {
			ErrCode int    `json:"errcode"`
			ErrMsg  string `json:"errmsg"`
		}
		_ = json.Unmarshal(resp.Body(), &e)
		return nil, fmt.Errorf("wxacode errcode=%d errmsg=%s", e.ErrCode, e.ErrMsg)
	}
	return resp.Body(), nil
}

// mockWxaCodePNG 本地联调用的最小合法 PNG（1x1）字节
func mockWxaCodePNG() []byte {
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00,
		0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}
}
