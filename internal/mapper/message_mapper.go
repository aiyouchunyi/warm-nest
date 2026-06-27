// Package mapper @Author larry
// @Date 2026/06/15
// @Desc 小程序消息映射器

package mapper

import (
	"fmt"
	"sync"

	"warm-nest/pkg/tool/database/mysqls"

	"warm-nest/internal/model"
)

// MessageMapper 小程序消息映射器
type MessageMapper struct {
	mysqls.MysqlMapper[model.Message]
}

var messageMapper *MessageMapper
var messageOnce sync.Once

// GetMessageMapper 获取小程序消息映射器单例
func GetMessageMapper() *MessageMapper {
	messageOnce.Do(func() {
		messageMapper = &MessageMapper{
			MysqlMapper: mysqls.NewMapper[model.Message](),
		}
	})
	return messageMapper
}

// ExistByDedupKey 判幂等键是否已存在（消息防重落库）
// 复用框架 ExistByUColumn（内部 CamelToUnderline，dedupKey→dedup_key），dedup_key 为唯一索引
func (m *MessageMapper) ExistByDedupKey(dedupKey string) (bool, error) {
	return m.ExistByUColumn("dedupKey", dedupKey)
}

// ListByReceiver 分页查某子女收到的消息（按创建倒序，可选只看某类型）。
// 返回当页列表 + 总数（前端算分页）。offset/limit 由调用方按页码算好。
func (m *MessageMapper) ListByReceiver(receiverUserId, msgType string, offset, limit int) ([]model.Message, int64, error) {
	base := m.Model().Where("receiver_user_id = ?", receiverUserId)
	if msgType != "" {
		base = base.Where("type = ?", msgType)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count message by receiver %s: %w", receiverUserId, err)
	}
	var list []model.Message
	err := base.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list message by receiver %s: %w", receiverUserId, err)
	}
	return list, total, nil
}

// MarkRead 把某子女的一条消息标记为已读（幂等：已读再标不报错；只更新归属本人的消息）。
// 返回受影响行数：0 表示该消息不存在或不属于该用户（前端可据此提示）。
func (m *MessageMapper) MarkRead(receiverUserId, messageId string, nowMs int64) (int64, error) {
	res := m.Model().
		Where("message_id = ? AND receiver_user_id = ? AND is_read = ?", messageId, receiverUserId, false).
		Updates(map[string]interface{}{"is_read": true, "read_at": nowMs})
	if res.Error != nil {
		return 0, fmt.Errorf("mark message read %s: %w", messageId, res.Error)
	}
	return res.RowsAffected, nil
}

// MarkElderChecked 老人当日补打卡后，把该老人当日已发出的未打卡提醒(NOT_REMIND)标记「已打卡」（§5.3）。
// 按 (elder_user_id, biz_date, type) 批量更新，仅标尚未标记的(elder_checked_at=0)保证幂等;不撤回消息、仅置标记时刻。
// 返回受影响行数（0 表示当日无已发提醒，正常）。
func (m *MessageMapper) MarkElderChecked(elderUserId, bizDate string, nowMs int64) (int64, error) {
	res := m.Model().
		Where("elder_user_id = ? AND biz_date = ? AND type = ? AND elder_checked_at = ?",
			elderUserId, bizDate, model.MessageTypeNotRemind, 0).
		Update("elder_checked_at", nowMs)
	if res.Error != nil {
		return 0, fmt.Errorf("mark elder checked %s %s: %w", elderUserId, bizDate, res.Error)
	}
	return res.RowsAffected, nil
}

// CountUnread 查某子女未读消息数（角标）
func (m *MessageMapper) CountUnread(receiverUserId string) (int64, error) {
	var cnt int64
	err := m.Model().
		Where("receiver_user_id = ? AND is_read = ?", receiverUserId, false).
		Count(&cnt).Error
	if err != nil {
		return 0, fmt.Errorf("count unread message by %s: %w", receiverUserId, err)
	}
	return cnt, nil
}
