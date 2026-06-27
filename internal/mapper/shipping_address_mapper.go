// Package mapper @Author larry
// @Date 2026/06/21
// @Desc 收货地址簿映射器（问题3）

package mapper

import (
	"fmt"
	"sync"

	"gorm.io/gorm"

	"warm-nest/pkg/tool/database/mysqls"

	"warm-nest/internal/model"
)

// ShippingAddressMapper 收货地址簿映射器
type ShippingAddressMapper struct {
	mysqls.MysqlMapper[model.ShippingAddress]
}

var shippingAddressMapper *ShippingAddressMapper
var shippingAddressOnce sync.Once

// GetShippingAddressMapper 获取收货地址簿映射器单例
func GetShippingAddressMapper() *ShippingAddressMapper {
	shippingAddressOnce.Do(func() {
		shippingAddressMapper = &ShippingAddressMapper{
			MysqlMapper: mysqls.NewMapper[model.ShippingAddress](),
		}
	})
	return shippingAddressMapper
}

// ListByElder 查某被守护人全部收货地址（默认地址优先，再按 sort_no 升序、创建升序）
func (m *ShippingAddressMapper) ListByElder(elderUserId string) ([]model.ShippingAddress, error) {
	var list []model.ShippingAddress
	err := m.Model().
		Where("elder_user_id = ?", elderUserId).
		Order("is_default DESC, sort_no ASC, created_at ASC").
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list shipping address by elder %s: %w", elderUserId, err)
	}
	return list, nil
}

// GetByAddressId 按业务地址ID查（找不到返回 nil,nil）
func (m *ShippingAddressMapper) GetByAddressId(addressId string) (*model.ShippingAddress, error) {
	var a model.ShippingAddress
	err := m.Model().Where("address_id = ?", addressId).First(&a).Error
	if err != nil {
		if mysqls.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get shipping address by id %s: %w", addressId, err)
	}
	return &a, nil
}

// GetDefaultByElder 查某被守护人的默认地址（无默认返回 nil,nil）
func (m *ShippingAddressMapper) GetDefaultByElder(elderUserId string) (*model.ShippingAddress, error) {
	var a model.ShippingAddress
	err := m.Model().
		Where("elder_user_id = ? AND is_default = ?", elderUserId, true).
		First(&a).Error
	if err != nil {
		if mysqls.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get default shipping address by elder %s: %w", elderUserId, err)
	}
	return &a, nil
}

// ClearDefaultByTx 事务内把某被守护人全部地址的 is_default 置 false（设默认前调用，保证唯一默认）
func (m *ShippingAddressMapper) ClearDefaultByTx(tx *gorm.DB, elderUserId string) error {
	err := tx.Model(&model.ShippingAddress{}).
		Where("elder_user_id = ? AND is_default = ?", elderUserId, true).
		Update("is_default", false).Error
	if err != nil {
		return fmt.Errorf("clear default shipping address %s: %w", elderUserId, err)
	}
	return nil
}
