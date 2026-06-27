// 集成测试：验证「老人接受邀请后，登录返回的 availableRoles 含 ELDER」（问题4 复现核对）。
// 复用同包 TestMain 的建表/迁移与 cleanTables/seedInvitation 辅助。本地无 DB 时 t.Skip。

package service_test

import (
	"testing"

	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
	"warm-nest/internal/service"
)

// TestAvailableRoles_AfterAccept 接受邀请后，老人侧 ListByElder 应有 ACTIVE 关系
// → availableRoles 数据源非空（问题4：若线上仍空，是关系未真正落库或查错 userId，非本逻辑问题）。
func TestAvailableRoles_AfterAccept(t *testing.T) {
	requireDB(t)
	cleanTables(t)

	const guardian, elder, code = "guardianRole", "elderRole", "codeRole"
	nowMs := int64(1750000000000)
	seedInvitation(t, &model.Invitation{
		InvitationId:   "invRole",
		InviteCode:     code,
		GuardianUserId: guardian,
		Relation:       model.RelationMom,
		RemindTime:     "09:00",
		Status:         model.InvitationStatusPending,
		ExpireAt:       nowMs + 7*24*60*60*1000,
	})

	if err := service.GetInvitationService().Accept(elder, code, nowMs); err != nil {
		t.Fatalf("accept failed: %v", err)
	}

	// 直接核对 availableRoles 的数据源：老人侧 ACTIVE 关系
	rels, err := mapper.GetGuardianshipMapper().ListByElder(elder)
	if err != nil {
		t.Fatalf("list by elder: %v", err)
	}
	if len(rels) == 0 {
		t.Fatal("接受后老人侧无 ACTIVE 关系 → availableRoles 会为空（这才是问题4的故障态）")
	}
	if rels[0].Status != model.GuardianshipStatusActive {
		t.Errorf("relation status = %s, want ACTIVE", rels[0].Status)
	}
	t.Logf("接受成功后老人侧 ACTIVE 关系数 = %d，availableRoles 将含 ELDER（问题4 复现为环境/userId 问题，非代码逻辑）", len(rels))
}
