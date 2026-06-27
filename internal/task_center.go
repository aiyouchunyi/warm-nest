// Package internal @Author larry
// @Desc 任务注册中心
//
// 【新增 task 后必做】在 RegisterTask() 切片里追加一行：
//
//	tasks.NewTask("<cron 表达式>", "<TaskName>", task.GetXxxTask().Refresh),
//
// cron 表达式 6 段制：秒 分 时 日 月 周
// 示例：
//
//	"0 */5 * * * *"  每 5 分钟
//	"0 0 7 * * *"    每天早 7 点
//	"10 * * * * *"   每分钟第 10 秒
package internal

import (
	"warm-nest/pkg/tool/tasks"

	"warm-nest/internal/task"
)

// RegisterTask 注册定时任务（前提：1 期单实例部署，见计划 §6.1 多实例前提）
func RegisterTask() []tasks.ITask {
	return []tasks.ITask{
		// 微信 access_token 刷新（每 90 分钟）
		tasks.NewTask("0 */90 * * * *", "WeChatTokenRefresh", task.GetWeChatTokenTask().Refresh),
		// 未打卡提醒（每 5 分钟轮询，支持任意提醒时间）：按各老人约定时间分两段触发——
		// 到点提醒老人本人 → 宽限 30min 老人仍未打卡 → 兜底提醒子女（PRD §8.3）。
		tasks.NewTask("0 */5 * * * *", "NotRemindPoll", task.GetNotRemindTask().Poll),
		// 邀请过期扫描（每 30 分钟）
		tasks.NewTask("0 */30 * * * *", "InvitationExpireScan", task.GetInvitationExpireTask().Scan),
		// 奖励地址预热提醒（每月 25 号 10:00，收货地址仍为空的子女提醒补地址，PRD §6.6.3）
		tasks.NewTask("0 0 10 25 * *", "AddressPreheatRemind", task.GetAddressPreheatTask().Remind),
	}
}
