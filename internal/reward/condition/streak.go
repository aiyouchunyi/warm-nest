// Package condition @Author larry
// @Date 2026/06/15
// @Desc 打卡天数算法（纯逻辑，无 IO，便于单测）

package condition

import "time"

const dateLayout = "2006-01-02"

// currentStreak 截至 today 的当前连续打卡天数（dates 升序去重，相邻差 1 天为续）
func currentStreak(dates []string, today string) int {
	set := make(map[string]bool, len(dates))
	for _, d := range dates {
		set[d] = true
	}
	t, err := time.Parse(dateLayout, today)
	if err != nil {
		return 0
	}
	if !set[today] {
		t = t.AddDate(0, 0, -1)
	}
	streak := 0
	for set[t.Format(dateLayout)] {
		streak++
		t = t.AddDate(0, 0, -1)
	}
	return streak
}

// monthDays dates 中属于 yearMonth(YYYY-MM) 的去重天数
func monthDays(dates []string, yearMonth string) int {
	seen := make(map[string]bool)
	for _, d := range dates {
		if len(d) >= 7 && d[:7] == yearMonth && !seen[d] {
			seen[d] = true
		}
	}
	return len(seen)
}

// cumulativeDays 累计去重打卡天数
func cumulativeDays(dates []string) int {
	seen := make(map[string]bool, len(dates))
	for _, d := range dates {
		seen[d] = true
	}
	return len(seen)
}
