// Package notify @Author larry
// @Date 2026/06/22
// @Desc 服务号模板消息渠道策略（touser=服务号 openid，服务号 access_token，PRD §5.2 主推）
//
// 收件人解析：业务 userId → User.UnionId → Fan.OfficialOpenId（跨主体锚点 UnionId）。
// 服务号模板消息天然只能发给已关注用户，故未关注（无 Fan / 无 OfficialOpenId）soft-skip。

package notify

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/kinds/jsons"

	// 【临时】恢复 Miniprogram 跳转时，需一并加回： "warm-nest/internal/config"
	"warm-nest/internal/feign/wechat"
	wxmodel "warm-nest/internal/feign/wechat/model"
	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
)

// 模板消息点击跳转配置（PRD §5.2：服务号模板卡片点「详情」→ 唤起小程序）。
// pagepath 须为小程序已发布版本里真实存在的页面，否则微信下发报 errcode=40165 invalid weapp pagepath。
//
// 跳转页取值优先级：notify_route 表配置（ctx.PagePath，运营可改）→ 代码场景兜底（templateScenePages）
// → 默认首页（templateDefaultJumpPage）。表为唯一真相源，代码两层兜底保证存量行/漏配场景仍可跳。
const (
	// templateDefaultJumpPage 未单独配置的场景兜底跳转页（登录页，走登录态分发）
	templateDefaultJumpPage = "pages/login/index"
	// templateFallbackURL 老微信版本不支持跳小程序时的 H5 兜底（降级打开，不卡死）
	templateFallbackURL = "http://www.changkankan.com"
)

// templateScenePages 各通知场景跳转页的代码兜底（DB 未配/存量行为空时用）；未列场景再退默认首页。
var templateScenePages = map[string]string{
	model.NotifySceneCheckInNotice:     "pages/login/index",         // 打卡成功通知→登录页
	model.NotifySceneNotRemindGuardian: "pages/login/index",         // 漏打卡-推子女→登录页
	model.NotifySceneNotRemindElder:    "pages/login/index",         // 漏打卡-推老人→登录页
	model.NotifySceneBindSuccess:       "pages/login/index",         // 绑定成功-推子女→登录页
	model.NotifySceneAddressPreheat:    "pages/child/address/index", // 预热填地址→填地址页
}

// templateJumpPage 取跳转页：DB 配置（pagePath）优先，空则按场景代码兜底，再无则默认首页。
func templateJumpPage(scene, pagePath string) string {
	if pagePath != "" {
		return pagePath
	}
	if p, ok := templateScenePages[scene]; ok {
		return p
	}
	return templateDefaultJumpPage
}

// templateNotifier 服务号模板消息渠道
type templateNotifier struct {
	userMapper *mapper.UserMapper
	fanMapper  *mapper.FanMapper
}

func newTemplateNotifier() *templateNotifier {
	return &templateNotifier{
		userMapper: mapper.GetUserMapper(),
		fanMapper:  mapper.GetFanMapper(),
	}
}

// Channel 渠道标识
func (n *templateNotifier) Channel() string { return model.NotifyChannelTemplate }

// Push 下发服务号模板消息（best-effort）：解析服务号 openid → 调模板消息接口。
func (n *templateNotifier) Push(ctx PushContext) error {
	if ctx.TemplateId == "" {
		return nil // 未配模板ID，不外推
	}
	log := logrus.WithFields(logrus.Fields{"receiverId": ctx.ReceiverUserId, "channel": n.Channel()})
	officialOpenId, err := n.resolveOfficialOpenId(ctx.ReceiverUserId)
	if err != nil {
		return err
	}
	if officialOpenId == "" {
		log.Warn("template push skip: receiver not subscribed to official account (no official openid)")
		return nil
	}
	// 语义 params → 该场景服务号模板字段（thing1.DATA/time3.DATA…）+ 模板固定文案
	fields := buildTemplateFields(ctx.Scene, ctx.Params)
	resp, err := wechat.Client().SendTemplateMessage(wxmodel.SendTemplateMsgReq{
		ToUser:     officialOpenId,
		TemplateId: ctx.TemplateId,
		// 【临时·小程序未发布线上版】小程序未正式发布时，模板消息带 miniprogram 节点会被微信拦
		// （errcode=40165 invalid weapp pagepath，连默认首页也拒），导致整条模板消息发不出。
		// 故暂时去掉 Miniprogram 节点、仅留 Url（H5 兜底）：消息正常下发，点击走 H5。
		// 小程序正式发布线上版后，恢复下方注释的 Miniprogram 块即可（miniprogram_state 仅订阅消息支持，模板消息无此字段、无法绕过）。
		// Miniprogram: &wxmodel.TmplMiniprogram{
		// 	AppId:    config.WeChatConf().AppId,
		// 	PagePath: templateJumpPage(ctx.Scene, ctx.PagePath),
		// },
		Url:  templateFallbackURL,
		Data: toTemplateData(fields),
	})
	if err != nil {
		return fmt.Errorf("template notifier send: %w", err)
	}
	switch resp.ErrCode {
	case wxmodel.TmplMsgOk:
	case wxmodel.TmplMsgInvalidToUser, wxmodel.TmplMsgNeedSubscribe:
		log.Warn("template push: receiver not subscribed / invalid openid")
	default:
		log.Warnf("template push errcode=%d errmsg=%s", resp.ErrCode, resp.ErrMsg)
	}
	return nil
}

// resolveOfficialOpenId 业务 userId → 服务号 openid（经 UnionId 关联 Fan）；未关注返回空串。
func (n *templateNotifier) resolveOfficialOpenId(userId string) (string, error) {
	user, err := n.userMapper.GetByUserId(userId)
	if err != nil {
		return "", fmt.Errorf("template notifier get user %s: %w", userId, err)
	}
	if user == nil || user.UnionId == "" {
		return "", nil
	}
	fan, err := n.fanMapper.GetByUnionId(user.UnionId)
	if err != nil {
		return "", fmt.Errorf("template notifier get fan: %w", err)
	}
	if fan == nil || !fan.Subscribed {
		return "", nil
	}
	return fan.OfficialOpenId, nil
}

// toTemplateData 把渲染参数转成服务号模板 data 结构（值统一包 value，颜色由模板侧定）
func toTemplateData(params jsons.JSONObject) map[string]wxmodel.TmplDataField {
	data := make(map[string]wxmodel.TmplDataField, len(params))
	for k, v := range params {
		data[k] = wxmodel.TmplDataField{Value: fmt.Sprintf("%v", v)}
	}
	return data
}
