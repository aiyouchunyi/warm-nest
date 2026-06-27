// Package migrate @Author larry
// @Date 2026/06/22
// @Desc 小程序消息表迁移（§5.3：补打卡标记「已打卡」所需字段）
//
// 为什么需要：Message 表已上线，新增 BizDate / ElderCheckedAt 字段光改 struct 不会在表里生成列
// （框架 EnableModel 见 HasTable 即跳过），查询/写入会报 Error 1054 Unknown column。
// 故用幂等 AddColumnAfter 在 ref_check_in_id 列后依次补两列。
// 注：i_elder_bizdate 复合索引由 EnableModel 的 AutoMigrate 在列就绪后补（加列在前、索引随后），
//
//	存量未打卡提醒无 biz_date（空串）不影响——补打卡标记只针对当日新发提醒。
package migrate

import (
	"sync"

	"warm-nest/pkg/tool/database/mysqls/widgets"

	"warm-nest/internal/model"
)

// MessageMigrate 小程序消息表迁移
type MessageMigrate struct{}

var messageMigrate *MessageMigrate
var messageMigrateOnce sync.Once

// GetMessageMigrate 获取小程序消息表迁移单例
func GetMessageMigrate() *MessageMigrate {
	messageMigrateOnce.Do(func() {
		messageMigrate = &MessageMigrate{}
	})
	return messageMigrate
}

// Migrate 在 ref_check_in_id 列后补 biz_date，再补 elder_checked_at（幂等：已存在则跳过）
func (m *MessageMigrate) Migrate() error {
	if err := widgets.AddColumnAfter(model.Message{}, "BizDate", "RefCheckInId"); err != nil {
		return err
	}
	return widgets.AddColumnAfter(model.Message{}, "ElderCheckedAt", "BizDate")
}
