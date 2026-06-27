// Package task @Author larry
// @Date 2026/06/15
// @Desc 邀请过期扫描（PENDING 且超时置 EXPIRED，避免列表展示态失真）

package task

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"warm-nest/internal/service"
)

// InvitationExpireTask 邀请过期扫描任务
type InvitationExpireTask struct{}

var invitationExpireTask *InvitationExpireTask
var invitationExpireTaskOnce sync.Once

// GetInvitationExpireTask 获取邀请过期扫描任务单例
func GetInvitationExpireTask() *InvitationExpireTask {
	invitationExpireTaskOnce.Do(func() {
		invitationExpireTask = &InvitationExpireTask{}
	})
	return invitationExpireTask
}

// Scan 扫描并置过期
func (t *InvitationExpireTask) Scan(args ...any) error {
	n, err := service.GetInvitationService().ExpireOverdue(time.Now().UnixMilli())
	if err != nil {
		return err
	}
	if n > 0 {
		logrus.WithField("count", n).Info("expired overdue invitations")
	}
	return nil
}
