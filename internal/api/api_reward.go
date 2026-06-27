// Package api @Author larry
// @Date 2026/06/15
// @Desc 奖励领取接口聚合 + 列表/领取

package api

import (
	"sync"

	"warm-nest/pkg/app/web/session"
	"warm-nest/pkg/kinds/jsons"
	"warm-nest/pkg/tool/machines/model/constant"
	"warm-nest/pkg/tool/machines/model/dto"

	"warm-nest/internal/reward/machine"
	"warm-nest/internal/service"
)

// ApiReward 奖励接口
type ApiReward struct{}

var apiReward *ApiReward
var apiRewardOnce sync.Once

// GetApiReward 获取奖励接口单例
func GetApiReward() *ApiReward {
	apiRewardOnce.Do(func() {
		apiReward = &ApiReward{}
	})
	return apiReward
}

// ListRewardReq 领取记录列表请求（公用：子女传 elderUserId 看老人的，老人不传看自己的）
type ListRewardReq struct {
	session.Session
	ElderUserId string `form:"elderUserId"` // 可选：看某老人的（守护校验），不传=自己
}

// List 查用户领取记录（问题6 公用）
func (a *ApiReward) List(req ListRewardReq) (interface{}, error) {
	elderUserId, err := service.GetFamilyService().ResolveElder(req.ReqUser, req.ElderUserId)
	if err != nil {
		return nil, err
	}
	return service.GetRewardService().ListByUser(elderUserId)
}

// ClaimDetailReq 领取详情请求（公用·模式B：只传 claimId，归属由记录定）
type ClaimDetailReq struct {
	session.Session
	ClaimId string `form:"claimId" validate:"required"`
}

// ClaimDetail 按 claimId 查领取详情（问题5，老人/子女公用，守护校验防越权）
func (a *ApiReward) ClaimDetail(req ClaimDetailReq) (interface{}, error) {
	return service.GetRewardService().GetClaimDetail(req.ReqUser, req.ClaimId)
}

// AddressPreheatReq 奖励地址预热查询请求（公用：子女传 elderUserId 查老人的，不传查自己）
type AddressPreheatReq struct {
	session.Session
	ElderUserId string `form:"elderUserId"` // 可选：查某老人的（守护校验），不传=自己
}

// AddressPreheat 查是否需要「奖励地址预热提醒」（PRD §6.6.3）：
// 返回 needAddress——该老人地址簿为空则 true，前端据此显红点/引导提前填收货地址。
// 与具体待领奖励无关，不阻断领取（待领不过期）。
func (a *ApiReward) AddressPreheat(req AddressPreheatReq) (interface{}, error) {
	need, err := service.GetAddressService().NeedAddressPreheat(req.ReqUser, req.ElderUserId)
	if err != nil {
		return nil, err
	}
	return map[string]bool{"needAddress": need}, nil
}

// ClaimReq 领取请求（需登录）
type ClaimReq struct {
	session.Session
	ClaimId   string `json:"claimId" validate:"required"`
	AddressId string `json:"addressId"` // 可选：指定收货地址（地址簿），不传用该老人默认地址（问题3）
}

// Claim 用户领取奖励（PENDING→CLAIMED，走 machine approve 动作）。
// 归属校验（模式B，防越权领他人奖励）：claimId 反查归属老人，校验当前登录人是本人或其守护人，
// 通过才推进状态机——老人 id 由记录定、不接受前端传，子女领取亦无需传老人 id。
// addressId 经 Params 透传进领取 transition（默认领默认地址，传了则覆盖，问题3）。
func (a *ApiReward) Claim(req ClaimReq) (interface{}, error) {
	if err := service.GetRewardService().EnsureClaimAccess(req.ReqUser, req.ClaimId); err != nil {
		return nil, err
	}
	mreq := machine.RewardClaimReq{MachineTaskReq: dto.MachineTaskReq{
		Session: req.Session,
		TaskId:  req.ClaimId,
		Action:  constant.EventActionApprove,
		Params:  jsons.JSONObject{"addressId": req.AddressId},
	}}
	if err := machine.GetRewardClaimService().Manual(mreq); err != nil {
		return nil, err
	}
	return map[string]bool{"ok": true}, nil
}
