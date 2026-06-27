// Package condition @Author larry
// @Date 2026/06/15
// @Desc 本月打卡满 N 天

package condition

import (
	"warm-nest/pkg/kinds/jsons"

	"warm-nest/internal/model"
)

type monthlyEvaluator struct{}

func (e *monthlyEvaluator) Type() string { return model.ConditionMonthlyCheckIn }

// Evaluate 月度达标。达标线统一为「窗口天数 − 3」：整月用当月天数、首月用授权日到月末的剩余天数 R。
// 达标线不读 task.ConditionValue（动态按月/按授权日算，见 PRD §6.6.1 / §6.6.1.1）。
// 首月 R < 15 不计奖励（skip），仅免费体验，不建领取记录。
func (e *monthlyEvaluator) Evaluate(dates []string, today, effectiveDate string, _ model.RewardTask) EvalResult {
	ym := today[:7] // today 恒为 YYYY-MM-DD（reward_service 按 Asia/Shanghai 生成）
	threshold, skip := monthlyThreshold(effectiveDate, ym)
	days := monthDays(dates, ym)
	snap := jsons.JSONObject{"month": ym, "days": days, "threshold": threshold}
	if skip {
		// 首月剩余不足，本月不计奖励：永不命中（老人照常打卡、子女照常收通知，不受影响）
		snap["firstMonthSkipped"] = true
		return EvalResult{Hit: false, PeriodKey: "monthly-" + ym, Snap: snap}
	}
	return EvalResult{
		Hit:       days >= threshold,
		PeriodKey: "monthly-" + ym,
		Snap:      snap,
	}
}
