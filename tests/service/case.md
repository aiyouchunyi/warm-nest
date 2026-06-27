# Test Cases - service
> 2026-06-18 14:04:25

## service

### InvitationService.Accept
> internal/service/invitation_service.go | 2026-06-18 14:04:25

老人凭 inviteCode 接受邀请、建立守护关系。本轮新增：一次性失效、一对一守卫 V1、ActivatedAt 落库、绑定成功反馈。
集成测试连本地 MySQL（连不上则全部 t.Skip）。

## Integration

- [x] IT-01 首次接受（PENDING 邀请、双方均未绑）, 建 t_elder_profile + t_guardianship(status=ACTIVE、relation 取邀请预设、ActivatedAt==nowMs) + t_invitation 置 ACCEPTED 回填 AcceptedElderUserId/AcceptedAt
- [x] IT-02 同一老人重复接受已 ACCEPTED 的码, 幂等返回 nil、不新增关系
- [x] IT-03 另一老人扫到已 ACCEPTED 的码（卡片被转发）, 拒绝、返回 ErrInviteUsed(10001001)、不建关系
- [x] IT-04 老人已被别的子女 ACTIVE 绑定、再接受新子女的邀请, 拒绝、返回 ErrElderBound(10001002)
- [x] IT-05 子女已绑别的老人、该子女的邀请被新老人接受, 拒绝、返回 ErrGuardianBound(10001003)
- [x] IT-06 接受 CANCELLED 状态的邀请, 拒绝、返回 ErrInviteState(10001004)
- [x] IT-07 接受 EXPIRED 状态的邀请, 拒绝、返回 ErrInviteState(10001004)
- [x] IT-08 接受仍 PENDING 但已超时(nowMs>ExpireAt)的邀请, 拒绝、返回 ErrInviteState(10001004)
- [x] IT-09 inviteCode 不存在, 返回错误(invitation not found)
- [x] IT-10 该子女-老人对已存在 ACTIVE 关系、重复接受, 幂等置邀请 ACCEPTED、不新建第二条关系
- [x] IT-11 首次接受且老人已有 ElderProfile（仅缺关系）, 不重复建档案、建关系成功、ActivatedAt==nowMs
