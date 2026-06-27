// Package approve @Author larry
// @Date 2025/3/7 11:22
// @Desc

package manual

import (
	"warm-nest/pkg/utils/maps"
	"warm-nest/pkg/utils/slices"
)

type ManualConfig struct {
	DisableSupper bool                   `json:"disableSupper" gorm:"comment:是否禁用超级管理员权限"`
	Limit         int                    `json:"limit" gorm:"comment:最少操作次数"`
	Auths         AuthSetting            `json:"auths" gorm:"comment:权限配置"`
	ActionAuths   map[string]AuthSetting `json:"actionAuths" gorm:"comment:动作权限配置"`
}

func NewManualConfig(options ...Option) ManualConfig {
	config := &ManualConfig{
		Limit:       1,
		Auths:       NewAuthSetting(),
		ActionAuths: make(map[string]AuthSetting),
	}
	for _, option := range options {
		option(config)
	}
	return *config
}

func (c *ManualConfig) GetAuth(kind string, action string) ([]string, bool) {
	actionAuth, ok1 := c.ActionAuths[action].GetAuth(kind)
	auth, ok2 := c.Auths.GetAuth(kind)
	return slices.Merge(actionAuth, auth), ok1 || ok2
}

func (c *ManualConfig) GetAuthFunc(action string) []AuthFunc {
	authFuncs := make([]AuthFunc, 0)
	authFuncs = append(authFuncs, c.Auths.GetAuthFunc()...)
	if actionAuth, ok := c.ActionAuths[action]; ok {
		authFuncs = append(authFuncs, actionAuth.GetAuthFunc()...)
	}
	return authFuncs
}

func (c *ManualConfig) AddAction(actions ...string) {
	for _, action := range actions {
		maps.PutIfAbsent(c.ActionAuths, action, NewAuthSetting())
	}
}

func (c *ManualConfig) AllowAction(action string) bool {
	_, ok := c.ActionAuths[action]
	return ok
}

type Option func(*ManualConfig)

func WithDisableSupper() Option {
	return func(c *ManualConfig) {
		c.DisableSupper = true
	}
}

func WithLimit(limit int) Option {
	return func(c *ManualConfig) {
		c.Limit = limit
	}
}

func WithActions(actions ...string) Option {
	return func(c *ManualConfig) {
		c.AddAction(actions...)
	}
}

func AuthUser(users ...string) Option {
	return func(c *ManualConfig) {
		c.Auths.AuthUser(users...)
	}
}

func AuthRole(roles ...string) Option {
	return func(c *ManualConfig) {
		c.Auths.AuthRole(roles...)
	}
}

func AuthExpr(exprs ...string) Option {
	return func(c *ManualConfig) {
		c.Auths.AuthExpr(exprs...)
	}
}

func ActionAuthUser(action string, users ...string) Option {
	return func(c *ManualConfig) {
		authSetting := maps.GetOrDefault(c.ActionAuths, action, NewAuthSetting())
		authSetting.AuthUser(users...)
		c.ActionAuths[action] = authSetting
	}
}

func ActionAuthRole(action string, roles ...string) Option {
	return func(c *ManualConfig) {
		authSetting := maps.GetOrDefault(c.ActionAuths, action, NewAuthSetting())
		authSetting.AuthRole(roles...)
		c.ActionAuths[action] = authSetting
	}
}

func ActionAuthExpr(action string, exprs ...string) Option {
	return func(c *ManualConfig) {
		authSetting := maps.GetOrDefault(c.ActionAuths, action, NewAuthSetting())
		authSetting.AuthExpr(exprs...)
		c.ActionAuths[action] = authSetting
	}
}
