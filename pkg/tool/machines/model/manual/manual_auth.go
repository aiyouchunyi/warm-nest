// Package auth @Author larry
// @Date 2025/11/13 13:49
// @Desc

package manual

import (
	"warm-nest/pkg/tool/machines/model/constant"
	"warm-nest/pkg/tool/machines/model/context"
)

type AuthFunc func(ctx *context.MachineContext, cfg ManualConfig, event ManualEvent) error

type AuthSetting struct {
	settings  map[string][]string
	authFuncs []AuthFunc
}

func NewAuthSetting() AuthSetting {
	return AuthSetting{
		settings:  make(map[string][]string),
		authFuncs: make([]AuthFunc, 0),
	}
}

// AuthUser 设置用户权限
func (auth AuthSetting) AuthUser(users ...string) AuthSetting {
	auth.settings[constant.AuthKindUser] = users
	return auth
}

// AuthRole 设置角色权限
func (auth AuthSetting) AuthRole(roles ...string) AuthSetting {
	auth.settings[constant.AuthKindRole] = roles
	return auth
}

// AuthExpr 设置表达式权限
func (auth AuthSetting) AuthExpr(exprs ...string) AuthSetting {
	auth.settings[constant.AuthKindExpr] = exprs
	return auth
}

func (auth AuthSetting) AuthFunc(fns ...AuthFunc) AuthSetting {
	auth.authFuncs = append(auth.authFuncs, fns...)
	return auth
}

// GetAuth 获取指定类型的权限信息
func (auth AuthSetting) GetAuth(kind string) ([]string, bool) {
	info, ok := auth.settings[kind]
	return info, ok
}

func (auth AuthSetting) GetAuthFunc() []AuthFunc {
	return auth.authFuncs
}
