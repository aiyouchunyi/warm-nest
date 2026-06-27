// Package condition @Author larry
// @Date 2026/06/15
// @Desc 连续打卡 N 天

package condition

import (
	"fmt"

	"warm-nest/pkg/kinds/jsons"

	"warm-nest/internal/model"
)

type continuousEvaluator struct{}

func (e *continuousEvaluator) Type() string { return model.ConditionContinuousCheckIn }

func (e *continuousEvaluator) Evaluate(dates []string, today, _ string, task model.RewardTask) EvalResult {
	streak := currentStreak(dates, today)
	hit := streak >= task.ConditionValue
	// periodKey 用「达成日 + 阈值」，保证一个连续段达成只发一次、阈值变更不串
	return EvalResult{
		Hit:       hit,
		PeriodKey: fmt.Sprintf("continuous-%s-%d", today, task.ConditionValue),
		Snap:      jsons.JSONObject{"continuousDays": streak},
	}
}
