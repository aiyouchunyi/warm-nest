// Package migrate @Author larry
// @Date 2026/06/19
// @Desc 守护邀请表迁移（已落库表加列必须走迁移——EnableModel 只建新表不加列）
//
// 为什么需要：Invitation 表已上线，新增 City 线索字段光改 struct 不会在表里生成列
// （框架 EnableModel 见到 HasTable 即跳过），写入会报 Error 1054 Unknown column。
// 故用幂等的 AddColumnAfter 在 remind_time 列后补 city。
package migrate

import (
	"sync"

	"warm-nest/pkg/tool/database/mysqls/widgets"

	"warm-nest/internal/model"
)

// InvitationMigrate 守护邀请表迁移
type InvitationMigrate struct{}

var invitationMigrate *InvitationMigrate
var invitationMigrateOnce sync.Once

// GetInvitationMigrate 获取守护邀请表迁移单例
func GetInvitationMigrate() *InvitationMigrate {
	invitationMigrateOnce.Do(func() {
		invitationMigrate = &InvitationMigrate{}
	})
	return invitationMigrate
}

// Migrate 在 remind_time 列后补 city（幂等：已存在则跳过）
func (m *InvitationMigrate) Migrate() error {
	return widgets.AddColumnAfter(model.Invitation{}, "City", "RemindTime")
}
