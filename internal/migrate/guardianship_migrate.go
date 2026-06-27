// Package migrate @Author larry
// @Date 2026/06/18
// @Desc 守护关系表迁移（已落库表加列必须走迁移——EnableModel 只建新表不加列）
//
// 为什么需要：Guardianship 表已上线，新增 ActivatedAt 字段光改 struct 不会在表里生成列
// （框架 EnableModel 见到 HasTable 即跳过），查询/写入会报 Error 1054 Unknown column。
// 故用幂等的 AddColumnAfter 在 status 列后补 activated_at。
package migrate

import (
	"sync"

	"warm-nest/pkg/tool/database/mysqls/widgets"

	"warm-nest/internal/model"
)

// GuardianshipMigrate 守护关系表迁移
type GuardianshipMigrate struct{}

var guardianshipMigrate *GuardianshipMigrate
var guardianshipMigrateOnce sync.Once

// GetGuardianshipMigrate 获取守护关系表迁移单例
func GetGuardianshipMigrate() *GuardianshipMigrate {
	guardianshipMigrateOnce.Do(func() {
		guardianshipMigrate = &GuardianshipMigrate{}
	})
	return guardianshipMigrate
}

// Migrate 在 status 列后补 activated_at（幂等：已存在则跳过）
func (m *GuardianshipMigrate) Migrate() error {
	return widgets.AddColumnAfter(model.Guardianship{}, "ActivatedAt", "Status")
}
