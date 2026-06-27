// Package api @Author larry
// @Date 2026/06/15
// @Desc 消息接口聚合 + 列表/未读

package api

import (
	"sync"
	"time"

	"warm-nest/pkg/app/web/session"

	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
	"warm-nest/internal/service"
)

// ApiMessage 消息接口
type ApiMessage struct {
	messageMapper *mapper.MessageMapper
}

var apiMessage *ApiMessage
var apiMessageOnce sync.Once

// GetApiMessage 获取消息接口单例
func GetApiMessage() *ApiMessage {
	apiMessageOnce.Do(func() {
		apiMessage = &ApiMessage{messageMapper: mapper.GetMessageMapper()}
	})
	return apiMessage
}

// 消息列表分页默认值
const (
	defaultMessagePage     = 1
	defaultMessagePageSize = 20
	maxMessagePageSize     = 100
)

// ListMessageReq 消息列表请求（子女端，需登录）。
// Type 默认 CHECK_IN：列表只展示老人打卡消息，不含未打卡提醒/绑定成功（问题9）；
// 显式传其它类型可覆盖（前端一般固定传 CHECK_IN，留扩展）。
type ListMessageReq struct {
	session.Session
	Type     string `form:"type"`     // 不传=默认 CHECK_IN（打卡消息）；显式传值覆盖
	Page     int    `form:"page"`     // 页码，从 1 开始，默认 1
	PageSize int    `form:"pageSize"` // 每页条数，默认 20，上限 100
}

// MessageListResp 消息列表 + 分页 + 未读数（list 每项含关联打卡照片，问题2）
type MessageListResp struct {
	List     []service.MessageItem `json:"list"`
	Total    int64                 `json:"total"`    // 该筛选条件下总条数
	Page     int                   `json:"page"`     // 当前页
	PageSize int                   `json:"pageSize"` // 每页条数
	Unread   int64                 `json:"unread"`   // 未读总数（角标）
}

// List 分页查消息列表 + 未读数（默认只返打卡消息，每条带关联打卡照片）
func (a *ApiMessage) List(req ListMessageReq) (interface{}, error) {
	// 默认只看打卡消息：前端未传 type 时按 CHECK_IN 过滤
	msgType := req.Type
	if msgType == "" {
		msgType = model.MessageTypeCheckIn
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)

	list, total, unread, err := service.GetMessageService().
		ListReceiverMessages(req.ReqUser, msgType, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	return MessageListResp{List: list, Total: total, Page: page, PageSize: pageSize, Unread: unread}, nil
}

// normalizePage 归一化页码/页大小（兜底默认值 + 上限保护）
func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = defaultMessagePage
	}
	if pageSize < 1 {
		pageSize = defaultMessagePageSize
	}
	if pageSize > maxMessagePageSize {
		pageSize = maxMessagePageSize
	}
	return page, pageSize
}

// ReadMessageReq 标记消息已读请求（子女端，需登录）
type ReadMessageReq struct {
	session.Session
	MessageId string `json:"messageId" validate:"required"`
}

// MarkRead 标记一条消息为已读（已读同步给后端，CountUnread 角标才准）
func (a *ApiMessage) MarkRead(req ReadMessageReq) (interface{}, error) {
	affected, err := a.messageMapper.MarkRead(req.ReqUser, req.MessageId, time.Now().UnixMilli())
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"ok": true, "updated": affected}, nil
}
