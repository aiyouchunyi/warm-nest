// Package constant @Author larry
// @Date 2026/06/18
// @Desc 业务错误码（warm-nest 自有，区别于框架 00000xxx 系统码）
//
// 为什么新建本文件：项目原先全用 fmt.Errorf（无 code），前端拿到的都归框架统一的
// "系统内部错误"码、无法分场景渲提示。本文件用框架 errors.CodeError（NewFail 会把 Code
// 透传到响应 mcode），给「需要前端分别处理」的业务拒绝场景分配可区分错误码。
// 仅在「前端要据此渲不同文案」的拒绝点用；纯内部错误仍用 fmt.Errorf 即可。
//
// 编码段位：warm-nest 业务码用 1xxxxxxx 段，避开框架 0xxxxxxx 系统码。
//
//	10001xxx 邀请/绑定（invitation 域）
package constant

import "warm-nest/pkg/app/errors"

// 邀请接受/绑定守卫错误码（§8.0.3 一对一守卫 + 一次性）
var (
	// ErrInviteUsed 邀请已被他人接受（卡片被转发场景）——前端提示"邀请已失效，请让家人重新发送"
	ErrInviteUsed = errors.NewCode("10001001", "邀请已被使用，请重新获取邀请")
	// ErrElderBound 老人已被别的子女绑定——前端提示"该老人已被其他家人绑定"
	ErrElderBound = errors.NewCode("10001002", "该老人已被其他家人绑定")
	// ErrGuardianBound 该子女已绑定别的老人（1 期一对一）——前端提示"你已绑定其他老人"
	ErrGuardianBound = errors.NewCode("10001003", "你已绑定其他老人，暂不支持绑定多位")
	// ErrInviteState 邀请状态不可接受（已撤销/已过期/PENDING 超时）——前端提示"邀请已失效"
	ErrInviteState = errors.NewCode("10001004", "邀请已失效")
	// ErrAlreadyGuardian 接受者已是守护人（已发起邀请或已绑老人）——角色互斥，前端提示"你已是守护人，不能接受邀请"
	ErrAlreadyGuardian = errors.NewCode("10001005", "你已是守护人，不能再接受邀请")
	// ErrAlreadyElder 发起者已是被守护人——角色互斥，前端提示"你已是被守护人，不能发起邀请"
	ErrAlreadyElder = errors.NewCode("10001006", "你已是被守护人，不能发起邀请")
	// ErrInviteNotCancelable 邀请非可撤销状态（仅 PENDING 可撤销）——前端提示"该邀请无法撤销"
	ErrInviteNotCancelable = errors.NewCode("10001007", "该邀请无法撤销")
	// ErrInviteNotOwner 非本人发起的邀请，无权撤销——前端提示"无权操作该邀请"
	ErrInviteNotOwner = errors.NewCode("10001008", "无权操作该邀请")
)

// 领取/地址错误码（reward 域，10002xxx）
var (
	// ErrAddressRequired 领取奖励前未填收货地址——前端提示"请先在家庭信息补充收货地址"
	ErrAddressRequired = errors.NewCode("10002001", "请先补充收货地址再领取")
	// ErrReceiverRequired 领取前未填收货人——前端提示"请先补充收货人"
	ErrReceiverRequired = errors.NewCode("10002002", "请先补充收货人再领取")
	// ErrReceiverPhoneRequired 领取前未填收货电话——前端提示"请先补充收货电话"
	ErrReceiverPhoneRequired = errors.NewCode("10002003", "请先补充收货电话再领取")
)

// 守护关系/权限错误码（family 域，10003xxx）
var (
	// ErrNotGuardian 当前登录人与目标老人无 ACTIVE 守护关系，无权访问其数据——前端提示"无权查看该长辈的信息"
	ErrNotGuardian = errors.NewCode("10003001", "无权访问该长辈的信息")
)
