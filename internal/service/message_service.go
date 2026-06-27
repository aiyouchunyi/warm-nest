// Package service @Author larry
// @Date 2026/06/15
// @Desc 消息服务（双通道：小程序内消息流必达 + 微信订阅消息 best-effort）

package service

import (
	"fmt"
	"maps"
	"sync"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/kinds/jsons"
	"warm-nest/pkg/utils/rands"

	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
	"warm-nest/internal/notify"
)

// MessageService 消息服务
type MessageService struct {
	messageMapper      *mapper.MessageMapper
	guardianshipMapper *mapper.GuardianshipMapper
	checkInMapper      *mapper.CheckInMapper
	elderProfileMapper *mapper.ElderProfileMapper
}

var messageService *MessageService
var messageServiceOnce sync.Once

// GetMessageService 获取消息服务单例
func GetMessageService() *MessageService {
	messageServiceOnce.Do(func() {
		messageService = &MessageService{
			messageMapper:      mapper.GetMessageMapper(),
			guardianshipMapper: mapper.GetGuardianshipMapper(),
			checkInMapper:      mapper.GetCheckInMapper(),
			elderProfileMapper: mapper.GetElderProfileMapper(),
		}
	})
	return messageService
}

// withParam 在原 params 基础上克隆一份并覆盖一个 key，返回新 map。
// 多子女循环里逐接收方注入 relation 必须克隆——共享同一 map 会让所有人的内消息流/外推串成最后一个值。
func withParam(params jsons.JSONObject, key string, val any) jsons.JSONObject {
	cloned := make(jsons.JSONObject, len(params)+1)
	maps.Copy(cloned, params)
	cloned[key] = val
	return cloned
}

// MessageItem 消息列表项：消息本体 + 关联打卡照片（问题2，CHECK_IN 消息才有）
type MessageItem struct {
	model.Message
	PhotoUrl    string `json:"photoUrl"`    // 关联打卡照片（非打卡消息或打卡不存在为空）
	Weather     string `json:"weather"`     // 关联打卡天气
	CheckInCity string `json:"checkInCity"` // 关联打卡城市
	CheckInDate string `json:"checkInDate"` // 关联打卡归属日
}

// ListReceiverMessages 分页查某子女收到的消息，并为 CHECK_IN 消息批量关联打卡照片（问题2）。
// 按本页消息的 RefCheckInId 一次 IN 查询带回 CheckIn，避免逐条 N+1。
// 返回当页 items + 总数 + 未读数。
func (s *MessageService) ListReceiverMessages(receiverUserId, msgType string, offset, limit int) ([]MessageItem, int64, int64, error) {
	list, total, err := s.messageMapper.ListByReceiver(receiverUserId, msgType, offset, limit)
	if err != nil {
		return nil, 0, 0, err
	}
	unread, err := s.messageMapper.CountUnread(receiverUserId)
	if err != nil {
		return nil, 0, 0, err
	}

	// 收集本页 RefCheckInId，批量查打卡（IN 查询，非 N+1）
	ids := make([]string, 0, len(list))
	for i := range list {
		if list[i].RefCheckInId != "" {
			ids = append(ids, list[i].RefCheckInId)
		}
	}
	checkIns, err := s.checkInMapper.ListByCheckInIds(ids)
	if err != nil {
		return nil, 0, 0, err
	}
	byId := make(map[string]model.CheckIn, len(checkIns))
	for i := range checkIns {
		byId[checkIns[i].CheckInId] = checkIns[i]
	}

	items := make([]MessageItem, 0, len(list))
	for i := range list {
		item := MessageItem{Message: list[i]}
		if c, ok := byId[list[i].RefCheckInId]; ok {
			item.PhotoUrl = c.PhotoUrl
			item.Weather = c.Weather
			item.CheckInCity = c.City
			item.CheckInDate = c.CheckInDate
		}
		items = append(items, item)
	}
	return items, total, unread, nil
}

// SendCheckInNotice 对老人的所有 ACTIVE 守护子女发打卡通知（内消息流 + 订阅推送）
func (s *MessageService) SendCheckInNotice(checkIn *model.CheckIn, params jsons.JSONObject) error {
	guards, err := s.guardianshipMapper.ListByElder(checkIn.ElderUserId)
	if err != nil {
		return err
	}
	for i := range guards {
		receiverId := guards[i].GuardianUserId
		dedupKey := fmt.Sprintf("checkin:%s:%s", checkIn.CheckInId, receiverId)
		// 逐子女注入其对老人的称呼（模板「用户名」字段）——每子女称呼不同，须克隆 params 防串值
		p := withParam(params, "relation", guards[i].Relation)
		s.deliver(model.MessageTypeCheckIn, model.NotifySceneCheckInNotice,
			receiverId, checkIn.ElderUserId, checkIn.CheckInId, dedupKey, "", p)
	}
	return nil
}

// SendNotRemind 对老人的所有 ACTIVE 守护子女发未打卡提醒
func (s *MessageService) SendNotRemind(elderUserId, date string, params jsons.JSONObject) error {
	guards, err := s.guardianshipMapper.ListByElder(elderUserId)
	if err != nil {
		return err
	}
	for i := range guards {
		receiverId := guards[i].GuardianUserId
		dedupKey := fmt.Sprintf("remind:%s:%s:%s", elderUserId, receiverId, date)
		// 逐子女注入称呼（模板「姓名」字段）——克隆防串值
		p := withParam(params, "relation", guards[i].Relation)
		s.deliver(model.MessageTypeNotRemind, model.NotifySceneNotRemindGuardian,
			receiverId, elderUserId, "", dedupKey, date, p)
	}
	return nil
}

// SendBindSuccess 绑定成立后向子女发「✅ 已成功绑定 XX」反馈（PRD §8.0.1.6，best-effort）
// guardianshipId 入 dedupKey 保证一次绑定只通知一次；params 由调用方备好（含老人称呼等）
func (s *MessageService) SendBindSuccess(guardianUserId, elderUserId, guardianshipId string, params jsons.JSONObject) {
	dedupKey := fmt.Sprintf("bind:%s", guardianshipId)
	s.deliver(model.MessageTypeBindSuccess, model.NotifySceneBindSuccess,
		guardianUserId, elderUserId, "", dedupKey, "", params)
}

// SendElderSelfRemind 未打卡提醒第一段：直接推老人本人（PRD §8.3，best-effort）
// 收件人即老人自己；dedupKey 区别于推子女的 remind，避免与子女提醒互相幂等顶掉
func (s *MessageService) SendElderSelfRemind(elderUserId, date string, params jsons.JSONObject) {
	dedupKey := fmt.Sprintf("self-remind:%s:%s", elderUserId, date)
	// 推老人本人，无「子女称呼」可用 → 注入老人真名供模板「姓名」字段兜底（profile 缺失时翻译层再兜「家人」）
	p := params
	if profile, err := s.elderProfileMapper.GetByUserId(elderUserId); err == nil && profile != nil && profile.RealName != "" {
		p = withParam(params, "elderName", profile.RealName)
	}
	s.deliver(model.MessageTypeNotRemind, model.NotifySceneNotRemindElder,
		elderUserId, elderUserId, "", dedupKey, date, p)
}

// SendAddressPreheat 奖励地址预热提醒：向某子女发一条「请提前补收货地址」（PRD §6.6.3，best-effort）。
// 收件人=子女；按月幂等(period=YYYY-MM)，25号任务当月重跑/多次扫到同一子女只发一次。
func (s *MessageService) SendAddressPreheat(guardianUserId, elderUserId, period string, params jsons.JSONObject) {
	dedupKey := fmt.Sprintf("preheat:%s:%s:%s", elderUserId, guardianUserId, period)
	s.deliver(model.MessageTypeAddressPreheat, model.NotifySceneAddressPreheat,
		guardianUserId, elderUserId, "", dedupKey, "", params)
}

// MarkRemindChecked 老人当日补打卡后，把当日已发出的未打卡提醒消息流标记「已打卡」（§5.3）。
// 提醒不撤回，仅置 elder_checked_at，让消息流不再显示为有效待处理。返回标记条数(0 正常)。
func (s *MessageService) MarkRemindChecked(elderUserId, bizDate string, nowMs int64) (int64, error) {
	return s.messageMapper.MarkElderChecked(elderUserId, bizDate, nowMs)
}

// deliver 单个子女的双通道投递：先落内消息流（幂等），再 best-effort 按场景外推。
// scene 决定外推走哪个渠道（小程序订阅/服务号模板/…），由 notify 包按 notify_route 表解析，
// 上层不认渠道（PRD §5.2）。msgType 是内消息流的展示类型，与外推渠道解耦。
// bizDate 为消息业务归属日（YYYY-MM-DD），仅未打卡提醒需要（供补打卡标记定位）；其余消息传 ""。
func (s *MessageService) deliver(msgType, scene, receiverId, elderUserId, refCheckInId, dedupKey, bizDate string,
	params jsons.JSONObject) {

	log := logrus.WithFields(logrus.Fields{"dedupKey": dedupKey, "type": msgType, "scene": scene})

	// 通道① 内消息流（DedupKey 唯一索引幂等，必达）
	exist, err := s.messageMapper.ExistByDedupKey(dedupKey)
	if err != nil {
		log.WithError(err).Error("check message dedup failed")
		return
	}
	if exist {
		// 已投递过（任务重跑 / 未关注老人第一段已推子女、第二段又扫到）→ 整体跳过，
		// 外推也不重发，避免对同一 dedupKey 重复骚扰。
		return
	}
	msg := &model.Message{
		MessageId:      rands.Numeric(),
		DedupKey:       dedupKey,
		ReceiverUserId: receiverId,
		ElderUserId:    elderUserId,
		Type:           msgType,
		Params:         params,
		RefCheckInId:   refCheckInId,
		BizDate:        bizDate,
	}
	if err = s.messageMapper.Create(msg); err != nil {
		log.WithError(err).Error("create in-app message failed")
		return
	}

	// 通道② 按场景外推（best-effort，失败不阻断内消息流必达）
	if err = notify.Dispatch(scene, receiverId, params); err != nil {
		log.WithError(err).Warn("notify dispatch failed")
	}
}
