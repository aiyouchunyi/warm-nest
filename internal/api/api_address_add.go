// Package api @Author larry
// @Date 2026/06/21
// @Desc 新增收货地址（公用：子女为老人加 / 老人为自己加）

package api

import (
	"warm-nest/pkg/app/web/session"

	"warm-nest/internal/model"
	"warm-nest/internal/service"
)

// AddAddressReq 新增地址请求（需登录）
type AddAddressReq struct {
	session.Session
	ElderUserId string        `json:"elderUserId" validate:"required"` // 地址归属老人
	Address     model.Address `json:"address" validate:"required"`     // 结构化地址（省市区街道详细+收货人+电话）
	SetDefault  bool          `json:"setDefault"`                      // 是否设为默认；首条地址强制默认
}

// Add 新增收货地址，返回新地址
func (a *ApiAddress) Add(req AddAddressReq) (interface{}, error) {
	return service.GetAddressService().Add(req.ReqUser, req.ElderUserId, req.Address, req.SetDefault)
}
