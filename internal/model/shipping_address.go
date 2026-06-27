// Package model @Author larry
// @Date 2026/06/21
// @Desc 收货地址簿模型（问题3：多地址 + 默认地址）
//
// 设计要点：
//   - 从 ElderProfile.Address 单地址升级为地址簿：一个被守护人可存多个收货地址，其一为默认。
//     归属被守护人（ElderUserId）而非通用 User —— 奖励（鸡蛋）只发给打卡老人、纯子女账号不产生收货，
//     地址属于「被守护人」语义（与 ElderProfile/RewardClaim 的 ElderUserId 一致）。
//   - 双端同源：老人本人或其 active 守护人均可增删改/设默认（鉴权复用 FamilyService.EnsureCanAccess）。
//   - 复用 model.Address 值对象作 json 列：领取下单时整条快照进 RewardClaim.ReceiverAddress（同为
//     model.Address，零映射直拷），改址不影响历史单。
//   - 默认唯一性不靠 DB 约束（一个老人多条、仅一条 IsDefault=true 是业务不变量），靠 service 在事务内
//     先 ClearDefault 再置默认保证；排序 IsDefault 优先、再按 SortNo。
package model

import (
	"warm-nest/pkg/tool/database"
)

// ShippingAddress 收货地址簿（一个被守护人多条，其一为默认）
type ShippingAddress struct {
	database.Identify
	AddressId   string `gorm:"size:32;uniqueIndex:u_address;comment:业务地址ID（rands.Numeric 生成，对外不可枚举）" json:"addressId"`
	ElderUserId string `gorm:"size:32;index:i_elder;comment:被守护人用户ID（地址归属人，User.UserId）" json:"elderUserId"`

	// Address 结构化地址（省市区街道详细 + 收货人 + 收货电话），与 RewardClaim.ReceiverAddress 同构，下单时整条快照。
	Address Address `gorm:"type:json;comment:结构化地址（省市区街道详细+收货人+收货电话）" json:"address"`

	IsDefault bool `gorm:"comment:是否默认地址（业务不变量：同一老人仅一条为 true，由 service 事务保证）" json:"isDefault"`
	SortNo    int  `gorm:"comment:排序号（默认地址优先，再按此升序；越小越靠前）" json:"sortNo"`

	database.SystemField
}
