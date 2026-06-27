// Package session @Author
// Larry Fine mysql_session.go
// @Date 2024/4/28 19:42:00
// @Desc
package session

import (
	"warm-nest/pkg/kinds"
)

type Session struct {
	ReqId   string `gorm:"comment:请求唯一标识" json:"reqId"`
	ReqUser string `gorm:"comment:请求用户" json:"reqUser"`
	User    User   `gorm:"comment:用户信息" json:"user"`
}

type User struct {
	UserId  string        `gorm:"comment:用户Id" json:"userId"`
	Status  string        `gorm:"comment:用户状态[NORMAL DISABLE]" json:"status"`
	Email   string        `gorm:"comment:用户邮箱" json:"email"`
	Name    string        `gorm:"comment:用户名称" json:"name"`
	Supper  bool          `gorm:"comment:是否超级管理员" json:"supper"`
	RoleIds kinds.Strings `gorm:"comment:角色ID列表" json:"roleIds"`
}

func SystemUser() User {
	return User{
		UserId: "system",
		Name:   "系统用户",
		Supper: true,
	}
}

// 移除 UnmarshalJSON 方法 - 该方法会破坏 Gin 的 ShouldBindJSON 对嵌入该结构体的外层结构的绑定
// See: https://github.com/gin-gonic/gin/issues/3148
