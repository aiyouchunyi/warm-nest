// Package api @Author larry
// @Date 2026/06/21
// @Desc 修改收货地址内容（不动默认标志，改默认走 set-default）

package api

import (
	"warm-nest/pkg/app/web/session"

	"warm-nest/internal/model"
	"warm-nest/internal/service"
)

// UpdateAddressReq 修改地址请求（需登录；模式B：归属由 addressId 反查校验）
type UpdateAddressReq struct {
	session.Session
	AddressId string        `json:"addressId" validate:"required"`
	Address   model.Address `json:"address" validate:"required"` // 全量覆盖该条地址内容
}

// Update 修改收货地址内容
func (a *ApiAddress) Update(req UpdateAddressReq) (interface{}, error) {
	if err := service.GetAddressService().Update(req.ReqUser, req.AddressId, req.Address); err != nil {
		return nil, err
	}
	return map[string]bool{"ok": true}, nil
}
