// Package controller @Author larry
// @Date 2026/06/18
// @Desc 服务号消息回调 Controller（PRD §8.2 关注/取关事件，免鉴权——微信服务器直连）
//
// 与其它 controller 不同：本回调由微信公众平台服务器直接请求，不是小程序前端，
// 故①免鉴权（不挂 widgets.Session）②不走 web.Handler 的 JSON 范式——
// 微信用 query 传签名、用原始 XML body 传事件，GET 校验要求原样回写 echostr，
// 都不是「结构体入参 + JSON 出参」，故用裸 gin.HandlerFunc 处理。

package controller

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"warm-nest/internal/config"
	"warm-nest/internal/feign/wechat"
	wxmodel "warm-nest/internal/feign/wechat/model"
	"warm-nest/internal/service"
)

// OfficialCallbackController 服务号消息回调 Controller
type OfficialCallbackController struct{}

var officialCallbackController *OfficialCallbackController
var officialCallbackControllerOnce sync.Once

// GetOfficialCallbackController 获取服务号消息回调 Controller 单例
func GetOfficialCallbackController() *OfficialCallbackController {
	officialCallbackControllerOnce.Do(func() {
		officialCallbackController = &OfficialCallbackController{}
	})
	return officialCallbackController
}

// Router 注册回调路由（免鉴权，裸 handler）
func (c *OfficialCallbackController) Router(router *gin.Engine) {
	// GET = 公众平台配置时的 URL 有效性校验；POST = 关注/取关等事件推送
	router.GET("/warm-nest/v1/wechat/official/callback", c.verify)
	router.POST("/warm-nest/v1/wechat/official/callback", c.event)
}

// verify URL 校验：签名通过则原样回写 echostr（纯文本）
func (c *OfficialCallbackController) verify(ctx *gin.Context) {
	token := config.WeChatConf().OfficialCallbackToken
	signature := ctx.Query("signature")
	timestamp := ctx.Query("timestamp")
	nonce := ctx.Query("nonce")
	echostr := ctx.Query("echostr")
	if !wechat.VerifyOfficialSignature(token, signature, timestamp, nonce) {
		logrus.WithField("signature", signature).Warn("official callback verify signature failed")
		ctx.String(403, "invalid signature")
		return
	}
	ctx.String(200, echostr)
}

// event 事件推送：验签 → 解析 XML → 关注/取关落库。无论处理结果如何都回 "success"，
// 否则微信会重试推送（重复事件由 ApplySubscribeEvent 幂等兜底）。
func (c *OfficialCallbackController) event(ctx *gin.Context) {
	token := config.WeChatConf().OfficialCallbackToken
	if !wechat.VerifyOfficialSignature(token, ctx.Query("signature"), ctx.Query("timestamp"), ctx.Query("nonce")) {
		logrus.Warn("official callback event signature failed")
		ctx.String(403, "invalid signature")
		return
	}

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		logrus.WithError(err).Error("official callback read body failed")
		ctx.String(200, "success")
		return
	}
	msg, err := wechat.ParseOfficialEvent(body)
	if err != nil {
		logrus.WithError(err).Error("official callback parse event failed")
		ctx.String(200, "success")
		return
	}

	if msg.MsgType == "event" {
		c.applyEvent(msg)
		// 关注事件：同步被动回复一条带小程序卡片的欢迎语（不调主动发送接口，零资质依赖）。
		// FromUserName 回填收到的 ToUserName（即服务号原始ID，原样回填永不写错）。
		if msg.Event == wxmodel.OfficialEventSubscribe {
			ctx.Data(200, "application/xml; charset=utf-8",
				[]byte(buildSubscribeReplyXML(msg.FromUserName, msg.ToUserName, time.Now().Unix())))
			return
		}
	}
	ctx.String(200, "success")
}

// buildSubscribeReplyXML 构造关注事件的被动回复文本消息 XML（明文模式）。
// toUser=用户 openid，fromUser=服务号原始ID（回填收到的 ToUserName）。
// Content 内嵌微信小程序卡片 <a> 标签：data-miniprogram-appid/path 指向小程序，
// href 为不支持小程序跳转环境下的降级地址。
func buildSubscribeReplyXML(toUser, fromUser string, createTime int64) string {
	content := "欢迎关注常看看！共同守护家人的健康生活。\n" +
		"点击进入 👉 <a href=\"https://www.changkankan.com\" " +
		"data-miniprogram-appid=\"" + config.WeChatConf().AppId + "\" " +
		"data-miniprogram-path=\"pages/login/index\">【常看看小程序】</a>，开启您与家人的每日习惯连接。"
	return fmt.Sprintf(
		"<xml><ToUserName><![CDATA[%s]]></ToUserName>"+
			"<FromUserName><![CDATA[%s]]></FromUserName>"+
			"<CreateTime>%d</CreateTime>"+
			"<MsgType><![CDATA[text]]></MsgType>"+
			"<Content><![CDATA[%s]]></Content></xml>",
		toUser, fromUser, createTime, content)
}

// applyEvent 关注/取关事件落库（仅 subscribe/unsubscribe；缺 unionId 降级记日志）
func (c *OfficialCallbackController) applyEvent(msg *wxmodel.OfficialEventMsg) {
	var subscribed bool
	switch msg.Event {
	case wxmodel.OfficialEventSubscribe:
		subscribed = true
	case wxmodel.OfficialEventUnsubscribe:
		subscribed = false
	default:
		return // 非关注/取关事件，忽略
	}
	// 关注事件 XML 常态不带 <Unionid>（微信行为，与是否绑定开放平台无强关系），
	// 但 cgi-bin/user/info 对已关注用户必返 unionid → 缺失时按 openid 调接口补，避免整条事件被丢弃。
	unionId := msg.UnionId
	if unionId == "" {
		info, err := wechat.Client().GetOfficialUserInfo(msg.FromUserName)
		if err != nil {
			logrus.WithError(err).WithField("openid", msg.FromUserName).
				Warn("official event missing unionid and backfill via user/info failed, skip")
			return
		}
		unionId = info.UnionId
		if unionId == "" {
			logrus.WithField("openid", msg.FromUserName).
				Warn("official event missing unionid and user/info returned empty unionid, skip")
			return
		}
	}
	if err := service.GetFanService().ApplySubscribeEvent(unionId, msg.FromUserName, subscribed, time.Now().UnixMilli()); err != nil {
		logrus.WithError(err).WithField("unionId", unionId).Error("apply subscribe event failed")
	}
}
