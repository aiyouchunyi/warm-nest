// Package service @Author larry
// @Date 2026/06/15
// @Desc 用户登录服务

package service

import (
	"fmt"
	"sync"
	"time"

	"warm-nest/pkg/tool/auth/jwt"
	"warm-nest/pkg/utils/rands"

	"warm-nest/internal/feign/wechat"
	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
)

// UserService 用户登录服务
type UserService struct {
	userMapper         *mapper.UserMapper
	guardianshipMapper *mapper.GuardianshipMapper
}

var userService *UserService
var userServiceOnce sync.Once

// GetUserService 获取用户登录服务单例
func GetUserService() *UserService {
	userServiceOnce.Do(func() {
		userService = &UserService{
			userMapper:         mapper.GetUserMapper(),
			guardianshipMapper: mapper.GetGuardianshipMapper(),
		}
	})
	return userService
}

// LoginResult 登录结果
type LoginResult struct {
	Token          string   `json:"token"`
	UserId         string   `json:"userId"`
	AvailableRoles []string `json:"availableRoles"` // 从守护关系边实时算（ELDER/GUARDIAN）
	LastActiveRole string   `json:"lastActiveRole"` // 上次选的端，前端默认选中
}

// LoginByCode 小程序 code 登录：换 openid → 查/建 User → 签 JWT → 算可用身份
func (s *UserService) LoginByCode(code string) (*LoginResult, error) {
	sess, err := wechat.Client().Code2Session(code)
	if err != nil {
		return nil, fmt.Errorf("login code2session: %w", err)
	}

	user, err := s.userMapper.GetByOpenId(sess.OpenId)
	if err != nil {
		return nil, fmt.Errorf("login get user by openid: %w", err)
	}
	if user == nil {
		user = &model.User{
			UserId:  rands.Numeric(),
			OpenId:  sess.OpenId,
			UnionId: sess.UnionId,
			Status:  model.UserStatusNormal,
		}
		if err = s.userMapper.Create(user); err != nil {
			return nil, fmt.Errorf("login create user: %w", err)
		}
	}

	token, err := jwt.GetJWTService().Token(user.UserId)
	if err != nil {
		return nil, fmt.Errorf("login sign jwt: %w", err)
	}

	roles, err := s.availableRoles(user.UserId)
	if err != nil {
		return nil, fmt.Errorf("login resolve available roles: %w", err)
	}
	return &LoginResult{
		Token:          token,
		UserId:         user.UserId,
		AvailableRoles: roles,
		LastActiveRole: user.LastActiveRole,
	}, nil
}

// availableRoles 算可用身份：作 elder 出现→ELDER；作 guardian 出现「或」发起了有效 PENDING 邀请→GUARDIAN。
// PENDING 邀请也认作 GUARDIAN（问题1）：子女发起邀请、老人尚未接受时还没有守护边，
// 但其已是子女身份，应能进入「等待老人接受」页——与 isAlreadyGuardian 的邀请判据同源。
func (s *UserService) availableRoles(userId string) ([]string, error) {
	roles := make([]string, 0, 2)
	asElder, err := s.guardianshipMapper.ListByElder(userId)
	if err != nil {
		return nil, fmt.Errorf("available roles list by elder: %w", err)
	}
	if len(asElder) > 0 {
		roles = append(roles, model.RoleElder)
	}
	asGuardian, err := s.guardianshipMapper.ListByGuardian(userId)
	if err != nil {
		return nil, fmt.Errorf("available roles list by guardian: %w", err)
	}
	isGuardian := len(asGuardian) > 0
	if !isGuardian {
		// 无守护边时再看有效 PENDING 邀请（去重：有守护边已判定为 GUARDIAN，不重复 append）
		hasInvite, err := GetInvitationService().HasActiveInvitationAsGuardian(userId, time.Now().UnixMilli())
		if err != nil {
			return nil, fmt.Errorf("available roles check active invitation: %w", err)
		}
		isGuardian = hasInvite
	}
	if isGuardian {
		roles = append(roles, model.RoleGuardian)
	}
	return roles, nil
}

// ResolvePhone 凭 phoneCode 向微信换真实手机号明文（问题5）。不落库，仅换号供前端回填。
// ⚠️ 手机号敏感，调用方勿打日志。
func (s *UserService) ResolvePhone(phoneCode string) (string, error) {
	phone, err := wechat.Client().GetPhoneNumber(phoneCode)
	if err != nil {
		return "", fmt.Errorf("resolve phone: %w", err)
	}
	return phone, nil
}

// MyProfile 当前登录用户的资料（问题2：子女查看/编辑本人头像、昵称、手机号）。
// 手机号返回明文供本人编辑页回填（仅本人可见自己的号；对家庭成员展示的脱敏号见 family/view）。
type MyProfile struct {
	UserId   string `json:"userId"`
	Nickname string `json:"nickname"` // 微信昵称
	Avatar   string `json:"avatar"`   // 头像 URL
	Phone    string `json:"phone"`    // 本人手机号明文（仅本人可见）
}

// GetMyProfile 查当前登录用户本人资料（问题2，供「我的资料」页回填）。
func (s *UserService) GetMyProfile(userId string) (*MyProfile, error) {
	user, err := s.userMapper.GetByUserId(userId)
	if err != nil {
		return nil, fmt.Errorf("get my profile %s: %w", userId, err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found: %s", userId)
	}
	return &MyProfile{
		UserId:   user.UserId,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Phone:    user.Phone,
	}, nil
}

// MyProfileUpdate 编辑本人资料入参（问题2）。指针字段=patch 语义：nil 不动该字段，非 nil（含空串）覆盖。
// 区别于 family/profile 的全量覆盖：本人资料各字段独立可改，避免前端漏传把头像/手机号清空。
type MyProfileUpdate struct {
	Nickname *string
	Avatar   *string
	Phone    *string // 明文（前端先经 /user/resolve-phone 换好再传）
}

// UpdateMyProfile 编辑当前登录用户本人资料（问题2，patch 语义只改传入的字段）。
// ⚠️ 手机号敏感，勿打日志。
func (s *UserService) UpdateMyProfile(userId string, in MyProfileUpdate) error {
	user, err := s.userMapper.GetByUserId(userId)
	if err != nil {
		return fmt.Errorf("update my profile get user %s: %w", userId, err)
	}
	if user == nil {
		return fmt.Errorf("user not found: %s", userId)
	}
	if in.Nickname != nil {
		user.Nickname = *in.Nickname
	}
	if in.Avatar != nil {
		user.Avatar = *in.Avatar
	}
	if in.Phone != nil {
		user.Phone = *in.Phone
	}
	if err := s.userMapper.Update(user); err != nil {
		return fmt.Errorf("update my profile %s: %w", userId, err)
	}
	return nil
}
