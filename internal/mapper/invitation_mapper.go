// Package mapper @Author larry
// @Date 2026/06/15
// @Desc 守护邀请映射器

package mapper

import (
	"fmt"
	"sync"

	"warm-nest/pkg/tool/database/mysqls"

	"warm-nest/internal/model"
)

// InvitationMapper 守护邀请映射器
type InvitationMapper struct {
	mysqls.MysqlMapper[model.Invitation]
}

var invitationMapper *InvitationMapper
var invitationOnce sync.Once

// GetInvitationMapper 获取守护邀请映射器单例
func GetInvitationMapper() *InvitationMapper {
	invitationOnce.Do(func() {
		invitationMapper = &InvitationMapper{
			MysqlMapper: mysqls.NewMapper[model.Invitation](),
		}
	})
	return invitationMapper
}

// GetByInviteCode 按邀请码查邀请（接受入口，找不到返回 nil,nil）
func (m *InvitationMapper) GetByInviteCode(inviteCode string) (*model.Invitation, error) {
	var inv model.Invitation
	err := m.Model().Where("invite_code = ?", inviteCode).First(&inv).Error
	if err != nil {
		if mysqls.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get invitation by code %s: %w", inviteCode, err)
	}
	return &inv, nil
}

// GetByInvitationId 按业务邀请ID查邀请（撤销入口，找不到返回 nil,nil）
func (m *InvitationMapper) GetByInvitationId(invitationId string) (*model.Invitation, error) {
	var inv model.Invitation
	err := m.Model().Where("invitation_id = ?", invitationId).First(&inv).Error
	if err != nil {
		if mysqls.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get invitation by id %s: %w", invitationId, err)
	}
	return &inv, nil
}

// ListExpiredPending 查已过期但仍 PENDING 的邀请（过期扫描任务用）
func (m *InvitationMapper) ListExpiredPending(nowMs int64) ([]model.Invitation, error) {
	var list []model.Invitation
	err := m.Model().
		Where("status = ? AND expire_at < ?", model.InvitationStatusPending, nowMs).
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list expired pending invitations: %w", err)
	}
	return list, nil
}

// ListByGuardian 查某子女发起的全部邀请（按创建倒序）
func (m *InvitationMapper) ListByGuardian(guardianUserId string) ([]model.Invitation, error) {
	var list []model.Invitation
	err := m.Model().
		Where("guardian_user_id = ?", guardianUserId).
		Order("created_at DESC").
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list invitation by guardian %s: %w", guardianUserId, err)
	}
	return list, nil
}
