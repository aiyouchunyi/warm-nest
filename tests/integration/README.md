# 集成测试脚本（本地真实链路回归）

> 依赖真实运行环境（本地 MySQL + dev 模式服务），非 unit test。
> 纯逻辑单测（连续天数算法等）见 `internal/reward/condition/streak_test.go`（go test）。

## 前置

1. 本地 MySQL 跑着，`warm-nest` 库存在（首次 dev 启动自动建表）
2. dev 模式启动服务（token 鉴权生效、微信走 mock）：
   ```bash
   ./scripts/start.sh dev
   ```

## 脚本

| 脚本 | 覆盖 |
|---|---|
| `full_api_smoke.py` | 全部 C 端接口冒烟：登录/邀请两阶段/打卡/今日/月份/消息/家庭双端视角/提醒/奖励列表 |
| `reward_state_machine_test.py` | 奖励状态机：PENDING→CLAIMED→SHIPPED→SIGNED 流转 + machine 按当前状态路由 + 终态拒绝保护 |
| `upload_static_test.py` | 图片 multipart 上传 + static 路由读取闭环 |

## 跑法

```bash
python3 tests/integration/full_api_smoke.py
python3 tests/integration/reward_state_machine_test.py
python3 tests/integration/upload_static_test.py
```

> 状态机测试如需重跑，先清测试数据：`DELETE FROM t_check_in; DELETE FROM t_reward_claim;`
> 奖励达成需先有规则：`INSERT INTO t_reward_task(...) VALUES('cum_egg',...,'CUMULATIVE_CHECK_IN',1,...)`
