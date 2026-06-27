// Package manual @Author larry
// @Date 2025/11/13 13:58
// @Desc

package manual

type ActionConfig struct {
	Action string      `json:"action" gorm:"comment:动作名称"`
	Auths  AuthSetting `json:"auths" gorm:"comment:权限配置"`
}
