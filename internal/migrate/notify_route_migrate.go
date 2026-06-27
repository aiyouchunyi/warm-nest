// Package migrate @Author larry
// @Date 2026/06/22
// @Desc 通知路由配置初始化（PRD §5.2 按场景配渠道）
//
// 为什么需要：notify_route 表本身由 EnableModel 自动建（同步早于本异步迁移），但默认的
// 「场景→渠道」映射不会自动出现。本迁移把 4 个内置场景按默认渠道(template)各 seed 一条，
// 之后运营在后台改渠道/模板ID。templateId 留空——由运营在管理接口按真实平台模板ID补。
//
// 已落库表加列必须走迁移（框架 EnableModel 只建新表、HasTable 命中即跳过，不加列，否则写 page_path
// 报 Error 1054 Unknown column）：故先 AddColumnAfter 补 page_path，再回填存量行预设跳转页，最后 seed 新场景。
//
// 幂等：列已存在跳过；存量行仅在 page_path 为空时回填（不覆盖运营改过的值）；该场景已有配置则不重复 seed。
package migrate

import (
	"fmt"
	"sync"

	"warm-nest/pkg/tool/database/mysqls"
	"warm-nest/pkg/tool/database/mysqls/widgets"

	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
	"warm-nest/internal/preset"
)

// NotifyRouteMigrate 通知路由配置初始化
type NotifyRouteMigrate struct {
	notifyRouteMapper *mapper.NotifyRouteMapper
}

var notifyRouteMigrate *NotifyRouteMigrate
var notifyRouteMigrateOnce sync.Once

// GetNotifyRouteMigrate 获取通知路由配置初始化单例
func GetNotifyRouteMigrate() *NotifyRouteMigrate {
	notifyRouteMigrateOnce.Do(func() {
		notifyRouteMigrate = &NotifyRouteMigrate{
			notifyRouteMapper: mapper.GetNotifyRouteMapper(),
		}
	})
	return notifyRouteMigrate
}

// Migrate 加列 page_path → 回填存量行预设跳转页 → seed 新场景（各步幂等）。
func (m *NotifyRouteMigrate) Migrate() error {
	// 1. 加列 page_path（在 template_id 后；幂等，已存在跳过）——必须先于回填/seed，否则写该列报 Unknown column
	if err := widgets.AddColumnAfter(model.NotifyRoute{}, "PagePath", "TemplateId"); err != nil {
		return fmt.Errorf("notify route migrate add page_path column: %w", err)
	}

	// 2. seed 新场景 + 回填存量行的空 page_path（按预设；不覆盖运营改过的非空值）
	for _, route := range preset.NotifyRoutes() {
		existing, err := m.notifyRouteMapper.GetByScene(route.Scene)
		if err != nil {
			return fmt.Errorf("notify route migrate check %s: %w", route.Scene, err)
		}
		if existing != nil {
			// 存量行：仅当 page_path 为空时回填预设（加列后存量行该列为空，补上让 DB 即配；非空=运营改过，不动）
			if existing.PagePath == "" && route.PagePath != "" {
				if err := mysqls.DB().Model(&model.NotifyRoute{}).
					Where("scene = ? AND (page_path IS NULL OR page_path = '')", route.Scene).
					Update("page_path", route.PagePath).Error; err != nil {
					return fmt.Errorf("notify route migrate backfill page_path %s: %w", route.Scene, err)
				}
			}
			continue // 其余字段幂等：已有配置不覆盖
		}
		r := route
		if err := m.notifyRouteMapper.Create(&r); err != nil {
			return fmt.Errorf("notify route migrate seed %s: %w", route.Scene, err)
		}
	}
	return nil
}
