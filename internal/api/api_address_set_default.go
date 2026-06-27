// Package api @Author larry
// @Date 2026/06/21
// @Desc 设置默认收货地址

package api

import (
	"warm-nest/pkg/app/web/session"

	"warm-nest/internal/service"
)

// SetDefaultAddressReq 设默认地址请求（需登录；模式B：归属由 addressId 反查校验）
type SetDefaultAddressReq struct {
	session.Session
	AddressId string `json:"addressId" validate:"required"`
}

// SetDefault 设置默认收货地址（同一老人仅一条为默认）
func (a *ApiAddress) SetDefault(req SetDefaultAddressReq) (interface{}, error) {
	if err := service.GetAddressService().SetDefault(req.ReqUser, req.AddressId); err != nil {
		return nil, err
	}
	return map[string]bool{"ok": true}, nil
}
