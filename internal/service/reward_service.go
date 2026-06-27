// Package service @Author larry
// @Date 2026/06/15
// @Desc 奖励服务（达成评估 + EGG 内联发放 + 受控状态推进）
//
// 1 期奖励仅一盒鸡蛋：达成判定走 condition 策略引擎（无 switch）；发放内联处理 EGG、
// 不配 Granter 工厂（单实现避免过早抽象，出现第二种奖励再抽）。状态用受控方法推进，
// 不起 machine 引擎（鸡蛋发货无自动流转语义）。

package service

import (
	"fmt"
	"sync"
	"time"

	"warm-nest/pkg/utils/rands"

	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
	"warm-nest/internal/reward/condition"
)

// RewardService 奖励服务
type RewardService struct {
	rewardTaskMapper   *mapper.RewardTaskMapper
	rewardClaimMapper  *mapper.RewardClaimMapper
	checkInMapper      *mapper.CheckInMapper
	guardianshipMapper *mapper.GuardianshipMapper
}

var rewardService *RewardService
var rewardServiceOnce sync.Once

// GetRewardService 获取奖励服务单例
func GetRewardService() *RewardService {
	rewardServiceOnce.Do(func() {
		rewardService = &RewardService{
			rewardTaskMapper:   mapper.GetRewardTaskMapper(),
			rewardClaimMapper:  mapper.GetRewardClaimMapper(),
			checkInMapper:      mapper.GetCheckInMapper(),
			guardianshipMapper: mapper.GetGuardianshipMapper(),
		}
	})
	return rewardService
}

// shanghai 奖励评估统一时区：月末/当日/授权日判定均按 UTC+8（业务面向中国老人），
// 不随部署环境时区漂移。CheckInDate 落库仍走打卡服务的服务端时区（不在本服务职责内）。
var shanghai = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600) // tzdata 缺失兜底
	}
	return loc
}()

// effectiveDate 守护授权日(YYYY-MM-DD，Asia/Shanghai)：取该老人最早一条 ACTIVE 关系的 CreatedAt。
// 1 期一对一只有一条；多对多时最早绑定即首月起算点。无 active 关系返回 ""（monthly 退化为整月规则）。
func (s *RewardService) effectiveDate(elderUserId string) (string, error) {
	relations, err := s.guardianshipMapper.ListByElder(elderUserId)
	if err != nil {
		return "", fmt.Errorf("list guardianship by elder %s: %w", elderUserId, err)
	}
	var earliest int64
	for i := range relations {
		if c := relations[i].CreatedAt; c > 0 && (earliest == 0 || c < earliest) {
			earliest = c
		}
	}
	if earliest == 0 {
		return "", nil
	}
	return time.UnixMilli(earliest).In(shanghai).Format(dateLayout), nil
}

// EvaluateRewards 打卡后评估全部启用任务，达成且未发过则建领取记录（PENDING）
func (s *RewardService) EvaluateRewards(elderUserId string) error {
	tasks, err := s.rewardTaskMapper.ListEnabled()
	if err != nil {
		return fmt.Errorf("evaluate rewards list enabled tasks: %w", err)
	}
	if len(tasks) == 0 {
		return nil
	}
	dates, err := s.checkInMapper.ListDatesByUser(elderUserId)
	if err != nil {
		return fmt.Errorf("evaluate rewards list checkin dates: %w", err)
	}
	today := time.Now().In(shanghai).Format(dateLayout)
	effectiveDate, err := s.effectiveDate(elderUserId)
	if err != nil {
		return fmt.Errorf("evaluate rewards effective date: %w", err)
	}

	engine := condition.GetEngine()
	for i := range tasks {
		task := tasks[i]
		result, err := engine.Evaluate(dates, today, effectiveDate, task)
		if err != nil {
			return fmt.Errorf("evaluate rewards engine evaluate taskKey=%s: %w", task.TaskKey, err)
		}
		if !result.Hit {
			continue
		}
		exist, err := s.rewardClaimMapper.ExistByUserTaskPeriod(elderUserId, task.TaskKey, result.PeriodKey)
		if err != nil {
			return fmt.Errorf("evaluate rewards check exist taskKey=%s: %w", task.TaskKey, err)
		}
		if exist {
			continue
		}
		claim := &model.RewardClaim{
			ClaimId:      rands.Numeric(),
			ElderUserId:  elderUserId,
			TaskKey:      task.TaskKey,
			PeriodKey:    result.PeriodKey,
			RewardKind:   task.RewardKind,
			RewardName:   task.RewardName,
			RewardSpec:   task.RewardSpec,
			Quantity:     task.Quantity,
			RewardParams: task.RewardParams,
			AchievedSnap: result.Snap,
		}
		claim.Status = model.ClaimStatusPending
		if err = s.rewardClaimMapper.Create(claim); err != nil {
			return fmt.Errorf("create reward claim: %w", err)
		}
	}
	return nil
}

// 状态流转（领取/发货/签收）由 reward/machine 包的 machine 引擎驱动，不在此处受控推进。

// ListByUser 查用户领取记录
func (s *RewardService) ListByUser(elderUserId string) ([]model.RewardClaim, error) {
	return s.rewardClaimMapper.ListByUser(elderUserId)
}

// GetClaimDetail 按 claimId 查领取详情（问题5，公用接口模式B）：
// 前端只传 claimId，先查出记录归属老人，再经守护关系访问校验防越权（撞库看他人领取）。
// 记录不存在返 (nil,nil)，权限不过返 ErrNotGuardian。
func (s *RewardService) GetClaimDetail(viewerUserId, claimId string) (*model.RewardClaim, error) {
	c, err := s.rewardClaimMapper.GetByClaimId(claimId)
	if err != nil {
		return nil, fmt.Errorf("get claim detail %s: %w", claimId, err)
	}
	if c == nil {
		return nil, nil
	}
	if err = GetFamilyService().EnsureCanAccess(viewerUserId, c.ElderUserId); err != nil {
		return nil, err
	}
	return c, nil
}

// EnsureClaimAccess 领取前归属校验（模式B，防越权领他人奖励）：
// 按 claimId 查出归属老人，校验 viewer 是本人或其 ACTIVE 守护人（双端可领），否则 ErrNotGuardian。
// claimId 不存在返 not found，供 reward/claim 在推进状态机前调用。
func (s *RewardService) EnsureClaimAccess(viewerUserId, claimId string) error {
	c, err := s.rewardClaimMapper.GetByClaimId(claimId)
	if err != nil {
		return fmt.Errorf("ensure claim access %s: %w", claimId, err)
	}
	if c == nil {
		return fmt.Errorf("reward claim not found: %s", claimId)
	}
	return GetFamilyService().EnsureCanAccess(viewerUserId, c.ElderUserId)
}

// 月份打卡状态（问题10 截图：每月卡片右下角状态标）
const (
	MonthStatusInProgress = "IN_PROGRESS" // 打卡中（该月无达成领取记录）
	MonthStatusClaimable  = "CLAIMABLE"   // 可领取（该月已达成、领取记录 PENDING）
	MonthStatusClaimed    = "CLAIMED"     // 已领取（该月领取记录已 CLAIMED 及以后）
)

// MonthOverview 打卡月份记录页（问题10）：顶部累计汇总 + 逐月打卡进度与领取状态。
type MonthOverview struct {
	TotalCheckInDays int           `json:"totalCheckInDays"` // 累计打卡天数
	ClaimedEggCount  int           `json:"claimedEggCount"`  // 已领鸡蛋数（领取记录已领取及以后，含非月度奖励，与月份列表不逐行对应）
	ClaimableMonths  int           `json:"claimableMonths"`  // 待领月份数（= months 中 CLAIMABLE 条数，仅月度奖励，与列表逐行对应）
	Months           []MonthRecord `json:"months"`           // 逐月记录（当月在前，倒序）
}

// MonthRecord 单月打卡记录
type MonthRecord struct {
	YearMonth   string `json:"yearMonth"`   // YYYY-MM
	CheckedDays int    `json:"checkedDays"` // 当月已打卡天数
	TotalDays   int    `json:"totalDays"`   // 当月自然天数
	Status      string `json:"status"`      // IN_PROGRESS / CLAIMABLE / CLAIMED
	IsCurrent   bool   `json:"isCurrent"`   // 是否当前月
	ClaimId     string `json:"claimId"`     // 该月领取记录ID（无则空，前端凭此调领取）
}

// MonthOverview 组装某老人的打卡月份概览（从授权月到当月，倒序）。
// 月份状态判据（按 monthly-YYYY-MM 周期的领取记录定，无记录=打卡中；见 MonthStatus* 常量）。
func (s *RewardService) MonthOverview(elderUserId string) (*MonthOverview, error) {
	dates, err := s.checkInMapper.ListDatesByUser(elderUserId)
	if err != nil {
		return nil, fmt.Errorf("month overview list checkin dates: %w", err)
	}
	claims, err := s.rewardClaimMapper.ListByUser(elderUserId)
	if err != nil {
		return nil, fmt.Errorf("month overview list claims: %w", err)
	}
	effectiveDate, err := s.effectiveDate(elderUserId)
	if err != nil {
		return nil, fmt.Errorf("month overview effective date: %w", err)
	}

	// 每月打卡天数（去重，按 YYYY-MM 聚合）
	daysByMonth := map[string]int{}
	seen := map[string]bool{}
	for _, d := range dates {
		if len(d) < 7 || seen[d] {
			continue
		}
		seen[d] = true
		daysByMonth[d[:7]]++
	}

	// 按月归集领取记录（monthly-YYYY-MM 周期）+ 顶部汇总。
	// ClaimedEggCount 是顶部独立汇总，含非月度奖励（连续/累计），不与月份列表逐行对应，故全量累加。
	// ClaimableMonths 不在此累加：它必须 = months 列表中 CLAIMABLE 条数（仅月度奖励），
	// 否则会把 continuous-*/cumulative-* 等非月度 PENDING 也计入、与列表对不上（问题2 根因），
	// 改在 buildMonthRecords 后按列表实际状态统计。
	ov := &MonthOverview{TotalCheckInDays: len(seen)}
	claimByMonth := map[string]model.RewardClaim{}
	for i := range claims {
		c := claims[i]
		if isClaimedStatus(c.Status) {
			ov.ClaimedEggCount += c.Quantity
		}
		if ym, ok := monthlyPeriodMonth(c.PeriodKey); ok {
			claimByMonth[ym] = c
		}
	}

	today := time.Now().In(shanghai)
	curYM := today.Format("2006-01")
	startYM := curYM
	if len(effectiveDate) >= 7 {
		startYM = effectiveDate[:7] // 从授权月起列
	}
	ov.Months = buildMonthRecords(startYM, curYM, daysByMonth, claimByMonth)
	// 待领月份数 = 月份列表中 CLAIMABLE 条数，保证顶部汇总与列表逐行一致（问题2）
	for i := range ov.Months {
		if ov.Months[i].Status == MonthStatusClaimable {
			ov.ClaimableMonths++
		}
	}
	return ov, nil
}

// isClaimedStatus 领取记录是否已领取（含发货/签收后续态）
func isClaimedStatus(status string) bool {
	return status == model.ClaimStatusClaimed || status == model.ClaimStatusShipped || status == model.ClaimStatusSigned
}

// monthlyPeriodMonth 从 monthly-YYYY-MM 周期键取月份；非月度周期返回 false
func monthlyPeriodMonth(periodKey string) (string, bool) {
	const prefix = "monthly-"
	if len(periodKey) == len(prefix)+7 && periodKey[:len(prefix)] == prefix {
		return periodKey[len(prefix):], true
	}
	return "", false
}

// buildMonthRecords 从 startYM 到 curYM 逐月组装记录（当月在前，倒序）
func buildMonthRecords(startYM, curYM string, daysByMonth map[string]int, claimByMonth map[string]model.RewardClaim) []MonthRecord {
	start, err := time.Parse("2006-01", startYM)
	if err != nil {
		start, _ = time.Parse("2006-01", curYM)
	}
	cur, _ := time.Parse("2006-01", curYM)

	records := make([]MonthRecord, 0)
	for m := cur; !m.Before(start); m = m.AddDate(0, -1, 0) {
		ym := m.Format("2006-01")
		rec := MonthRecord{
			YearMonth:   ym,
			CheckedDays: daysByMonth[ym],
			TotalDays:   m.AddDate(0, 1, -1).Day(),
			Status:      MonthStatusInProgress,
			IsCurrent:   ym == curYM,
		}
		if claim, ok := claimByMonth[ym]; ok {
			rec.ClaimId = claim.ClaimId
			if claim.Status == model.ClaimStatusPending {
				rec.Status = MonthStatusClaimable
			} else if isClaimedStatus(claim.Status) {
				rec.Status = MonthStatusClaimed
			}
		}
		records = append(records, rec)
	}
	return records
}
