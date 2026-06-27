// Package api @Author larry
// @Date 2026/06/21
// @Desc 查收货地址簿列表（公用：子女传 elderUserId 看老人的，老人不传看自己的）

package api

import (
	"warm-nest/pkg/app/web/session"

	"warm-nest/internal/service"
)

// ListAddressReq 地址列表请求（需登录；可选 elderUserId）
type ListAddressReq struct {
	session.Session
	ElderUserId string `form:"elderUserId"` // 可选：看某老人的（守护校验），不传=自己
}

// List 查收货地址簿（默认地址优先排序）
func (a *ApiAddress) List(req ListAddressReq) (interface{}, error) {
	return service.GetAddressService().List(req.ReqUser, req.ElderUserId)
}
