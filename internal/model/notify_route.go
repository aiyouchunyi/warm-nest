// Package model @Author larry
// @Date 2026/06/22
// @Desc 通知路由配置模型（PRD §5.2：通知渠道按业务场景可配，非全局一刀切）
//
// 设计要点：
//   - 上层业务只认「场景」(Scene)，不认渠道。某场景走哪个渠道由本表配置决定，运营可改、不重启。
//     例：未打卡提醒推老人走小程序订阅、推子女走服务号模板——同类消息不同场景可配不同渠道。
//   - 渠道(Channel)对应一个 Notifier 策略实现（subscribe/template/未来 sms）；新增渠道=加实现+注册，
//     本表与上层 deliver 不动（开闭）。
//   - TemplateId 随渠道走：订阅消息模板ID 与服务号模板ID 不同主体，改渠道须同时改模板ID，故同表存。
//   - 未配/禁用的场景由代码默认渠道(DefaultChannel)兜底；内消息流(通道①)不受本表影响、必达。
package model

import "warm-nest/pkg/tool/database"

// 通知场景（Scene）——与消息触发点一一对应，是按业务配渠道的最小粒度（PRD §5.2）
const (
	NotifySceneCheckInNotice     = "CHECKIN_NOTICE"      // 打卡通知（老人打卡→推子女）
	NotifySceneNotRemindGuardian = "NOT_REMIND_GUARDIAN" // 未打卡提醒（宽限后→推子女）
	NotifySceneNotRemindElder    = "NOT_REMIND_ELDER"    // 未打卡提醒第一段（→推老人本人）
	NotifySceneBindSuccess       = "BIND_SUCCESS"        // 绑定成功反馈（→推子女）
	NotifySceneAddressPreheat    = "ADDRESS_PREHEAT"     // 奖励地址预热提醒（每月25号、收货地址空→推子女，§6.6.3）
)

// 通知渠道（Channel）——每个值对应一个 Notifier 策略实现
const (
	NotifyChannelSubscribe = "subscribe" // 小程序订阅消息（touser=小程序 openid，小程序 access_token）
	NotifyChannelTemplate  = "template"  // 服务号模板消息（touser=服务号 openid，服务号 access_token）
	NotifyChannelSms       = "sms"       // 短信（扩展点，暂未实现）
	NotifyChannelLog       = "log"       // 只打日志不真发：无任何微信配置时跑通完整业务链路用（测试/联调）
)

// IsValidNotifyChannel 判渠道是否为已注册的合法值（config 兜底渠道 / 管理接口改配置时校验）。
// sms 是预留枚举但未实现 Notifier，不算合法可用渠道。
func IsValidNotifyChannel(channel string) bool {
	switch channel {
	case NotifyChannelSubscribe, NotifyChannelTemplate, NotifyChannelLog:
		return true
	default:
		return false
	}
}

// NotifyRoute 通知路由配置（一个场景一条，运营可改渠道与模板ID）
type NotifyRoute struct {
	database.Identify
	Scene      string `gorm:"size:32;uniqueIndex:u_notify_scene;comment:通知场景[CHECKIN_NOTICE/NOT_REMIND_GUARDIAN/NOT_REMIND_ELDER/BIND_SUCCESS]" json:"scene"`
	Channel    string `gorm:"size:16;comment:下发渠道[subscribe:小程序订阅,template:服务号模板,sms:短信]" json:"channel"`
	TemplateId string `gorm:"size:128;comment:该渠道对应的模板ID（订阅消息模板与服务号模板不同主体，随渠道走）" json:"templateId"`
	PagePath   string `gorm:"size:128;comment:服务号模板消息点击跳转的小程序页（template 渠道用，空则代码按场景兜底首页）" json:"pagePath"`
	Enabled    bool   `gorm:"comment:是否启用外推（false 仅落内消息流、不走外部渠道）" json:"enabled"`

	database.SystemField
}
