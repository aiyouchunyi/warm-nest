// Package service @Author larry
// @Date 2026/06/19
// @Desc 测试数据服务（仅非 prod 环境可用，供前端自助造数据测试打卡/领取链路）
//
// 为什么需要：打卡连续天数、奖励达成、领取等链路依赖「历史打卡记录 + 奖励任务规则」，
// 线上要等真实老人连打很多天才能复现，前端无法自测领取。故提供测试专用造数据能力：
//   - 指定 userId 从某日起回填连续 N 天打卡（绕过一日一卡的当日限制，可造过去日期）
//   - 初始化奖励任务规则种子（RewardTask 表无种子则 EvaluateRewards 永远评估不出可领取）
//
// 安全边界：本服务所有方法只在非 prod 环境暴露（controller 层用 configs.IsProd 拦截）。
package service

import (
	"fmt"
	"sync"
	"time"

	"warm-nest/pkg/tool/database/mysqls"
	"warm-nest/pkg/utils/rands"

	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
	"warm-nest/internal/preset"
)

// TestService 测试数据服务
type TestService struct {
	checkInMapper         *mapper.CheckInMapper
	rewardTaskMapper      *mapper.RewardTaskMapper
	rewardClaimMapper     *mapper.RewardClaimMapper
	userMapper            *mapper.UserMapper
	elderProfileMapper    *mapper.ElderProfileMapper
	guardianshipMapper    *mapper.GuardianshipMapper
	shippingAddressMapper *mapper.ShippingAddressMapper
	invitationMapper      *mapper.InvitationMapper
}

var testService *TestService
var testServiceOnce sync.Once

// GetTestService 获取测试数据服务单例
func GetTestService() *TestService {
	testServiceOnce.Do(func() {
		testService = &TestService{
			checkInMapper:         mapper.GetCheckInMapper(),
			rewardTaskMapper:      mapper.GetRewardTaskMapper(),
			rewardClaimMapper:     mapper.GetRewardClaimMapper(),
			userMapper:            mapper.GetUserMapper(),
			elderProfileMapper:    mapper.GetElderProfileMapper(),
			guardianshipMapper:    mapper.GetGuardianshipMapper(),
			shippingAddressMapper: mapper.GetShippingAddressMapper(),
			invitationMapper:      mapper.GetInvitationMapper(),
		}
	})
	return testService
}

// SeedGuardianshipResult 造守护关系结果
type SeedGuardianshipResult struct {
	GuardianUserId string `json:"guardianUserId"` // 子女
	ElderUserId    string `json:"elderUserId"`    // 老人
	GuardianshipId string `json:"guardianshipId"` // 关系边ID
	Relation       string `json:"relation"`       // 称呼
}

// SeedGuardianship 造「子女→老人」ACTIVE 守护关系 + 双方 User + 老人档案（带结构化地址），幂等。
// 供前端测试公用接口（子女传 elderUserId 看老人数据）。guardianUserId/elderUserId 自定，relation 默认 DAD。
func (s *TestService) SeedGuardianship(guardianUserId, elderUserId, relation string) (*SeedGuardianshipResult, error) {
	if guardianUserId == "" || elderUserId == "" {
		return nil, fmt.Errorf("seed guardianship: guardianUserId/elderUserId required")
	}
	if relation == "" {
		relation = model.RelationDad
	}
	// 幂等建双方 User
	for _, uid := range []string{guardianUserId, elderUserId} {
		u, err := s.userMapper.GetByUserId(uid)
		if err != nil {
			return nil, fmt.Errorf("seed guardianship get user %s: %w", uid, err)
		}
		if u == nil {
			nu := &model.User{UserId: uid, OpenId: "openid-" + uid, Nickname: uid, Status: model.UserStatusNormal}
			if err = s.userMapper.Create(nu); err != nil {
				return nil, fmt.Errorf("seed guardianship create user %s: %w", uid, err)
			}
		}
	}
	// 幂等建老人档案（提醒时间/城市等；收货地址已迁到地址簿，见下方单独造）
	profile, err := s.elderProfileMapper.GetByUserId(elderUserId)
	if err != nil {
		return nil, fmt.Errorf("seed guardianship get profile %s: %w", elderUserId, err)
	}
	if profile == nil {
		np := &model.ElderProfile{
			UserId:     elderUserId,
			RealName:   "测试老人",
			City:       "上海",
			RemindTime: model.DefaultRemindTime,
		}
		if err = s.elderProfileMapper.Create(np); err != nil {
			return nil, fmt.Errorf("seed guardianship create profile %s: %w", elderUserId, err)
		}
	}
	// 幂等建默认收货地址（地址簿，问题3）：领取下单从地址簿读，无地址簿则领取直接报缺地址。
	addrs, err := s.shippingAddressMapper.ListByElder(elderUserId)
	if err != nil {
		return nil, fmt.Errorf("seed guardianship list address %s: %w", elderUserId, err)
	}
	if len(addrs) == 0 {
		def := &model.ShippingAddress{
			AddressId:   rands.Numeric(),
			ElderUserId: elderUserId,
			Address: model.Address{
				Province: "上海市", City: "上海市", District: "浦东新区",
				Street: "陆家嘴街道", Detail: "测试路1号101室",
				ReceiverName: "测试收货人", ReceiverPhone: "13800138000",
			},
			IsDefault: true,
			SortNo:    0,
		}
		if err = s.shippingAddressMapper.Create(def); err != nil {
			return nil, fmt.Errorf("seed guardianship create address %s: %w", elderUserId, err)
		}
	}
	// 幂等建 ACTIVE 守护关系
	g, err := s.guardianshipMapper.GetActive(guardianUserId, elderUserId)
	if err != nil {
		return nil, fmt.Errorf("seed guardianship get active: %w", err)
	}
	if g == nil {
		g = &model.Guardianship{
			GuardianshipId: rands.Numeric(),
			GuardianUserId: guardianUserId,
			ElderUserId:    elderUserId,
			Relation:       relation,
			Status:         model.GuardianshipStatusActive,
			ActivatedAt:    time.Now().UnixMilli(),
		}
		if err = s.guardianshipMapper.Create(g); err != nil {
			return nil, fmt.Errorf("seed guardianship create relation: %w", err)
		}
	}
	return &SeedGuardianshipResult{
		GuardianUserId: guardianUserId, ElderUserId: elderUserId,
		GuardianshipId: g.GuardianshipId, Relation: g.Relation,
	}, nil
}

// SeedPendingInvitationResult 造待接受邀请结果
type SeedPendingInvitationResult struct {
	GuardianUserId string `json:"guardianUserId"` // 子女（发起人）
	InvitationId   string `json:"invitationId"`   // 邀请ID
	InviteCode     string `json:"inviteCode"`     // 邀请口令（凭此测 inviter-info / accept）
	ExpireAt       int64  `json:"expireAt"`       // 过期时间，毫秒
}

// SeedPendingInvitation 造一条 PENDING 待接受邀请 + 子女 User（带头像/昵称/手机号，便于测问题4/5）。
// 供前端测：① 问题1 子女发起后 login.availableRoles 含 GUARDIAN、bind-status.Bound=true；
// ② 问题4 凭 inviteCode 查邀请人信息；③ 问题5 子女 User 头像/手机号已落。幂等：同子女已有 PENDING 则复用。
func (s *TestService) SeedPendingInvitation(guardianUserId, relation string) (*SeedPendingInvitationResult, error) {
	if guardianUserId == "" {
		return nil, fmt.Errorf("seed pending invitation: guardianUserId required")
	}
	if relation == "" {
		relation = model.RelationDad
	}
	// 幂等建子女 User（带头像/昵称/手机号，模拟问题5 已采集）
	u, err := s.userMapper.GetByUserId(guardianUserId)
	if err != nil {
		return nil, fmt.Errorf("seed pending invitation get user %s: %w", guardianUserId, err)
	}
	if u == nil {
		nu := &model.User{
			UserId: guardianUserId, OpenId: "openid-" + guardianUserId, Nickname: "测试子女",
			Avatar: "https://via.placeholder.com/120x120.png?text=avatar", Phone: "13800138000",
			Status: model.UserStatusNormal,
		}
		if err = s.userMapper.Create(nu); err != nil {
			return nil, fmt.Errorf("seed pending invitation create user %s: %w", guardianUserId, err)
		}
	}
	// 幂等复用已有有效 PENDING
	nowMs := time.Now().UnixMilli()
	existing, err := s.invitationMapper.ListByGuardian(guardianUserId)
	if err != nil {
		return nil, fmt.Errorf("seed pending invitation list %s: %w", guardianUserId, err)
	}
	for i := range existing {
		inv := existing[i]
		if inv.Status == model.InvitationStatusPending && nowMs <= inv.ExpireAt {
			return &SeedPendingInvitationResult{
				GuardianUserId: guardianUserId, InvitationId: inv.InvitationId,
				InviteCode: inv.InviteCode, ExpireAt: inv.ExpireAt,
			}, nil
		}
	}
	inv := &model.Invitation{
		InvitationId:   rands.Numeric(),
		InviteCode:     rands.NumericN(10),
		GuardianUserId: guardianUserId,
		ElderPhone:     "13900139000",
		Relation:       relation,
		RemindTime:     model.DefaultRemindTime,
		City:           "上海",
		Status:         model.InvitationStatusPending,
		ExpireAt:       nowMs + invitationTTLMs,
	}
	if err = s.invitationMapper.Create(inv); err != nil {
		return nil, fmt.Errorf("seed pending invitation create %s: %w", guardianUserId, err)
	}
	return &SeedPendingInvitationResult{
		GuardianUserId: guardianUserId, InvitationId: inv.InvitationId,
		InviteCode: inv.InviteCode, ExpireAt: inv.ExpireAt,
	}, nil
}

// SeedCheckInsResult 造打卡结果
type SeedCheckInsResult struct {
	ElderUserId string   `json:"elderUserId"`
	Created     int      `json:"created"` // 实际新建条数（已存在的当日跳过）
	Skipped     int      `json:"skipped"` // 已存在跳过条数
	Dates       []string `json:"dates"`   // 造的日期列表
}

// SeedContinuousCheckIns 给 elderUserId 造「截至 endDate 往前连续 days 天」的打卡记录（幂等，已存在跳过）。
// endDate 空则默认今天（Asia/Shanghai）。造完调一次奖励评估，方便直接测领取。
func (s *TestService) SeedContinuousCheckIns(elderUserId, endDate string, days int) (*SeedCheckInsResult, error) {
	if elderUserId == "" || days <= 0 {
		return nil, fmt.Errorf("seed checkins: invalid args elderUserId=%s days=%d", elderUserId, days)
	}
	end := time.Now().In(shanghai)
	if endDate != "" {
		t, err := time.ParseInLocation(dateLayout, endDate, shanghai)
		if err != nil {
			return nil, fmt.Errorf("seed checkins parse endDate %s: %w", endDate, err)
		}
		end = t
	}

	res := &SeedCheckInsResult{ElderUserId: elderUserId, Dates: make([]string, 0, days)}
	for i := 0; i < days; i++ {
		date := end.AddDate(0, 0, -i).Format(dateLayout)
		existing, err := s.checkInMapper.GetByUserDate(elderUserId, date)
		if err != nil {
			return nil, fmt.Errorf("seed checkins query %s@%s: %w", elderUserId, date, err)
		}
		if existing != nil {
			res.Skipped++
			continue
		}
		checkIn := &model.CheckIn{
			CheckInId:   rands.Numeric(),
			ElderUserId: elderUserId,
			CheckInDate: date,
			Kind:        model.CheckInKindNormal,
			PhotoUrl:    "https://via.placeholder.com/300x300.png?text=checkin", // 占位照片，便于测消息列表带图（问题2）
			Weather:     "测试数据",
			City:        "测试城市",
		}
		if err = s.checkInMapper.Create(checkIn); err != nil {
			return nil, fmt.Errorf("seed checkins create %s@%s: %w", elderUserId, date, err)
		}
		res.Created++
		res.Dates = append(res.Dates, date)
	}

	// 造完触发一次奖励评估（best-effort 不阻断造数据结果返回）
	_ = GetRewardService().EvaluateRewards(elderUserId)
	return res, nil
}

// SeedRewardTasksResult 初始化奖励规则结果
type SeedRewardTasksResult struct {
	Created []string `json:"created"` // 新建的 taskKey
	Existed []string `json:"existed"` // 已存在跳过的 taskKey
}

// 预设奖励规则已搬到 internal/preset.RewardTasks()（migrate 幂等 seed 的唯一真相源）；
// 本测试接口复用同一份数据手动重灌，不再各自维护一份。

// wipeModels 全清覆盖的全部表。仅运行时业务数据；预设数据表（reward_task/notify_route）不在此列，
// 由 preset+migrate 维护，WipeAll 不清（避免清库后出厂规则丢失需手工重灌）。
func wipeModels() []struct {
	name  string
	model any
} {
	return []struct {
		name  string
		model any
	}{
		{"checkIn", &model.CheckIn{}},
		{"message", &model.Message{}},
		{"rewardClaim", &model.RewardClaim{}},
		// 注：rewardTask 是预设数据，不清（preset+migrate 维护，见 preset 包说明）
		{"guardianship", &model.Guardianship{}},
		{"elderProfile", &model.ElderProfile{}},
		{"shippingAddress", &model.ShippingAddress{}},
		{"invitation", &model.Invitation{}},
		{"fan", &model.Fan{}},
		{"user", &model.User{}},
	}
}

// WipeAll 清空全部业务表 + 账号表（问题10，⚠️不可逆，仅测试通道用）。
// 用 Unscoped 真删（绕过软删 DeletedAt），逐表返回清理条数。清后需重新微信登录+邀请绑定。
func (s *TestService) WipeAll() (map[string]int64, error) {
	result := make(map[string]int64)
	for _, t := range wipeModels() {
		res := mysqls.DB().Unscoped().Where("1 = 1").Delete(t.model)
		if res.Error != nil {
			return result, fmt.Errorf("wipe table %s: %w", t.name, res.Error)
		}
		result[t.name] = res.RowsAffected
	}
	return result, nil
}

// SeedRewardClaimsResult 造领取记录结果
type SeedRewardClaimsResult struct {
	ElderUserId string   `json:"elderUserId"` // 归属老人
	Status      string   `json:"status"`      // 造的状态
	Created     int      `json:"created"`     // 实际新建条数
	ClaimIds    []string `json:"claimIds"`    // 新建记录的业务领取ID列表
}

// SeedRewardClaims 直插 N 条领取记录到指定老人名下，可任意指定状态/数量/物流，供前端测列表多状态展示。
//
// 与 SeedContinuousCheckIns 的区别：后者走「打卡→评估」真实路径、只能产出 PENDING；本方法绕过评估直插，
// 可造 CLAIMED/SHIPPED/SIGNED 等任意态及物流字段，专用于测列表/详情的状态与物流展示。
// task_key/period_key 每条随机唯一（避开 u_user_task_period 唯一索引）；按 status 合理回填各状态时间戳与物流。
func (s *TestService) SeedRewardClaims(elderUserId, status string, count, quantity int) (*SeedRewardClaimsResult, error) {
	if elderUserId == "" {
		return nil, fmt.Errorf("seed reward claims: elderUserId required")
	}
	if count <= 0 {
		count = 1
	}
	if quantity <= 0 {
		quantity = 1
	}
	if status == "" {
		status = model.ClaimStatusPending
	}
	switch status {
	case model.ClaimStatusPending, model.ClaimStatusClaimed, model.ClaimStatusShipped, model.ClaimStatusSigned:
	default:
		return nil, fmt.Errorf("seed reward claims: invalid status %s (PENDING/CLAIMED/SHIPPED/SIGNED)", status)
	}

	nowMs := time.Now().UnixMilli()
	res := &SeedRewardClaimsResult{ElderUserId: elderUserId, Status: status, ClaimIds: make([]string, 0, count)}
	for i := 0; i < count; i++ {
		claimId := rands.Numeric()
		claim := &model.RewardClaim{
			ClaimId:     claimId,
			ElderUserId: elderUserId,
			TaskKey:     "seed_" + claimId,        // 随机唯一，避开 u_user_task_period 唯一索引
			PeriodKey:   "seed-period-" + claimId, // 同上
			RewardKind:  "EGG",
			RewardName:  "安心鸡蛋",
			RewardSpec:  "30枚/盒",
			Quantity:    quantity,
		}
		claim.Status = status
		// 收货信息快照（CLAIMED 及以后才有意义；统一造一份便于详情展示）
		claim.ReceiverAddress = model.Address{
			Province: "上海市", City: "上海市", District: "浦东新区",
			Street: "陆家嘴街道", Detail: "测试路1号101室",
			ReceiverName: "测试收货人", ReceiverPhone: "13800138000",
		}
		// 按状态递进回填各节点时间戳与物流（SHIPPED/SIGNED 带快递单号）
		if status == model.ClaimStatusClaimed || status == model.ClaimStatusShipped || status == model.ClaimStatusSigned {
			claim.ClaimedAt = nowMs
		}
		if status == model.ClaimStatusShipped || status == model.ClaimStatusSigned {
			claim.ShippedAt = nowMs
			claim.ExpressCompany = "顺丰速运"
			claim.ExpressNo = "SF" + claimId
		}
		if status == model.ClaimStatusSigned {
			claim.SignedAt = nowMs
		}
		if err := s.rewardClaimMapper.Create(claim); err != nil {
			return nil, fmt.Errorf("seed reward claims create %s: %w", claimId, err)
		}
		res.Created++
		res.ClaimIds = append(res.ClaimIds, claimId)
	}
	return res, nil
}

// SeedRewardTasks 初始化默认奖励任务规则（幂等：按 taskKey 已存在则跳过）
func (s *TestService) SeedRewardTasks() (*SeedRewardTasksResult, error) {
	res := &SeedRewardTasksResult{Created: make([]string, 0), Existed: make([]string, 0)}
	for _, task := range preset.RewardTasks() {
		existing, err := s.rewardTaskMapper.GetByTaskKey(task.TaskKey)
		if err != nil {
			return nil, fmt.Errorf("seed reward tasks query %s: %w", task.TaskKey, err)
		}
		if existing != nil {
			res.Existed = append(res.Existed, task.TaskKey)
			continue
		}
		t := task
		if err = s.rewardTaskMapper.Create(&t); err != nil {
			return nil, fmt.Errorf("seed reward tasks create %s: %w", task.TaskKey, err)
		}
		res.Created = append(res.Created, task.TaskKey)
	}
	return res, nil
}

// TriggerRemindPollingResult 手动触发提醒轮询结果
type TriggerRemindPollingResult struct {
	TriggeredAt string `json:"triggeredAt"` // 实际用于触发的时刻（YYYY-MM-DD HH:mm）
}

// TriggerRemindPolling 手动触发一次未打卡提醒轮询（测试用）：
// at 为空 → 用真实当前时刻；at=HH:mm → 模拟当天该时刻触发（便于精确命中某老人的第一段/第二段窗口）。
func (s *TestService) TriggerRemindPolling(at string) (*TriggerRemindPollingResult, error) {
	now := time.Now().In(shanghai)
	if at != "" {
		hm, err := time.Parse("15:04", at)
		if err != nil {
			return nil, fmt.Errorf("trigger remind polling parse at %s: %w", at, err)
		}
		y, m, d := now.Date()
		now = time.Date(y, m, d, hm.Hour(), hm.Minute(), 0, 0, shanghai)
	}
	if err := GetCheckInService().RemindByPolling(now); err != nil {
		return nil, fmt.Errorf("trigger remind polling: %w", err)
	}
	return &TriggerRemindPollingResult{TriggeredAt: now.Format(dateTimeLayout)}, nil
}
