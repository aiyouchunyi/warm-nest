// Package service @Author larry
// @Date 2026/06/21
// @Desc 收货地址簿服务（问题3：多地址 + 默认地址，双端同源）
//
// 鉴权全部收口到 FamilyService（本人或 active 守护人）：列表走 ResolveElder，
// 增删改/设默认按 addressId 反查归属老人后 EnsureCanAccess（模式B 防越权）。
// 默认唯一性靠事务内「先 ClearDefault 再置默认」保证，不依赖 DB 约束。

package service

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/kinds/jsons"
	"warm-nest/pkg/tool/database/mysqls"
	"warm-nest/pkg/utils/rands"

	wnconst "warm-nest/internal/constant"
	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
)

// AddressService 收货地址簿服务
type AddressService struct {
	shippingAddressMapper *mapper.ShippingAddressMapper
	guardianshipMapper    *mapper.GuardianshipMapper
}

var addressService *AddressService
var addressServiceOnce sync.Once

// GetAddressService 获取收货地址簿服务单例
func GetAddressService() *AddressService {
	addressServiceOnce.Do(func() {
		addressService = &AddressService{
			shippingAddressMapper: mapper.GetShippingAddressMapper(),
			guardianshipMapper:    mapper.GetGuardianshipMapper(),
		}
	})
	return addressService
}

// List 查某被守护人的地址簿（默认优先）。鉴权走 ResolveElder：
// 守护人传 elderUserId 看老人的、老人不传看自己的，非授权返 ErrNotGuardian。
func (s *AddressService) List(viewerUserId, reqElderUserId string) ([]model.ShippingAddress, error) {
	elderUserId, err := GetFamilyService().ResolveElder(viewerUserId, reqElderUserId)
	if err != nil {
		return nil, err
	}
	return s.shippingAddressMapper.ListByElder(elderUserId)
}

// Add 为某被守护人新增收货地址。viewer 须为该老人本人或 active 守护人（EnsureCanAccess）。
// 首条地址强制为默认；setDefault=true 时事务内先清旧默认再置本条默认。
func (s *AddressService) Add(viewerUserId, elderUserId string, addr model.Address, setDefault bool) (*model.ShippingAddress, error) {
	if err := GetFamilyService().EnsureCanAccess(viewerUserId, elderUserId); err != nil {
		return nil, err
	}
	existing, err := s.shippingAddressMapper.ListByElder(elderUserId)
	if err != nil {
		return nil, err
	}
	makeDefault := setDefault || len(existing) == 0 // 首条必为默认

	addressModel := &model.ShippingAddress{
		AddressId:   rands.Numeric(),
		ElderUserId: elderUserId,
		Address:     addr,
		IsDefault:   makeDefault,
		SortNo:      len(existing), // 追加到末尾
	}
	err = mysqls.DB().Transaction(func(tx *gorm.DB) error {
		if makeDefault {
			if err := s.shippingAddressMapper.ClearDefaultByTx(tx, elderUserId); err != nil {
				return err
			}
		}
		if err := tx.Create(addressModel).Error; err != nil {
			return fmt.Errorf("create shipping address: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return addressModel, nil
}

// Update 修改某条地址内容（不动默认标志）。按 addressId 反查归属老人 + EnsureCanAccess 防越权。
func (s *AddressService) Update(viewerUserId, addressId string, addr model.Address) error {
	existing, err := s.ensureAccessByAddressId(viewerUserId, addressId)
	if err != nil {
		return err
	}
	existing.Address = addr
	return s.shippingAddressMapper.Update(existing)
}

// Delete 删除某条地址。删默认地址后若仍有其他地址，把最早一条置默认（避免老人无默认地址）。
func (s *AddressService) Delete(viewerUserId, addressId string) error {
	existing, err := s.ensureAccessByAddressId(viewerUserId, addressId)
	if err != nil {
		return err
	}
	return mysqls.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(existing).Error; err != nil {
			return fmt.Errorf("delete shipping address %s: %w", addressId, err)
		}
		if !existing.IsDefault {
			return nil
		}
		// 删的是默认地址：在剩余地址里挑最早一条补默认（ListByElder 已按默认/排序/创建升序）
		rest, err := s.listByElderInTx(tx, existing.ElderUserId)
		if err != nil {
			return err
		}
		if len(rest) == 0 {
			return nil
		}
		rest[0].IsDefault = true
		if err := tx.Model(&model.ShippingAddress{}).
			Where("address_id = ?", rest[0].AddressId).
			Update("is_default", true).Error; err != nil {
			return fmt.Errorf("promote default shipping address %s: %w", rest[0].AddressId, err)
		}
		return nil
	})
}

// SetDefault 把某条地址设为默认（事务内先清旧默认）。
func (s *AddressService) SetDefault(viewerUserId, addressId string) error {
	existing, err := s.ensureAccessByAddressId(viewerUserId, addressId)
	if err != nil {
		return err
	}
	if existing.IsDefault {
		return nil // 已是默认，幂等
	}
	return mysqls.DB().Transaction(func(tx *gorm.DB) error {
		if err := s.shippingAddressMapper.ClearDefaultByTx(tx, existing.ElderUserId); err != nil {
			return err
		}
		if err := tx.Model(&model.ShippingAddress{}).
			Where("address_id = ?", addressId).
			Update("is_default", true).Error; err != nil {
			return fmt.Errorf("set default shipping address %s: %w", addressId, err)
		}
		return nil
	})
}

// NeedAddressPreheat 判某被守护人是否需要「奖励地址预热提醒」（PRD §6.6.3）：
// 判据 = 该老人地址簿为空（一条都没有）。鉴权走 ResolveElder（守护人传 elderUserId
// 看老人的、老人不传看自己的）。前端据此显红点/引导提前去填地址，与具体待领奖励无关。
func (s *AddressService) NeedAddressPreheat(viewerUserId, reqElderUserId string) (bool, error) {
	elderUserId, err := GetFamilyService().ResolveElder(viewerUserId, reqElderUserId)
	if err != nil {
		return false, err
	}
	list, err := s.shippingAddressMapper.ListByElder(elderUserId)
	if err != nil {
		return false, err
	}
	return len(list) == 0, nil
}

// RemindAddressPreheat 奖励地址预热提醒扫描（PRD §6.6.3，每月25号定时任务调用）：
// 遍历全部 ACTIVE 守护关系，对「老人地址簿仍为空」的，向其守护子女发预热提醒(best-effort)。
// period=YYYY-MM 入 message 幂等键，当月重跑/多关系扫到同子女只发一次。返回实际发提醒的关系数。
// 单条关系判定/发送失败只记日志、继续处理其余，不中断整体扫描。
func (s *AddressService) RemindAddressPreheat(period string) (int, error) {
	relations, err := s.guardianshipMapper.ListAllActive()
	if err != nil {
		return 0, fmt.Errorf("preheat scan list active guardianship: %w", err)
	}
	reminded := 0
	for i := range relations {
		elderUserId := relations[i].ElderUserId
		guardianUserId := relations[i].GuardianUserId
		addrs, err := s.shippingAddressMapper.ListByElder(elderUserId)
		if err != nil {
			logrus.WithError(err).WithField("elderUserId", elderUserId).
				Warn("preheat scan list address failed, skip this relation")
			continue
		}
		if len(addrs) > 0 {
			continue // 已有地址，不提醒
		}
		// relation=该子女对老人的称呼（模板「客户名称」）；欠料名称为固定文案，翻译层补
		params := jsons.JSONObject{"period": period, "relation": relations[i].Relation}
		GetMessageService().SendAddressPreheat(guardianUserId, elderUserId, period, params)
		reminded++
	}
	return reminded, nil
}

// ResolveShippingAddress 领取下单时按 addressId 选地址（问题3 连锁）：
//   - 传了 addressId：取该条，并校验确属本老人（防越权用他人地址ID）；
//   - 没传：取该老人默认地址。
//
// 无可用地址返 ErrAddressRequired（前端跳「添加收货地址」页）。返回的 model.Address 可直接快照进 RewardClaim。
func (s *AddressService) ResolveShippingAddress(elderUserId, addressId string) (model.Address, error) {
	if addressId != "" {
		a, err := s.shippingAddressMapper.GetByAddressId(addressId)
		if err != nil {
			return model.Address{}, err
		}
		if a == nil || a.ElderUserId != elderUserId {
			return model.Address{}, errors.NewWithArgs(wnconst.ErrAddressRequired)
		}
		return a.Address, nil
	}
	def, err := s.shippingAddressMapper.GetDefaultByElder(elderUserId)
	if err != nil {
		return model.Address{}, err
	}
	if def == nil {
		return model.Address{}, errors.NewWithArgs(wnconst.ErrAddressRequired)
	}
	return def.Address, nil
}

// ensureAccessByAddressId 按 addressId 反查归属老人并校验 viewer 可访问（模式B 共用）。
// 地址不存在或无权访问均返 ErrNotGuardian（不区分，避免撞库探测地址ID是否存在）。
func (s *AddressService) ensureAccessByAddressId(viewerUserId, addressId string) (*model.ShippingAddress, error) {
	a, err := s.shippingAddressMapper.GetByAddressId(addressId)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, errors.NewWithArgs(wnconst.ErrNotGuardian)
	}
	if err := GetFamilyService().EnsureCanAccess(viewerUserId, a.ElderUserId); err != nil {
		return nil, err
	}
	return a, nil
}

// listByElderInTx 事务内查某老人地址（删默认后补默认用），排序同 ListByElder
func (s *AddressService) listByElderInTx(tx *gorm.DB, elderUserId string) ([]model.ShippingAddress, error) {
	var list []model.ShippingAddress
	err := tx.Model(&model.ShippingAddress{}).
		Where("elder_user_id = ?", elderUserId).
		Order("is_default DESC, sort_no ASC, created_at ASC").
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list shipping address in tx %s: %w", elderUserId, err)
	}
	return list, nil
}
