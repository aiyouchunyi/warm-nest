// Package service @Author larry
// @Date 2026/06/15
// @Desc 家庭信息服务（双端同源，viewerRole 由登录用户在守护关系边的位置决定）

package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/errors"

	wnconst "warm-nest/internal/constant"
	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
)

// FamilyService 家庭信息服务
type FamilyService struct {
	guardianshipMapper    *mapper.GuardianshipMapper
	elderProfileMapper    *mapper.ElderProfileMapper
	shippingAddressMapper *mapper.ShippingAddressMapper
	userMapper            *mapper.UserMapper
}

var familyService *FamilyService
var familyServiceOnce sync.Once

// GetFamilyService 获取家庭信息服务单例
func GetFamilyService() *FamilyService {
	familyServiceOnce.Do(func() {
		familyService = &FamilyService{
			guardianshipMapper:    mapper.GetGuardianshipMapper(),
			elderProfileMapper:    mapper.GetElderProfileMapper(),
			shippingAddressMapper: mapper.GetShippingAddressMapper(),
			userMapper:            mapper.GetUserMapper(),
		}
	})
	return familyService
}

// FamilyRelation 一条守护关系的双端中性视图（view 与 profile 读同源同结构）。
//
// 字段对齐说明（PRD 家庭信息，问题 5/6/7）：老人信息与子女信息一并返回，前端按 viewerRole 决定展示哪侧。
//   - 老人侧：elderUserId / elderRealName / elderCity / elderPhone / elderHealthNote / elderAddress
//   - 子女侧：guardianUserId / guardianPhone
//   - 关系：relation（子女对老人的称呼，挂关系边）/ remindTime（老人级，挂档案）
//
// elderPhone/guardianPhone 来自 ElderProfile（非微信登录手机号，见档案模型说明）。
type FamilyRelation struct {
	GuardianshipId string `json:"guardianshipId"`
	// Status 关系状态：ACTIVE（守护中，老人已接受、已建守护边）/ PENDING（待接受，子女已发起邀请、老人尚未接受，需求1）。
	// PENDING 条无 guardianshipId/elderUserId（老人未登录、关系未建立），仅回填子女填的线索 + 邀请口令供等待页展示。
	Status     string `json:"status"`
	ViewerRole string `json:"viewerRole"` // ELDER（被守护，看护我）/ GUARDIAN（守护，我看护）
	Relation   string `json:"relation"`   // 子女对老人的称呼

	// —— 老人侧信息 ——
	ElderUserId     string        `json:"elderUserId"`     // 老人用户ID
	ElderRealName   string        `json:"elderRealName"`   // 老人真名（来自档案）
	ElderCity       string        `json:"elderCity"`       // 老人城市（天气/定位）
	ElderPhone      string        `json:"elderPhone"`      // 老人本人电话（档案可编辑，脱敏由前端处理）
	ElderHealthNote string        `json:"elderHealthNote"` // 老人健康备注（敏感，供编辑页回填）
	ElderAddress    model.Address `json:"elderAddress"`    // 老人默认收货地址（结构化对象，供编辑页回填）

	// —— 子女侧信息 ——
	GuardianUserId string `json:"guardianUserId"` // 子女用户ID
	GuardianPhone  string `json:"guardianPhone"`  // 守护人电话（老人突发情况时通知，档案可编辑）

	// 子女账号资料（问题2）：来自子女 User 账号（非档案），供老人端展示「我的子女」头像/昵称/手机号。
	// 手机号脱敏返回，与 invitation/inviter-info 一致，不向家庭成员暴露完整号。
	GuardianAvatar       string `json:"guardianAvatar"`       // 子女头像（微信账号，问题5 采集后才有）
	GuardianName         string `json:"guardianName"`         // 子女微信昵称
	GuardianAccountPhone string `json:"guardianAccountPhone"` // 子女微信手机号（脱敏 138****8000，区别于档案里的 guardianPhone）

	RemindTime string `json:"remindTime"` // 提醒时间（老人级）

	// TaUserId 对方用户ID（兼容旧前端：老人端=子女ID，子女端=老人ID），新前端建议用上面分侧字段
	TaUserId string `json:"taUserId"`

	// —— PENDING 待接受邀请专用（需求1，仅 Status=PENDING 时有值）——
	// ACTIVE 关系这两字段留零值。
	InviteCode string `json:"inviteCode"` // 邀请口令（供等待页展示/再次分享小程序码）
	ExpireAt   int64  `json:"expireAt"`   // 邀请过期时间，毫秒
}

// ViewFamily 按登录用户算其参与的所有 ACTIVE 关系的双端视图（1 期一对一通常 1 条）
func (s *FamilyService) ViewFamily(viewerUserId string) ([]FamilyRelation, error) {
	result := make([]FamilyRelation, 0)

	// 作为老人被守护的关系（viewer == elder → 老人端视角）
	asElder, err := s.guardianshipMapper.ListByElder(viewerUserId)
	if err != nil {
		return nil, fmt.Errorf("view family list by elder: %w", err)
	}
	for i := range asElder {
		rel, err := s.buildRelation(asElder[i], model.RoleElder)
		if err != nil {
			return nil, err
		}
		result = append(result, rel)
	}

	// 作为子女守护的关系（viewer == guardian → 子女端视角）
	asGuardian, err := s.guardianshipMapper.ListByGuardian(viewerUserId)
	if err != nil {
		return nil, fmt.Errorf("view family list by guardian: %w", err)
	}
	for i := range asGuardian {
		rel, err := s.buildRelation(asGuardian[i], model.RoleGuardian)
		if err != nil {
			return nil, err
		}
		result = append(result, rel)
	}

	// 追加「已发起、老人尚未接受」的 PENDING 邀请（需求1）：子女创建邀请后、老人 accept 前还没有守护边，
	// 若只返 ACTIVE 关系，view 为空会被前端误判为未填写、又引导重填。这里把有效 PENDING 邀请也作为
	// 一条 GUARDIAN 视角的「待接受」关系返回，供前端展示「等待老人接受」而非空列表。
	pending, err := GetInvitationService().ListPendingAsGuardian(viewerUserId, time.Now().UnixMilli())
	if err != nil {
		return nil, fmt.Errorf("view family list pending invitations: %w", err)
	}
	for i := range pending {
		result = append(result, s.buildPendingRelation(pending[i]))
	}
	return result, nil
}

// buildPendingRelation 把一条 PENDING 邀请组装成 GUARDIAN 视角的「待接受」关系（需求1）。
// 老人尚未登录/接受：无 guardianshipId、无 elderUserId、无老人档案；仅回填子女填的线索（relation/
// remindTime/elderPhone/elderCity）+ 子女本人账号资料 + 邀请口令，供前端「等待老人接受」页展示。
func (s *FamilyService) buildPendingRelation(inv model.Invitation) FamilyRelation {
	rel := FamilyRelation{
		Status:         model.InvitationStatusPending,
		ViewerRole:     model.RoleGuardian,
		Relation:       inv.Relation,
		GuardianUserId: inv.GuardianUserId,
		ElderPhone:     inv.ElderPhone, // 子女填的老人手机号线索
		ElderCity:      inv.City,       // 子女填的老人城市线索
		RemindTime:     inv.RemindTime,
		InviteCode:     inv.InviteCode,
		ExpireAt:       inv.ExpireAt,
	}
	// 子女账号资料（与 buildRelation 一致）：头像/昵称/脱敏手机号，供等待页展示「我（子女）」一侧。
	if gu, err := s.userMapper.GetByUserId(inv.GuardianUserId); err != nil {
		logrus.WithError(err).WithField("guardianUserId", inv.GuardianUserId).Warn("build pending relation get guardian user failed")
	} else if gu != nil {
		rel.GuardianAvatar = gu.Avatar
		rel.GuardianName = gu.Nickname
		rel.GuardianAccountPhone = maskPhone(gu.Phone)
	}
	return rel
}

// BindStatus 当前用户绑定状态（PRD §8.0.3：老人未经邀请直接打开小程序需拦截）
type BindStatus struct {
	Bound      bool `json:"bound"`      // 是否已建立任一 ACTIVE 守护关系
	AsElder    bool `json:"asElder"`    // 是否作为被守护人（老人）已被绑定
	AsGuardian bool `json:"asGuardian"` // 是否作为守护人（子女）已绑定老人
}

// GetBindStatus 查当前用户绑定态：供前端对「未经邀请直接进入的老人」拦截到等待页（§8.0.3）。
// 纯未绑定用户 Bound=false；前端据此拦截（老人侧显示「等待孩子邀请」，子女侧引导去邀请）。
// AsGuardian 除守护边外，还认「发起了有效 PENDING 邀请」（问题1）：子女发起邀请、老人尚未接受时
// 即视为已有子女身份、Bound=true，使其能进入「等待老人接受」页而不被未绑定拦截。
func (s *FamilyService) GetBindStatus(viewerUserId string) (*BindStatus, error) {
	asElder, err := s.guardianshipMapper.ListByElder(viewerUserId)
	if err != nil {
		return nil, fmt.Errorf("bind status list by elder: %w", err)
	}
	asGuardian, err := s.guardianshipMapper.ListByGuardian(viewerUserId)
	if err != nil {
		return nil, fmt.Errorf("bind status list by guardian: %w", err)
	}
	isGuardian := len(asGuardian) > 0
	if !isGuardian {
		hasInvite, err := GetInvitationService().HasActiveInvitationAsGuardian(viewerUserId, time.Now().UnixMilli())
		if err != nil {
			return nil, fmt.Errorf("bind status check active invitation: %w", err)
		}
		isGuardian = hasInvite
	}
	st := &BindStatus{
		AsElder:    len(asElder) > 0,
		AsGuardian: isGuardian,
	}
	st.Bound = st.AsElder || st.AsGuardian
	return st, nil
}

// buildRelation 组装一条关系视图（老人+子女信息同时返回，前端按 viewerRole 渲染）。
// 老人信息与档案联系电话来自 ElderProfile；子女账号资料（头像/昵称/手机号，问题2）另读子女 User。
func (s *FamilyService) buildRelation(g model.Guardianship, viewerRole string) (FamilyRelation, error) {
	rel := FamilyRelation{
		GuardianshipId: g.GuardianshipId,
		Status:         model.GuardianshipStatusActive,
		ViewerRole:     viewerRole,
		Relation:       g.Relation,
		ElderUserId:    g.ElderUserId,
		GuardianUserId: g.GuardianUserId,
	}
	// TaUserId 兼容旧前端：老人端看子女、子女端看老人
	if viewerRole == model.RoleElder {
		rel.TaUserId = g.GuardianUserId
	} else {
		rel.TaUserId = g.ElderUserId
	}
	profile, err := s.elderProfileMapper.GetByUserId(g.ElderUserId)
	if err != nil {
		logrus.WithError(err).WithField("elderUserId", g.ElderUserId).Warn("build relation get elder profile failed")
	} else if profile != nil {
		rel.ElderRealName = profile.RealName
		rel.ElderCity = profile.City
		rel.RemindTime = profile.RemindTime
		rel.ElderHealthNote = profile.HealthNote
		rel.ElderPhone = profile.ElderPhone
		rel.GuardianPhone = profile.GuardianPhone
	}
	// ElderAddress 收货真相源已迁到地址簿（问题3）：取默认地址回填，便于编辑页展示当前默认收货地址。
	// 完整地址簿走 /address/list 取；此处仅给默认地址快照供概览。无默认地址则留零值。
	if def, err := s.shippingAddressMapper.GetDefaultByElder(g.ElderUserId); err != nil {
		logrus.WithError(err).WithField("elderUserId", g.ElderUserId).Warn("build relation get default address failed")
	} else if def != nil {
		rel.ElderAddress = def.Address
	}
	// 子女账号资料（问题2）：从子女 User 取头像/昵称/手机号（脱敏），供老人端展示「我的子女」。
	// best-effort：查不到不阻断，留零值。
	if gu, err := s.userMapper.GetByUserId(g.GuardianUserId); err != nil {
		logrus.WithError(err).WithField("guardianUserId", g.GuardianUserId).Warn("build relation get guardian user failed")
	} else if gu != nil {
		rel.GuardianAvatar = gu.Avatar
		rel.GuardianName = gu.Nickname
		rel.GuardianAccountPhone = maskPhone(gu.Phone)
	}
	return rel, nil
}

// ProfileUpdate 编辑老人档案的入参（整页全量覆盖，前端从 view 回填后提交）。
// 注意：收货地址已迁到独立地址簿（问题3），不再随档案更新——增删改/设默认走 /warm-nest/v1/address/*。
type ProfileUpdate struct {
	RealName      string
	City          string
	RemindTime    string
	HealthNote    string
	Relation      string // 子女对老人的称呼（挂守护边，按 viewer→elder 关系更新；老人本人调用时忽略）
	ElderPhone    string // 老人本人电话（可改）
	GuardianPhone string // 守护人电话（可改）
}

// UpdateProfile 编辑老人档案（双端同源：老人本人或其 active 守护人均可改）。
// 全量覆盖语义——调用方须回填全部字段，未传字段会被覆盖为空。
// 提醒时间是老人级唯一设置，随档案一并落库，任一端改动对该老人所有守护关系即时生效。
// relation（问题1）挂在守护边而非档案：按「viewer→elder」的 ACTIVE 关系边更新称呼；
// 老人本人调用（无对应守护边）时 relation 传值忽略、不报错。
func (s *FamilyService) UpdateProfile(viewerUserId, elderUserId string, in ProfileUpdate) error {
	// 权限：经统一的守护关系访问校验（本人或 active 守护人），否则拒 ErrNotGuardian
	ok, err := s.canAccessElder(viewerUserId, elderUserId)
	if err != nil {
		return fmt.Errorf("update profile check access %s->%s: %w", viewerUserId, elderUserId, err)
	}
	if !ok {
		return errors.NewWithArgs(wnconst.ErrNotGuardian)
	}

	profile, err := s.elderProfileMapper.GetByUserId(elderUserId)
	if err != nil {
		return fmt.Errorf("update profile get profile %s: %w", elderUserId, err)
	}
	if profile == nil {
		return fmt.Errorf("elder profile not found: %s", elderUserId)
	}
	profile.RealName = in.RealName
	profile.City = in.City
	profile.RemindTime = in.RemindTime
	profile.HealthNote = in.HealthNote
	profile.ElderPhone = in.ElderPhone
	profile.GuardianPhone = in.GuardianPhone
	if err = s.elderProfileMapper.Update(profile); err != nil {
		return err
	}

	// relation 挂在守护边：仅子女端（viewer 是该老人的守护人）才有对应边可改；老人本人调用无边、忽略
	if viewerUserId != elderUserId && in.Relation != "" {
		g, err := s.guardianshipMapper.GetActive(viewerUserId, elderUserId)
		if err != nil {
			return fmt.Errorf("update profile get guardianship %s->%s: %w", viewerUserId, elderUserId, err)
		}
		if g != nil && g.Relation != in.Relation {
			if err = s.guardianshipMapper.UpdateRelation(g.GuardianshipId, in.Relation); err != nil {
				return fmt.Errorf("update relation %s: %w", g.GuardianshipId, err)
			}
		}
	}
	return nil
}

// canAccessElder 守护关系访问校验的唯一落点（问题4/5/6/7/8/11 共用）：
// viewer 是老人本人（viewer==elder）→ 放行；否则查 viewer 作为守护人→elder 的 ACTIVE 边命中→放行。
// 所有「访问某老人数据」的鉴权都收口到此，不在各处内联 GetActive。
func (s *FamilyService) canAccessElder(viewerUserId, elderUserId string) (bool, error) {
	if viewerUserId == elderUserId {
		return true, nil
	}
	g, err := s.guardianshipMapper.GetActive(viewerUserId, elderUserId)
	if err != nil {
		return false, fmt.Errorf("can access elder %s->%s: %w", viewerUserId, elderUserId, err)
	}
	return g != nil, nil
}

// isElder 当前用户是否为被守护人（作为 elder 存在任一 ACTIVE 守护边）。
func (s *FamilyService) isElder(userId string) (bool, error) {
	asElder, err := s.guardianshipMapper.ListByElder(userId)
	if err != nil {
		return false, fmt.Errorf("check is elder %s: %w", userId, err)
	}
	return len(asElder) > 0, nil
}

// ResolveElder 列表/概览类公用接口的入参解析（模式A），按需求两分支：
//   - 传了 elderUserId（且非自己）：守护人查老人——校验 viewer 是该老人的 ACTIVE 守护人，否则 ErrNotGuardian。
//   - 没传（或传自己）：老人查自己——校验 viewer 确实是被守护人，否则 ErrNotGuardian
//     （守护人不传 elderUserId 属错误用法：子女自身非被守护人、无数据，必须显式拦而非静默返空）。
func (s *FamilyService) ResolveElder(viewerUserId, reqElderUserId string) (string, error) {
	if reqElderUserId == "" || reqElderUserId == viewerUserId {
		isElder, err := s.isElder(viewerUserId)
		if err != nil {
			return "", err
		}
		if !isElder {
			return "", errors.NewWithArgs(wnconst.ErrNotGuardian)
		}
		return viewerUserId, nil
	}
	ok, err := s.canAccessElder(viewerUserId, reqElderUserId)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.NewWithArgs(wnconst.ErrNotGuardian)
	}
	return reqElderUserId, nil
}

// EnsureCanAccess detail 类公用接口的校验（模式B）：记录归属老人已知时，校验 viewer 是否可访问。
// 不可访问返 ErrNotGuardian。供 checkin/reward detail 在查出记录归属后调用。
func (s *FamilyService) EnsureCanAccess(viewerUserId, elderUserId string) error {
	ok, err := s.canAccessElder(viewerUserId, elderUserId)
	if err != nil {
		return err
	}
	if !ok {
		return errors.NewWithArgs(wnconst.ErrNotGuardian)
	}
	return nil
}
