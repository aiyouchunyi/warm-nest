// Package notify @Author larry
// @Date 2026/06/22
// @Desc 小程序订阅消息渠道策略（touser=小程序 openid，小程序 access_token）
//
// 一次性订阅：用户授权一次发一次，授权次数微信侧管理、后端无状态直发。
// 缺接收方 openid / 模板ID 时 soft-skip（记日志返 nil）；未授权/失效记日志不报错。

package notify

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/kinds/jsons"

	"warm-nest/internal/config"
	"warm-nest/internal/feign/wechat"
	wxmodel "warm-nest/internal/feign/wechat/model"
	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
)

// subscribeNotifier 小程序订阅消息渠道
type subscribeNotifier struct {
	userMapper *mapper.UserMapper
}

func newSubscribeNotifier() *subscribeNotifier {
	return &subscribeNotifier{userMapper: mapper.GetUserMapper()}
}

// Channel 渠道标识
func (n *subscribeNotifier) Channel() string { return model.NotifyChannelSubscribe }

// Push 下发小程序订阅消息（best-effort）：取接收方小程序 openid → 调订阅消息接口。
func (n *subscribeNotifier) Push(ctx PushContext) error {
	if ctx.TemplateId == "" {
		return nil // 未配模板ID，不外推
	}
	log := logrus.WithFields(logrus.Fields{"receiverId": ctx.ReceiverUserId, "channel": n.Channel()})
	user, err := n.userMapper.GetByUserId(ctx.ReceiverUserId)
	if err != nil {
		return fmt.Errorf("subscribe notifier get user %s: %w", ctx.ReceiverUserId, err)
	}
	if user == nil || user.OpenId == "" {
		log.Warn("subscribe push skip: receiver miniprogram openid missing")
		return nil
	}
	resp, err := wechat.Client().SendSubscribeMessage(wxmodel.SendSubscribeMsgReq{
		ToUser:           user.OpenId,
		TemplateId:       ctx.TemplateId,
		MiniprogramState: config.WeChatConf().MiniprogramState,
		Lang:             "zh_CN",
		Data:             toSubscribeData(ctx.Params),
	})
	if err != nil {
		return fmt.Errorf("subscribe notifier send: %w", err)
	}
	switch resp.ErrCode {
	case wxmodel.SubMsgOk:
	case wxmodel.SubMsgNotAuthorized, wxmodel.SubMsgRejected:
		log.Warn("subscribe push: user not authorized, need re-subscribe on frontend")
	default:
		log.Warnf("subscribe push errcode=%d errmsg=%s", resp.ErrCode, resp.ErrMsg)
	}
	return nil
}

// toSubscribeData 把渲染参数转成订阅消息 data 结构（值统一包 value）
func toSubscribeData(params jsons.JSONObject) map[string]wxmodel.MsgDataField {
	data := make(map[string]wxmodel.MsgDataField, len(params))
	for k, v := range params {
		data[k] = wxmodel.MsgDataField{Value: fmt.Sprintf("%v", v)}
	}
	return data
}
