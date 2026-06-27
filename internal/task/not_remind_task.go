// Package task @Author larry
// @Date 2026/06/15
// @Desc 未打卡提醒任务（PRD §8.3 两段式：先提醒老人本人 → 宽限 N 分钟 → 老人仍未打卡才提醒子女）
//
// 轮询模型：每 5 分钟跑一次，遍历所有已设提醒时间的老人，按「当前时刻距其约定时间多久」分两段触发。
// 支持任意提醒时间（不再限 09/10/11/12 固定整点）——老人/子女自由设的提醒时间都能精确生效。
// 窗口判定 + dedupKey 按天去重，重复轮询不重发。

package task

import (
	"sync"
	"time"

	"warm-nest/internal/service"
)

// NotRemindTask 未打卡提醒任务
type NotRemindTask struct{}

var notRemindTask *NotRemindTask
var notRemindTaskOnce sync.Once

// GetNotRemindTask 获取未打卡提醒任务单例
func GetNotRemindTask() *NotRemindTask {
	notRemindTaskOnce.Do(func() {
		notRemindTask = &NotRemindTask{}
	})
	return notRemindTask
}

// Poll 每 5 分钟轮询：按各老人约定提醒时间分两段触发未打卡提醒（PRD §8.3）。
func (t *NotRemindTask) Poll(args ...any) error {
	return service.GetCheckInService().RemindByPolling(time.Now())
}
