// Package service @Author larry
// @Date 2026/06/15
// @Desc 守护邀请服务（子女发起 → 老人接受 两阶段绑定）

package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/kinds/jsons"
	"warm-nest/pkg/tool/database/mysqls"
	"warm-nest/pkg/utils/rands"

	"warm-nest/internal/constant"
	"warm-nest/internal/feign/wechat"
	wxmodel "warm-nest/internal/feign/wechat/model"
	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
)

// 邀请有效期：7 天（毫秒）
const invitationTTLMs = int64(7 * 24 * 60 * 60 * 1000)

// InvitationService 守护邀请服务
type InvitationService struct {
	invitationMapper   *mapper.InvitationMapper
	guardianshipMapper *mapper.GuardianshipMapper
	elderProfileMapper *mapper.ElderProfileMapper
	userMapper         *mapper.UserMapper
}

var invitationService *InvitationService
var invitationServiceOnce sync.Once

// GetInvitationService 获取守护邀请服务单例
func GetInvitationService() *InvitationService {
	invitationServiceOnce.Do(func() {
		invitationService = &InvitationService{
			invitationMapper:   mapper.GetInvitationMapper(),
			guardianshipMapper: mapper.GetGuardianshipMapper(),
			elderProfileMapper: mapper.GetElderProfileMapper(),
			userMapper:         mapper.GetUserMapper(),
		}
	})
	return invitationService
}

// CreateResult 发起邀请结果
type CreateResult struct {
	InviteCode string `json:"inviteCode"`
	WxaCodeUrl string `json:"wxaCodeUrl"` // 小程序码图片 URL
	ExpireAt   int64  `json:"expireAt"`
}

// CreateInvitationInput 发起邀请入参（参数较多，结构化避免长参数列表）。
// Avatar/Phone 是子女本人资料（问题5）：发起邀请时顺带采集，落入子女 User。
type CreateInvitationInput struct {
	GuardianUserId string // 发起邀请的子女用户ID（登录态）
	ElderPhone     string // 子女填的老人手机号（线索）
	Relation       string // 称呼 MOM/DAD/...
	RemindTime     string // 提醒时间 HH:mm，空则默认 09:00
	City           string // 老人所在城市（线索）
	Avatar         string // 子女头像URL（选填，wx.chooseAvatar 上传后得到）
	Phone          string // 子女手机号明文（选填，前端已用 phoneCode 经 /user/resolve-phone 换好再传）
}

// Create 子女发起邀请，生成小程序码（nowMs 由调用方传入，便于测试）。
// 一对一守卫（问题1 角色互斥）：已作为被守护人(elder)被绑定的用户不能再发起邀请当子女。
// 顺带采集子女资料（问题5）：avatar / phoneCode 非空时更新子女 User 头像/手机号（best-effort，
// 失败不阻断建邀请——邀请是主流程，资料采集是附带增强）。
func (s *InvitationService) Create(in CreateInvitationInput, nowMs int64,
	saveWxaCode func(code []byte) (string, error)) (*CreateResult, error) {

	// 角色互斥：发起者若已被绑为老人 → 拒绝（不能既是老人又是子女）
	boundAsElder, err := s.guardianshipMapper.ListByElder(in.GuardianUserId)
	if err != nil {
		return nil, fmt.Errorf("create invitation check elder role: %w", err)
	}
	if len(boundAsElder) > 0 {
		logrus.WithField("guardianUserId", in.GuardianUserId).Warn("create rejected: user already bound as elder")
		return nil, errors.NewWithArgs(constant.ErrAlreadyElder)
	}

	remindTime := in.RemindTime
	if remindTime == "" {
		remindTime = model.DefaultRemindTime
	}
	inv := &model.Invitation{
		InvitationId:   rands.Numeric(),
		InviteCode:     rands.NumericN(10),
		GuardianUserId: in.GuardianUserId,
		ElderPhone:     in.ElderPhone,
		Relation:       in.Relation,
		RemindTime:     remindTime,
		City:           in.City,
		Status:         model.InvitationStatusPending,
		ExpireAt:       nowMs + invitationTTLMs,
	}
	if err := s.invitationMapper.Create(inv); err != nil {
		return nil, fmt.Errorf("create invitation: %w", err)
	}

	// 采集子女资料（问题5）：落头像 + 手机号明文，best-effort
	s.saveGuardianProfile(in.GuardianUserId, in.Avatar, in.Phone)

	codeImg, err := wechat.Client().GetUnlimitedWxaCode(wxmodel.GetWxaCodeReq{
		Scene:     inv.InviteCode,
		Page:      "pages/invite/accept",
		CheckPath: false,
	})
	if err != nil {
		return nil, fmt.Errorf("create invitation wxacode: %w", err)
	}
	url, err := saveWxaCode(codeImg)
	if err != nil {
		return nil, err
	}
	return &CreateResult{InviteCode: inv.InviteCode, WxaCodeUrl: url, ExpireAt: inv.ExpireAt}, nil
}

// saveGuardianProfile 采集子女头像/手机号写入 User（问题5，best-effort，失败仅告警不阻断邀请）。
// phone 是已换好的明文（前端先调 /user/resolve-phone 用 phoneCode 换得），后端不再调微信换号，直接落。
// avatar/phone 仅在非空时更新，不覆盖已有值为空。⚠️ 手机号敏感，不打进日志。
func (s *InvitationService) saveGuardianProfile(guardianUserId, avatar, phone string) {
	user, err := s.userMapper.GetByUserId(guardianUserId)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("guardianUserId", guardianUserId).Warn("save guardian profile: get user failed")
		return
	}
	changed := false
	if avatar != "" {
		user.Avatar = avatar
		changed = true
	}
	if phone != "" {
		user.Phone = phone
		changed = true
	}
	if changed {
		if err := s.userMapper.Update(user); err != nil {
			logrus.WithError(err).WithField("guardianUserId", guardianUserId).Warn("save guardian profile: update user failed")
		}
	}
}

// Accept 老人接受邀请（老人已登录，elderUserId 来自登录态）。事务内建档案+关系+置已接受。
//
// PRD §8.0.3 守卫：
//   - 一次性——邀请被某老人接受后即失效；他人扫到这张已用过的码（典型：卡片被转发）一律拒绝，
//     不再静默"成功"（原 bug：第二个老人扫到 ACCEPTED 码会 return nil 假成功却不建任何关系）。
//   - 一对一 V1——1 期一老人仅绑一子女、一子女仅绑一老人；任一侧已被占用则拒绝，不静默新增第二条关系。
func (s *InvitationService) Accept(elderUserId, inviteCode string, nowMs int64) error {
	inv, err := s.invitationMapper.GetByInviteCode(inviteCode)
	if err != nil {
		return fmt.Errorf("accept get invitation by code: %w", err)
	}
	if inv == nil {
		return fmt.Errorf("invitation not found: %s", inviteCode)
	}

	switch inv.Status {
	case model.InvitationStatusAccepted:
		// 一次性：只有「同一老人」重复点才幂等成功；其他老人扫到已用码 → 拒绝（防转发误绑）
		if inv.AcceptedElderUserId == elderUserId {
			return nil
		}
		logrus.WithFields(logrus.Fields{"inviteCode": inviteCode, "elderUserId": elderUserId}).
			Warn("accept rejected: invite already used by another elder")
		return errors.NewWithArgs(constant.ErrInviteUsed)
	case model.InvitationStatusCancelled, model.InvitationStatusExpired:
		return errors.NewWithArgs(constant.ErrInviteState)
	}
	if nowMs > inv.ExpireAt {
		return errors.NewWithArgs(constant.ErrInviteState)
	}

	// 防重复绑定：当前 (子女,老人) 对已有 ACTIVE 关系 → 幂等置邀请已接受，不新建关系
	existing, err := s.guardianshipMapper.GetActive(inv.GuardianUserId, elderUserId)
	if err != nil {
		return fmt.Errorf("accept get active guardianship: %w", err)
	}
	if existing != nil {
		inv.Status = model.InvitationStatusAccepted
		inv.AcceptedElderUserId = elderUserId
		inv.AcceptedAt = nowMs
		return s.invitationMapper.Update(inv)
	}

	// 角色互斥（问题1）：接受者若已是子女（发起过有效邀请 或 已绑老人）→ 拒绝，不能既是子女又是老人。
	isGuardian, err := s.isAlreadyGuardian(elderUserId, nowMs)
	if err != nil {
		return err
	}
	if isGuardian {
		logrus.WithField("userId", elderUserId).Warn("accept rejected: user is already a guardian")
		return errors.NewWithArgs(constant.ErrAlreadyGuardian)
	}

	// 一对一守卫 V1：老人已被别的子女绑定 → 拒绝（existing 已排除"被本子女绑"的情形）
	elderBound, err := s.guardianshipMapper.ListByElder(elderUserId)
	if err != nil {
		return fmt.Errorf("accept check elder bound: %w", err)
	}
	if len(elderBound) > 0 {
		logrus.WithField("elderUserId", elderUserId).Warn("accept rejected: elder already bound by another guardian")
		return errors.NewWithArgs(constant.ErrElderBound)
	}
	// 该子女已绑别的老人 → 拒绝
	guardianBound, err := s.guardianshipMapper.ListByGuardian(inv.GuardianUserId)
	if err != nil {
		return fmt.Errorf("accept check guardian bound: %w", err)
	}
	if len(guardianBound) > 0 {
		logrus.WithField("guardianUserId", inv.GuardianUserId).Warn("accept rejected: guardian already bound another elder")
		return errors.NewWithArgs(constant.ErrGuardianBound)
	}

	guardianshipId := rands.Numeric()
	err = mysqls.DB().Transaction(func(tx *gorm.DB) error {
		// 1. 建/补被守护人档案（老人本人补全城市等，这里先落提醒时间默认值）
		profile, err := s.elderProfileMapper.GetByUserId(elderUserId)
		if err != nil {
			return fmt.Errorf("accept get elder profile: %w", err)
		}
		if profile == nil {
			// 落子女预填的线索：提醒时间 + 老人城市(打卡取天气) + 老人电话。老人本人后续可在档案页改。
			if err = tx.Create(&model.ElderProfile{
				UserId:     elderUserId,
				RemindTime: inv.RemindTime,
				City:       inv.City,
				ElderPhone: inv.ElderPhone,
			}).Error; err != nil {
				return fmt.Errorf("accept create elder profile: %w", err)
			}
		}
		// 2. 建守护关系，ActivatedAt 记绑定成立时刻（首月奖励窗口起点）
		if err = tx.Create(&model.Guardianship{
			GuardianshipId: guardianshipId,
			GuardianUserId: inv.GuardianUserId,
			ElderUserId:    elderUserId,
			Relation:       inv.Relation,
			Status:         model.GuardianshipStatusActive,
			ActivatedAt:    nowMs,
		}).Error; err != nil {
			return fmt.Errorf("accept create guardianship: %w", err)
		}
		// 3. 置邀请为已接受（一次性失效）
		inv.Status = model.InvitationStatusAccepted
		inv.AcceptedElderUserId = elderUserId
		inv.AcceptedAt = nowMs
		if err = tx.Save(inv).Error; err != nil {
			return fmt.Errorf("accept update invitation: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 绑定成立后向子女发「✅ 已成功绑定 XX」（PRD §8.0.1.6，best-effort 不回滚绑定）
	// bindTime=接受邀请时刻（模板「绑定时间」）；relation=子女对老人称呼（模板「绑定用户」）
	params := jsons.JSONObject{
		"relation": inv.Relation,
		"bindTime": time.UnixMilli(nowMs).In(shanghai).Format(dateTimeLayout),
	}
	GetMessageService().SendBindSuccess(inv.GuardianUserId, elderUserId, guardianshipId, params)
	return nil
}

// ListByGuardian 子女查自己发起的邀请
func (s *InvitationService) ListByGuardian(guardianUserId string) ([]model.Invitation, error) {
	return s.invitationMapper.ListByGuardian(guardianUserId)
}

// isAlreadyGuardian 判定用户是否已是子女（角色互斥用，问题1）。
// 判据「填了守护人信息就锁定为子女」：① 已有 ACTIVE 守护关系，或 ② 发起过仍有效的邀请
// （PENDING 未过期 / 已 ACCEPTED）。EXPIRED/CANCELLED 的历史邀请不锁定（已失效，不算"当前是子女"）。
func (s *InvitationService) isAlreadyGuardian(userId string, nowMs int64) (bool, error) {
	bound, err := s.guardianshipMapper.ListByGuardian(userId)
	if err != nil {
		return false, fmt.Errorf("check guardian role list guardianship: %w", err)
	}
	if len(bound) > 0 {
		return true, nil
	}
	return s.HasActiveInvitationAsGuardian(userId, nowMs)
}

// HasActiveInvitationAsGuardian 判定用户是否发起过「仍有效」的邀请，从而对外暴露为子女身份（问题1）。
// 仅看邀请侧（PENDING 未过期 / 已 ACCEPTED），不含守护边——守护边由调用方（登录/绑定状态）自行判定。
// 用途：login.availableRoles 与 bind-status 据此把「发起了 PENDING 邀请、老人尚未接受」的子女
// 暴露为 GUARDIAN 身份，使其能进入「等待老人接受」页、不被未绑定拦截。
// 与 isAlreadyGuardian 的邀请判据同源，保证「能发起 = 被认作子女」一致。
func (s *InvitationService) HasActiveInvitationAsGuardian(userId string, nowMs int64) (bool, error) {
	invitations, err := s.invitationMapper.ListByGuardian(userId)
	if err != nil {
		return false, fmt.Errorf("check active invitation as guardian list invitations: %w", err)
	}
	for i := range invitations {
		inv := invitations[i]
		if inv.Status == model.InvitationStatusAccepted {
			return true, nil
		}
		if inv.Status == model.InvitationStatusPending && nowMs <= inv.ExpireAt {
			return true, nil
		}
	}
	return false, nil
}

// ListPendingAsGuardian 列出用户作为子女发起、仍待接受且未过期的 PENDING 邀请（需求1）。
// 用途：family/view 在 ACTIVE 守护边之外，追加展示「已发起、老人尚未接受」的邀请，使子女创建邀请后
// 首页即看到「等待老人接受」的关系而非空列表（避免被误判为未填写、又引导重填）。
// 仅取 PENDING 未过期——ACCEPTED 的已落为 ACTIVE 守护边由 ViewFamily 正常返回，不在此重复；
// EXPIRED/CANCELLED 已失效不展示。与 HasActiveInvitationAsGuardian 的邀请判据同源（PENDING 未过期）。
func (s *InvitationService) ListPendingAsGuardian(userId string, nowMs int64) ([]model.Invitation, error) {
	invitations, err := s.invitationMapper.ListByGuardian(userId)
	if err != nil {
		return nil, fmt.Errorf("list pending as guardian: %w", err)
	}
	pending := make([]model.Invitation, 0, len(invitations))
	for i := range invitations {
		inv := invitations[i]
		if inv.Status == model.InvitationStatusPending && nowMs <= inv.ExpireAt {
			pending = append(pending, inv)
		}
	}
	return pending, nil
}

// InviterInfoResult 凭邀请码查到的邀请人信息（问题4，老人接受页展示「是否接受 XX 的邀请」）。
// 手机号脱敏返回，不暴露完整号码给任意登录用户。
type InviterInfoResult struct {
	GuardianUserId string `json:"guardianUserId"` // 邀请人（子女）用户ID
	GuardianAvatar string `json:"guardianAvatar"` // 邀请人头像（问题5 采集后才有，未采集为空）
	GuardianName   string `json:"guardianName"`   // 邀请人微信昵称（未采集为空）
	GuardianPhone  string `json:"guardianPhone"`  // 邀请人手机号（脱敏，如 138****8000）
	Relation       string `json:"relation"`       // 子女对老人的预设称呼（MOM/DAD/...）
	Status         string `json:"status"`         // 邀请状态（PENDING/ACCEPTED/...）
	ExpireAt       int64  `json:"expireAt"`       // 邀请过期时间，毫秒
}

// InviterInfo 凭邀请码查邀请人信息（问题4）：供老人扫码后的接受页展示「是否接受 XX 的邀请」。
// 任意登录用户凭码可查（inviteCode 本身不可枚举，登录态再加一道门）；邀请不存在/已失效返 ErrInviteState。
// 手机号脱敏返回。
func (s *InvitationService) InviterInfo(inviteCode string, nowMs int64) (*InviterInfoResult, error) {
	inv, err := s.invitationMapper.GetByInviteCode(inviteCode)
	if err != nil {
		return nil, fmt.Errorf("inviter info get invitation by code: %w", err)
	}
	if inv == nil {
		return nil, errors.NewWithArgs(constant.ErrInviteState)
	}
	// 已撤销/已过期，或 PENDING 但已超时 → 视为失效，不展示邀请人信息
	if inv.Status == model.InvitationStatusCancelled || inv.Status == model.InvitationStatusExpired ||
		(inv.Status == model.InvitationStatusPending && nowMs > inv.ExpireAt) {
		return nil, errors.NewWithArgs(constant.ErrInviteState)
	}

	res := &InviterInfoResult{
		GuardianUserId: inv.GuardianUserId,
		Relation:       inv.Relation,
		Status:         inv.Status,
		ExpireAt:       inv.ExpireAt,
	}
	user, err := s.userMapper.GetByUserId(inv.GuardianUserId)
	if err != nil {
		return nil, fmt.Errorf("inviter info get guardian user %s: %w", inv.GuardianUserId, err)
	}
	if user != nil {
		res.GuardianAvatar = user.Avatar
		res.GuardianName = user.Nickname
		res.GuardianPhone = maskPhone(user.Phone)
	}
	return res, nil
}

// maskPhone 手机号脱敏：保留前3后4，中间 ****（如 13800008000 → 138****8000）。
// 非 11 位（异常/空）原样返回前最多保留首尾或返回空，避免 panic。
func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// Cancel 子女撤销自己发起的邀请（问题3：仅 PENDING 可撤销 → CANCELLED）。
// 校验：邀请存在 + 归属本人 + 当前为 PENDING；已接受/已过期/已撤销均不可撤销。
func (s *InvitationService) Cancel(guardianUserId, invitationId string) error {
	inv, err := s.invitationMapper.GetByInvitationId(invitationId)
	if err != nil {
		return fmt.Errorf("cancel get invitation: %w", err)
	}
	if inv == nil {
		return errors.NewWithArgs(constant.ErrInviteState)
	}
	if inv.GuardianUserId != guardianUserId {
		logrus.WithFields(logrus.Fields{"invitationId": invitationId, "guardianUserId": guardianUserId}).
			Warn("cancel rejected: not invitation owner")
		return errors.NewWithArgs(constant.ErrInviteNotOwner)
	}
	if inv.Status != model.InvitationStatusPending {
		return errors.NewWithArgs(constant.ErrInviteNotCancelable)
	}
	inv.Status = model.InvitationStatusCancelled
	return s.invitationMapper.Update(inv)
}

// ExpireOverdue 将已过期但仍 PENDING 的邀请置为 EXPIRED（定时任务调用）
func (s *InvitationService) ExpireOverdue(nowMs int64) (int, error) {
	list, err := s.invitationMapper.ListExpiredPending(nowMs)
	if err != nil {
		return 0, fmt.Errorf("expire overdue list expired pending: %w", err)
	}
	for i := range list {
		inv := list[i]
		inv.Status = model.InvitationStatusExpired
		if err = s.invitationMapper.Update(&inv); err != nil {
			return 0, fmt.Errorf("expire invitation %s: %w", inv.InvitationId, err)
		}
	}
	return len(list), nil
}
