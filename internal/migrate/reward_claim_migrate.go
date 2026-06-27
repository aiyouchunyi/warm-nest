// Package migrate @Author larry
// @Date 2026/06/21
// @Desc 领取记录表迁移（收货快照三字段 → 结构化 json）
//
// 为什么需要：RewardClaim 收货快照原为 receiver_name/receiver_phone/receiver_addr 三个 varchar 列，
// 改为单个 receiver_address json 列（与 ElderProfile.Address 同构，取数统一、保留省市区结构）。
// 已落库表只改 struct 不迁移 = 表里既有旧三列又缺新列（写入报 Unknown column）。
// 测试期旧快照不保留，直接 drop 旧三列 + add 新列；幂等可重跑。
package migrate

import (
	"sync"

	"warm-nest/pkg/tool/database/mysqls/widgets"

	"warm-nest/internal/model"
)

// RewardClaimMigrate 领取记录表迁移
type RewardClaimMigrate struct{}

var rewardClaimMigrate *RewardClaimMigrate
var rewardClaimMigrateOnce sync.Once

// GetRewardClaimMigrate 获取领取记录表迁移单例
func GetRewardClaimMigrate() *RewardClaimMigrate {
	rewardClaimMigrateOnce.Do(func() {
		rewardClaimMigrate = &RewardClaimMigrate{}
	})
	return rewardClaimMigrate
}

// Migrate 删旧收货三列 + 加结构化 receiver_address json 列（幂等：DropColumn 列不存在跳过，AddColumnAfter 已存在跳过）
func (m *RewardClaimMigrate) Migrate() error {
	for _, col := range []string{"receiver_name", "receiver_phone", "receiver_addr"} {
		if err := widgets.DropColumn(model.RewardClaim{}, col); err != nil {
			return err
		}
	}
	return widgets.AddColumnAfter(model.RewardClaim{}, "ReceiverAddress", "ExpressSyncAt")
}
