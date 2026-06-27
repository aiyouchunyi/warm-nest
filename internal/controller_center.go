// Package internal @Author larry
// @Desc 控制器注册中心
//
// 【新增 controller 后必做】在 RegisterController() 切片里追加一行：
//
//	controller.GetXxxController(),
package internal

import (
	"warm-nest/pkg/app/web"

	"warm-nest/internal/controller"
)

// RegisterController 注册控制器
func RegisterController() []web.Controller {
	return []web.Controller{
		controller.GetUserController(),
		controller.GetInvitationController(),
		controller.GetCheckInController(),
		controller.GetMessageController(),
		controller.GetRewardController(),
		controller.GetRewardAdminController(),
		controller.GetNotifyRouteAdminController(),
		controller.GetFamilyController(),
		controller.GetAddressController(),
		controller.GetUploadController(),
		controller.GetOfficialCallbackController(),
		controller.GetTestController(), // ⚠临时测试造数据接口，验证完手动删除本行 + test_controller.go
	}
}
