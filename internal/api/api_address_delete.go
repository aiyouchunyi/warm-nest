// Package api @Author larry
// @Date 2026/06/21
// @Desc 删除收货地址

package api

import (
	"warm-nest/pkg/app/web/session"

	"warm-nest/internal/service"
)

// DeleteAddressReq 删除地址请求（需登录；模式B：归属由 addressId 反查校验）
type DeleteAddressReq struct {
	session.Session
	AddressId string `json:"addressId" validate:"required"`
}

// Delete 删除收货地址（删默认后自动把最早一条补为默认）
func (a *ApiAddress) Delete(req DeleteAddressReq) (interface{}, error) {
	if err := service.GetAddressService().Delete(req.ReqUser, req.AddressId); err != nil {
		return nil, err
	}
	return map[string]bool{"ok": true}, nil
}
