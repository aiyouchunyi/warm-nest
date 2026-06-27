# Test Cases - api

## api

### ApiReward.List
> internal/api/api_reward.go | GET /warm-nest/v1/reward/list

领取记录列表（公用接口·模式A）：`List(req)` = `ResolveElder(viewer, elderUserId)` 解析归属 + 权限校验 → `ListByUser(elderUserId)` 按 `elder_user_id` 查、`created_at DESC` 倒序。
直接构造 `ListRewardReq`（`Session.ReqUser` = 登录人，`ElderUserId` = 入参）调用，跑通 ResolveElder + ListByUser 整条链路。
集成测试连本地 MySQL（连不上则全部 t.Skip）。涉及表：`t_guardianship`、`t_reward_claim`。

权限判据（ResolveElder 两分支）：
- 不传 elderUserId（或传自己）→ 老人查自己：viewer 必须是被守护人（作为 elder 有 ACTIVE 守护边），否则 ErrNotGuardian(10003001)。
- 传了 elderUserId（且非自己）→ 子女查老人：viewer 必须是该老人的 ACTIVE 守护人，否则 ErrNotGuardian(10003001)。

## Integration

- [x] IT-01 老人不传 elderUserId 查自己（自身有 ACTIVE 守护边、名下 2 条领取记录）, 返回该老人全部记录、按 created_at 倒序（新在前）
- [x] IT-02 子女传 elderUserId 查所守护老人（viewer 是该老人 ACTIVE 守护人、老人名下有记录）, 返回该老人记录、归属正确
- [x] IT-03 老人传自己 userId 作 elderUserId（elderUserId==viewer）, 走查自己分支、返回自己记录（等价 IT-01）
- [x] IT-04 子女不传 elderUserId（守护人误用：自身非被守护人、无 ACTIVE 守护边）, 拒绝、返回 ErrNotGuardian(10003001)、不静默返空
- [x] IT-05 非守护人传他人 elderUserId（viewer 与该老人无 ACTIVE 守护关系）, 拒绝、返回 ErrNotGuardian(10003001)
- [x] IT-06 守护关系为非 ACTIVE（如 INACTIVE/REVOKED）传该老人 elderUserId, 拒绝、返回 ErrNotGuardian(10003001)
- [x] IT-07 老人查自己但名下无任何领取记录, 返回空列表（非 nil error）、长度 0
- [x] IT-08 归属隔离：库内同时存在他人(elderB)记录，查 elderA, 仅返回 elderA 的记录、不串入 elderB 的
