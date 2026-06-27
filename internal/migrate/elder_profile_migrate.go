// Package migrate @Author larry
// @Date 2026/06/19
// @Desc 被守护人档案表迁移（已落库表加列必须走迁移——EnableModel 只建新表不加列）
//
// 为什么需要：ElderProfile 表已上线，新增 ElderPhone/GuardianPhone 联系电话字段光改 struct
// 不会在表里生成列（框架 EnableModel 见到 HasTable 即跳过），查询/写入会报 Error 1054 Unknown column。
// 故用幂等的 AddColumnAfter 在 address 列后依次补两列。
package migrate

import (
	"fmt"
	"sync"

	"warm-nest/pkg/tool/database/mysqls"
	"warm-nest/pkg/tool/database/mysqls/widgets"

	"warm-nest/internal/model"
)

// ElderProfileMigrate 被守护人档案表迁移
type ElderProfileMigrate struct{}

var elderProfileMigrate *ElderProfileMigrate
var elderProfileMigrateOnce sync.Once

// GetElderProfileMigrate 获取被守护人档案表迁移单例
func GetElderProfileMigrate() *ElderProfileMigrate {
	elderProfileMigrateOnce.Do(func() {
		elderProfileMigrate = &ElderProfileMigrate{}
	})
	return elderProfileMigrate
}

// Migrate 在 address 列后补 elder_phone / guardian_phone（幂等：已存在则跳过）；
// 并把 address 列类型由 varchar 改为 json（问题11 地址结构化，测试期不兼容旧值、直接改列）。
func (m *ElderProfileMigrate) Migrate() error {
	if err := widgets.AddColumnAfter(model.ElderProfile{}, "ElderPhone", "Address"); err != nil {
		return err
	}
	if err := widgets.AddColumnAfter(model.ElderProfile{}, "GuardianPhone", "ElderPhone"); err != nil {
		return err
	}
	// 地址结构化（问题11）：改列类型为 json 前，先把旧 varchar 值（纯文本地址/空串/非法 JSON）
	// 统一重置为合法空对象 '{}'——否则 ALTER MODIFY ... json 遇非法 JSON 会报 3140 致迁移失败、服务起不来。
	// 测试期旧地址不保留，重置无业务损失。已是合法 JSON 的不动（幂等可重跑）。
	if err := mysqls.DB().Exec(
		"UPDATE t_elder_profile SET address = '{}' WHERE address IS NULL OR NOT JSON_VALID(address)",
	).Error; err != nil {
		return fmt.Errorf("normalize elder_profile.address before json migrate: %w", err)
	}
	// 按 model 当前 tag(type:json) 重定义 address 列；幂等可重跑
	return widgets.ModifyColumn(model.ElderProfile{}, "Address")
}
