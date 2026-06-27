// Package condition @Author larry
// @Date 2026/06/15
// @Desc 累计打卡满 N 天

package condition

import (
	"fmt"

	"warm-nest/pkg/kinds/jsons"

	"warm-nest/internal/model"
)

type cumulativeEvaluator struct{}

func (e *cumulativeEvaluator) Type() string { return model.ConditionCumulativeCheckIn }

func (e *cumulativeEvaluator) Evaluate(dates []string, _, _ string, task model.RewardTask) EvalResult {
	days := cumulativeDays(dates)
	return EvalResult{
		Hit:       days >= task.ConditionValue,
		PeriodKey: fmt.Sprintf("cumulative-%d", task.ConditionValue),
		Snap:      jsons.JSONObject{"cumulativeDays": days},
	}
}
