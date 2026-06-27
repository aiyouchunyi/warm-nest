// Package internal @Author larry
// @Date 2025/8/1 14:12
// @Desc

package internal

import (
	"warm-nest/pkg/app/widgets"

	"warm-nest/internal/migrate"
)

func RegisterMigrate() map[string]widgets.DoMigrate {
	return map[string]widgets.DoMigrate{
		// 守护关系表补 activated_at 列（绑定成立时刻，首月奖励窗口起点）
		"guardianshipActivatedAt": migrate.GetGuardianshipMigrate().Migrate,
		// 被守护人档案表补 elder_phone / guardian_phone 联系电话列
		"elderProfileContactPhones": migrate.GetElderProfileMigrate().Migrate,
		// 守护邀请表补 city 线索列（接受时落入档案供打卡取天气）
		"invitationCity": migrate.GetInvitationMigrate().Migrate,
		// 领取记录收货快照三字段 → 结构化 receiver_address json 列
		"rewardClaimReceiverJson": migrate.GetRewardClaimMigrate().Migrate,
		// 存量 ElderProfile.Address 单地址搬迁为地址簿首条默认地址（问题3）
		"shippingAddressSeedFromProfile": migrate.GetShippingAddressMigrate().Migrate,
		// 通知路由配置 seed 内置场景默认渠道（PRD §5.2 按场景配渠道）
		"notifyRouteSeed": migrate.GetNotifyRouteMigrate().Migrate,
		// 奖励任务规则 seed 预设规则（出厂内置，幂等不存在则插入；不被 WipeAll 清理）
		"rewardTaskSeed": migrate.GetRewardTaskMigrate().Migrate,
		// 消息表补 biz_date / elder_checked_at 列（§5.3 补打卡标记「已打卡」）
		"messageBizDateCheckedAt": migrate.GetMessageMigrate().Migrate,
	}
}
