// Package notify @Author larry
// @Date 2026/06/24
// @Desc 服务号模板消息「场景→模板字段」翻译层
//
// 为什么需要：上层业务 params 用语义 key（relation/checkInTime/deadline…），同一份 params 还要喂
// 内消息流展示（Message.Params，前端按语义 key 渲染）；而服务号模板要求字段名是平台分配的占位
// （thing1.DATA / time3.DATA…，每个模板各异）。两套命名不能混用，故在「外推到服务号」这一步按场景
// 把语义 params 翻译成对应模板的字段，并补模板固定文案（如「考勤情况=超过约定时间仍未打卡」）。
//
// 字段归属：本翻译只服务于 template（服务号模板）渠道——字段ID与文案是服务号平台模板特有。若将来某
// 场景改走 subscribe（小程序订阅），订阅模板字段名不同，应在订阅侧另建映射，互不影响（开闭）。
//
// 语义 params 约定（由 message_service 各 Send* 调用方备好）：
//   - relation     子女对老人的关系称呼枚举（model.Relation*），翻译时转中文
//   - elderName    老人真实姓名（仅 NOT_REMIND_ELDER 发给老人本人时用，无 relation 可用）
//   - checkInTime  打卡时刻，已格式化 "YYYY-MM-DD HH:mm"
//   - deadline     打卡截止时刻，已格式化 "YYYY-MM-DD HH:mm"
//   - bindTime     绑定成立时刻，已格式化 "YYYY-MM-DD HH:mm"

package notify

import (
	"fmt"

	"warm-nest/pkg/kinds/jsons"

	"warm-nest/internal/model"
)

// 模板固定文案（平台模板里语义恒定、不随单次数据变的字段值）
const (
	preheatAddressTip = "填好长辈的收货地址，达标后好第一时间为您寄出" // 预热填写地址 thing2.DATA
	notRemindAttend   = "超过约定时间仍未打卡"             // 漏打卡 thing4.DATA（考勤情况）
)

// templateFieldBuilder 把一个场景的语义 params 翻译成该场景服务号模板的字段集。
type templateFieldBuilder func(params jsons.JSONObject) jsons.JSONObject

// elderTitle 取「老人称呼」用于模板展示：优先 relation 中文（子女视角），无 relation 时退老人真名。
// 两者皆空兜底「家人」（model.RelationText 已对未知值兜底），避免模板字段为空被微信拒发。
func elderTitle(params jsons.JSONObject) string {
	if r, ok := params["relation"].(string); ok && r != "" {
		return model.RelationText(r)
	}
	if n, ok := params["elderName"].(string); ok && n != "" {
		return n
	}
	return "家人"
}

func str(params jsons.JSONObject, key string) string {
	if v, ok := params[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// templateFieldBuilders 场景→服务号模板字段翻译（字段ID来自各平台模板定义，prd 联调下发）。
var templateFieldBuilders = map[string]templateFieldBuilder{
	// 打卡成功通知：用户名=老人称呼，签到时间=打卡时刻
	model.NotifySceneCheckInNotice: func(p jsons.JSONObject) jsons.JSONObject {
		return jsons.JSONObject{"thing1": elderTitle(p), "time3": str(p, "checkInTime")}
	},
	// 预热填写地址：客户名称=老人称呼，欠料名称=固定提示文案
	model.NotifySceneAddressPreheat: func(p jsons.JSONObject) jsons.JSONObject {
		return jsons.JSONObject{"thing1": elderTitle(p), "thing2": preheatAddressTip}
	},
	// 漏打卡-推子女：姓名=老人称呼，考勤情况=固定文案，截止时间=打卡截止时刻
	model.NotifySceneNotRemindGuardian: func(p jsons.JSONObject) jsons.JSONObject {
		return jsons.JSONObject{"thing8": elderTitle(p), "thing4": notRemindAttend, "time5": str(p, "deadline")}
	},
	// 漏打卡-推老人本人：同模板（姓名退老人真名），考勤情况/截止时间同上
	model.NotifySceneNotRemindElder: func(p jsons.JSONObject) jsons.JSONObject {
		return jsons.JSONObject{"thing8": elderTitle(p), "thing4": notRemindAttend, "time5": str(p, "deadline")}
	},
	// 老人绑定成功-推子女：绑定用户=老人称呼，绑定时间=接受邀请时刻
	model.NotifySceneBindSuccess: func(p jsons.JSONObject) jsons.JSONObject {
		return jsons.JSONObject{"thing1": elderTitle(p), "time2": str(p, "bindTime")}
	},
}

// buildTemplateFields 按场景把语义 params 翻译成服务号模板字段；无映射的场景原样返回（兜底不外推空）。
func buildTemplateFields(scene string, params jsons.JSONObject) jsons.JSONObject {
	if b, ok := templateFieldBuilders[scene]; ok {
		return b(params)
	}
	return params
}
