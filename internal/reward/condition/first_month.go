// Package condition @Author larry
// @Date 2026/06/16
// @Desc 首个不完整月份达标线算法（PRD §6.6.1.1，纯逻辑无 IO 便于单测）
//
// 月度奖励达标线统一为「窗口天数 − 容错天数」：
//   - 整月（§6.6.1）：窗口 = 当月自然天数 → 达标线 = 当月天数 − 3（二月 25、大月 28）。
//   - 首月（§6.6.1.1，非 1 号入会）：窗口 = 授权日到月末剩余天数 R（含当天）→ 达标线 = R − 3；
//     R < 15 时首月不计奖励（仅免费体验，老人照常打卡、子女照常收通知）。
//   - 1 号入会按整月处理（窗口即整月，无特殊规则）。
//
// 「授权日」由 service 层按 Asia/Shanghai 把 Guardianship.CreatedAt(毫秒) 转成 YYYY-MM-DD 传入，
// 本文件只做纯日期运算（操作 YYYY-MM-DD / YYYY-MM 字符串），不碰时区与 IO。

package condition

import "time"

const (
	// monthlyToleranceDays 月度达标容错天数（全勤 − 3，整月与首月统一口径）
	monthlyToleranceDays = 3
	// firstMonthMinRemaining 首月计奖的剩余天数下限（R < 15 不计奖励，呼应运费净损耗一刀切）
	firstMonthMinRemaining = 15
)

// daysInMonth 返回 yearMonth(YYYY-MM) 的自然天数；解析失败返回 0
func daysInMonth(yearMonth string) int {
	t, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return 0
	}
	// 下月 1 号往前一天即本月最后一天
	return t.AddDate(0, 1, -1).Day()
}

// firstMonthRemaining 授权日到当月月末的剩余天数 R（含授权日当天）。
// effectiveDate(YYYY-MM-DD) 必须落在 yearMonth(YYYY-MM) 内，否则返回 0。
func firstMonthRemaining(effectiveDate, yearMonth string) int {
	if len(effectiveDate) < 10 || effectiveDate[:7] != yearMonth {
		return 0
	}
	dim := daysInMonth(yearMonth)
	if dim == 0 {
		return 0
	}
	t, err := time.Parse(dateLayout, effectiveDate)
	if err != nil {
		return 0
	}
	return dim - t.Day() + 1
}

// monthlyThreshold 计算 yearMonth 月的 monthly 达标线。
//   - skip=true：该月不计奖励（首月且 R < 15），threshold 无意义。
//   - 首月（授权日在本月且非 1 号）：threshold = R − 3。
//   - 其余（整月 / 1 号入会 / 授权日不在本月）：threshold = 当月天数 − 3。
func monthlyThreshold(effectiveDate, yearMonth string) (threshold int, skip bool) {
	// 首月判定：授权日落在本月、且非 1 号入会
	if len(effectiveDate) >= 10 && effectiveDate[:7] == yearMonth && effectiveDate[8:10] != "01" {
		r := firstMonthRemaining(effectiveDate, yearMonth)
		if r < firstMonthMinRemaining {
			return 0, true
		}
		return r - monthlyToleranceDays, false
	}
	return daysInMonth(yearMonth) - monthlyToleranceDays, false
}
