// Package task @Author larry
// @Date 2026/06/22
// @Desc 奖励地址预热提醒（PRD §6.6.3：每月25号、收货地址仍为空的子女提醒补地址）
//
// 系统主动推送，非用户行为触发。扫描+判定+推送编排在 AddressService.RemindAddressPreheat，
// 本 task 仅是 cron 入口壳：算当月 period(YYYY-MM，Asia/Shanghai) 传入。
// 不强制不阻断领取——只是提前备好地址让发货更顺。

package task

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"warm-nest/internal/service"
)

// AddressPreheatTask 奖励地址预热提醒任务
type AddressPreheatTask struct{}

var addressPreheatTask *AddressPreheatTask
var addressPreheatTaskOnce sync.Once

// GetAddressPreheatTask 获取奖励地址预热提醒任务单例
func GetAddressPreheatTask() *AddressPreheatTask {
	addressPreheatTaskOnce.Do(func() {
		addressPreheatTask = &AddressPreheatTask{}
	})
	return addressPreheatTask
}

// Remind 扫描地址簿为空的老人、提醒其子女补地址（period 按 Asia/Shanghai 当月）
func (t *AddressPreheatTask) Remind(args ...any) error {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600) // tzdata 缺失兜底固定 +8，与奖励评估时区口径一致
	}
	period := time.Now().In(loc).Format("2006-01")
	n, err := service.GetAddressService().RemindAddressPreheat(period)
	if err != nil {
		return err
	}
	if n > 0 {
		logrus.WithFields(logrus.Fields{"period": period, "count": n}).Info("address preheat reminders sent")
	}
	return nil
}
