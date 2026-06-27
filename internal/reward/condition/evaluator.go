// Package condition @Author larry
// @Date 2026/06/15
// @Desc 奖励达成条件策略（接口 + Engine 注册，加类型只加文件不改老代码）

package condition

import (
	"fmt"
	"sync"

	"warm-nest/pkg/kinds/jsons"

	"warm-nest/internal/model"
)

// EvalResult 达成判定结果
type EvalResult struct {
	Hit       bool             // 是否达成
	PeriodKey string           // 周期标识（防重发）
	Snap      jsons.JSONObject // 达成快照（进度）
}

// Evaluator 达成条件策略：每种 ConditionType 一个实现
type Evaluator interface {
	Type() string
	// Evaluate 用打卡日期序列(升序去重) + 服务端当日(Asia/Shanghai) + 守护授权日 + 任务规则，判是否达成。
	// effectiveDate(YYYY-MM-DD，Asia/Shanghai)仅 monthly 用于首月规则；continuous/cumulative 忽略。
	Evaluate(dates []string, today, effectiveDate string, task model.RewardTask) EvalResult
}

// Engine 条件引擎：按 type 取策略
type Engine struct {
	evaluators map[string]Evaluator
}

var engine *Engine
var engineOnce sync.Once

// GetEngine 获取条件引擎单例（注册全部策略）
func GetEngine() *Engine {
	engineOnce.Do(func() {
		engine = &Engine{evaluators: map[string]Evaluator{}}
		for _, e := range []Evaluator{
			&monthlyEvaluator{},
			&continuousEvaluator{},
			&cumulativeEvaluator{},
		} {
			engine.evaluators[e.Type()] = e
		}
	})
	return engine
}

// Evaluate 按任务的 ConditionType 取策略判定（无 switch）
func (e *Engine) Evaluate(dates []string, today, effectiveDate string, task model.RewardTask) (EvalResult, error) {
	ev, ok := e.evaluators[task.ConditionType]
	if !ok {
		return EvalResult{}, fmt.Errorf("unknown condition type: %s", task.ConditionType)
	}
	return ev.Evaluate(dates, today, effectiveDate, task), nil
}
