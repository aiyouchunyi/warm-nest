// Package migrate @Author larry
// @Date 2026/06/21
// @Desc 收货地址簿数据搬迁（问题3：单地址 → 地址簿）
//
// 为什么需要：收货地址从 ElderProfile.Address 单字段升级为独立地址簿表（t_shipping_address）。
// 表本身由 EnableModel 自动建（同步早于本异步迁移，新表已就绪），但存量老人在 ElderProfile.Address
// 里的地址不会自动出现在地址簿。本迁移把每个老人 ElderProfile 里非空的地址搬成地址簿首条并置默认。
//
// 幂等：该老人在地址簿已有任意记录则跳过（避免重跑重复搬入）；地址主体为空的档案不搬。
package migrate

import (
	"fmt"
	"sync"

	"warm-nest/pkg/utils/rands"

	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
)

// ShippingAddressMigrate 收货地址簿数据搬迁
type ShippingAddressMigrate struct {
	elderProfileMapper    *mapper.ElderProfileMapper
	shippingAddressMapper *mapper.ShippingAddressMapper
}

var shippingAddressMigrate *ShippingAddressMigrate
var shippingAddressMigrateOnce sync.Once

// GetShippingAddressMigrate 获取收货地址簿数据搬迁单例
func GetShippingAddressMigrate() *ShippingAddressMigrate {
	shippingAddressMigrateOnce.Do(func() {
		shippingAddressMigrate = &ShippingAddressMigrate{
			elderProfileMapper:    mapper.GetElderProfileMapper(),
			shippingAddressMapper: mapper.GetShippingAddressMapper(),
		}
	})
	return shippingAddressMigrate
}

// Migrate 把存量 ElderProfile.Address 搬成地址簿首条默认地址（幂等）
func (m *ShippingAddressMigrate) Migrate() error {
	profiles, err := m.elderProfileMapper.GetAll()
	if err != nil {
		return fmt.Errorf("shipping address migrate list profiles: %w", err)
	}
	for i := range profiles {
		p := profiles[i]
		// 地址主体为空（省+详细都空）不搬
		if p.Address.MissingShippingField() == "address" {
			continue
		}
		existing, err := m.shippingAddressMapper.ListByElder(p.UserId)
		if err != nil {
			return fmt.Errorf("shipping address migrate check existing %s: %w", p.UserId, err)
		}
		if len(existing) > 0 {
			continue // 幂等：已有地址簿记录，跳过
		}
		addr := &model.ShippingAddress{
			AddressId:   rands.Numeric(),
			ElderUserId: p.UserId,
			Address:     p.Address,
			IsDefault:   true,
			SortNo:      0,
		}
		if err := m.shippingAddressMapper.Create(addr); err != nil {
			return fmt.Errorf("shipping address migrate create %s: %w", p.UserId, err)
		}
	}
	return nil
}
