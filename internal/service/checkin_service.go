// Package service @Author larry
// @Date 2026/06/15
// @Desc 打卡服务（核心闭环）

package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/kinds/jsons"
	"warm-nest/pkg/utils/rands"

	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
)

// dateLayout 打卡自然日格式
const dateLayout = "2006-01-02"

// dateTimeLayout 通知模板的时间字段格式（YYYY-MM-DD HH:mm，微信 time 类型字段用）
const dateTimeLayout = "2006-01-02 15:04"

// CheckInService 打卡服务
type CheckInService struct {
	checkInMapper      *mapper.CheckInMapper
	elderProfileMapper *mapper.ElderProfileMapper
}

var checkInService *CheckInService
var checkInServiceOnce sync.Once

// GetCheckInService 获取打卡服务单例
func GetCheckInService() *CheckInService {
	checkInServiceOnce.Do(func() {
		checkInService = &CheckInService{
			checkInMapper:      mapper.GetCheckInMapper(),
			elderProfileMapper: mapper.GetElderProfileMapper(),
		}
	})
	return checkInService
}

// DoCheckIn 老人打卡：幂等（一日一卡）→ 落库 → 发通知 → 触发奖励评估
func (s *CheckInService) DoCheckIn(elderUserId, photoUrl, weather, city string) (*model.CheckIn, error) {
	date := time.Now().In(shanghai).Format(dateLayout)

	// 幂等：今日已打卡直接返回已有记录
	existing, err := s.checkInMapper.GetByUserDate(elderUserId, date)
	if err != nil {
		return nil, fmt.Errorf("do checkin idempotent query: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	checkIn := &model.CheckIn{
		CheckInId:   rands.Numeric(),
		ElderUserId: elderUserId,
		CheckInDate: date,
		Kind:        model.CheckInKindNormal,
		PhotoUrl:    photoUrl,
		Weather:     weather,
		City:        city,
	}
	if err = s.checkInMapper.Create(checkIn); err != nil {
		return nil, fmt.Errorf("do checkin create: %w", err)
	}

	// 发打卡通知（best-effort，失败不回滚打卡这一核心事实）
	// checkInTime=打卡实际发生时刻（模板「签到时间」）；relation 由 message_service 逐子女注入
	params := jsons.JSONObject{
		"checkInDate": date,
		"weather":     weather,
		"city":        city,
		"checkInTime": time.Now().In(shanghai).Format(dateTimeLayout),
	}
	if err = GetMessageService().SendCheckInNotice(checkIn, params); err != nil {
		logrus.WithError(err).WithField("checkInId", checkIn.CheckInId).Warn("send checkin notice failed")
	}

	// 补打卡：把当日已发出的未打卡提醒消息流标记「已打卡」（§5.3，best-effort，不撤回仅标状态）
	if _, err = GetMessageService().MarkRemindChecked(elderUserId, date, time.Now().In(shanghai).UnixMilli()); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"elderUserId": elderUserId, "date": date}).
			Warn("mark not-remind messages checked failed")
	}

	// 触发奖励评估（best-effort）
	if err = GetRewardService().EvaluateRewards(elderUserId); err != nil {
		logrus.WithError(err).WithField("elderUserId", elderUserId).Warn("evaluate rewards failed")
	}

	return checkIn, nil
}

// TodayChecked 今日是否已打卡
func (s *CheckInService) TodayChecked(elderUserId string) (bool, error) {
	date := time.Now().In(shanghai).Format(dateLayout)
	c, err := s.checkInMapper.GetByUserDate(elderUserId, date)
	if err != nil {
		return false, err
	}
	return c != nil, nil
}

// MonthCheckIns 查某月打卡记录
func (s *CheckInService) MonthCheckIns(elderUserId, yearMonth string) ([]model.CheckIn, error) {
	return s.checkInMapper.ListByMonth(elderUserId, yearMonth)
}

// GetCheckInDetail 按 checkInId 查单日打卡详情（问题4，公用接口模式B）：
// 前端只传 checkInId，先查出记录归属老人，再经守护关系访问校验防越权（撞库看他人打卡）。
// 记录不存在返 (nil,nil)，权限不过返 ErrNotGuardian。
func (s *CheckInService) GetCheckInDetail(viewerUserId, checkInId string) (*model.CheckIn, error) {
	c, err := s.checkInMapper.GetByCheckInId(checkInId)
	if err != nil {
		return nil, fmt.Errorf("get checkin detail %s: %w", checkInId, err)
	}
	if c == nil {
		return nil, nil
	}
	if err = GetFamilyService().EnsureCanAccess(viewerUserId, c.ElderUserId); err != nil {
		return nil, err
	}
	return c, nil
}

// 未打卡提醒宽限期：到约定时间先提醒老人本人，宽限 graceMinutes 后老人仍未打卡才提醒子女（PRD §8.3）。
const remindGraceMinutes = 30

// pollIntervalMinutes 轮询任务的执行间隔（分钟）；窗口判定用，须与 task_center 注册的 cron 间隔一致。
// 每个时段（到点 / 到点+宽限）在一天内恰好命中一次轮询窗口 [x, x+interval)，不重不漏；
// 叠加 message dedupKey 按天去重双保险。
const pollIntervalMinutes = 5

// RemindByPolling 未打卡提醒轮询入口（每 pollIntervalMinutes 分钟一次）：
// 遍历所有已设提醒时间的老人，按「当前时刻距其约定时间多久」分两段触发（PRD §8.3），
// 支持任意提醒时间（不再限固定整点）。now 由任务传入（便于测试），按 Asia/Shanghai 解读。
//
// 判据（elapsed = now - 老人当日约定时间，单位分钟）：
//   - elapsed ∈ [0, interval)            → 第一段：到点，提醒老人本人（未关注则转子女兜底）
//   - elapsed ∈ [grace, grace+interval)  → 第二段：宽限期满老人仍未打卡，提醒子女
//   - 其他                                → 本轮不处理该老人
func (s *CheckInService) RemindByPolling(now time.Time) error {
	now = now.In(shanghai)
	date := now.Format(dateLayout)
	nowMinutes := now.Hour()*60 + now.Minute()

	profiles, err := s.elderProfileMapper.ListWithRemindTime()
	if err != nil {
		return fmt.Errorf("remind polling list profiles: %w", err)
	}
	for i := range profiles {
		rtMinutes, ok := parseRemindMinutes(profiles[i].RemindTime)
		if !ok {
			logrus.WithFields(logrus.Fields{"elderUserId": profiles[i].UserId, "remindTime": profiles[i].RemindTime}).
				Warn("remind polling skip: invalid remind time")
			continue
		}
		elapsed := nowMinutes - rtMinutes
		switch {
		case elapsed >= 0 && elapsed < pollIntervalMinutes:
			s.remindElderSelfOne(profiles[i].UserId, profiles[i].RemindTime, date)
		case elapsed >= remindGraceMinutes && elapsed < remindGraceMinutes+pollIntervalMinutes:
			s.remindGuardiansOne(profiles[i].UserId, profiles[i].RemindTime, date)
		}
	}
	return nil
}

// parseRemindMinutes 解析 HH:mm → 当日零点起的分钟数；格式非法返回 false。
func parseRemindMinutes(remindTime string) (int, bool) {
	t, err := time.Parse("15:04", remindTime)
	if err != nil {
		return 0, false
	}
	return t.Hour()*60 + t.Minute(), true
}

// notChecked 当日是否仍未打卡（查不出按"已打卡"保守处理，避免误推）。
func (s *CheckInService) notChecked(elderUserId, date string) bool {
	checkedIn, err := s.checkInMapper.GetByUserDate(elderUserId, date)
	if err != nil {
		logrus.WithError(err).WithField("elderUserId", elderUserId).Warn("remind check failed")
		return false
	}
	return checkedIn == nil
}

// remindElderSelfOne 未打卡提醒第一段（PRD §8.3）：到约定时间，处理单个当日未打卡老人。
//   - 已关注服务号 → 先提醒老人本人，留宽限期，第二段再兜底子女；
//   - 未关注服务号 → 老人本人收不到（无可达通道），此刻直接推子女兜底，不让子女白等宽限期。
//
// 已打卡的老人直接跳过（不推任何提醒）。
func (s *CheckInService) remindElderSelfOne(elderUserId, remindTime, date string) {
	if !s.notChecked(elderUserId, date) {
		return
	}
	// deadline=打卡截止时刻（当日+约定提醒时间，模板「截止时间」）；relation 逐子女注入
	params := jsons.JSONObject{"date": date, "remindTime": remindTime, "deadline": date + " " + remindTime}
	subscribed, err := GetFanService().IsSubscribedByUserId(elderUserId)
	if err != nil {
		// 关注态查不出时，保守按"已关注"走本人提醒，不漏第一段（子女仍有第二段兜底）
		logrus.WithError(err).WithField("elderUserId", elderUserId).Warn("remind self check subscribe failed, fallback to self remind")
		GetMessageService().SendElderSelfRemind(elderUserId, date, params)
		return
	}
	if subscribed {
		GetMessageService().SendElderSelfRemind(elderUserId, date, params)
		return
	}
	// 未关注服务号：老人本人无可达通道 → 直接推子女兜底（不等第二段）
	if err = GetMessageService().SendNotRemind(elderUserId, date, params); err != nil {
		logrus.WithError(err).WithField("elderUserId", elderUserId).Warn("send not-remind (unsubscribed elder) failed")
	}
}

// remindGuardiansOne 未打卡提醒第二段（PRD §8.3）：宽限期满老人仍未打卡，兜底提醒单个老人的子女。
// 此刻再查一次打卡状态，宽限内已补打卡的老人自然被排除（dedupKey 按天去重，重复轮询不重发）。
func (s *CheckInService) remindGuardiansOne(elderUserId, remindTime, date string) {
	if !s.notChecked(elderUserId, date) {
		return
	}
	// deadline=打卡截止时刻（当日+约定提醒时间，模板「截止时间」）；relation 逐子女注入
	params := jsons.JSONObject{"date": date, "remindTime": remindTime, "deadline": date + " " + remindTime}
	if err := GetMessageService().SendNotRemind(elderUserId, date, params); err != nil {
		logrus.WithError(err).WithField("elderUserId", elderUserId).Warn("send not-remind failed")
	}
}
