#!/usr/bin/env bash
# warm-nest 全场景测试数据构造脚本（一键清空 + 重建，供前端测试）
#
# 用法：bash deploy/seed_scenarios.sh [host]   host 默认 http://8.137.189.95:8080
# 幂等可重复跑：每次先 wipe-all 清空，再按固定 code 重建全部场景。
# 前端用脚本输出的「登录 code」调 POST /user/login 即进入对应账号（mock 登录，同 code 同账号）。
#
# 依赖的测试通道接口（无鉴权，仅测试期）：
#   POST /warm-nest/test/v1/wipe-all                 全清
#   POST /warm-nest/test/v1/reward/seed-tasks        初始化奖励规则
#   POST /warm-nest/test/v1/guardianship/seed        造守护关系+档案（带结构化地址）
#   POST /warm-nest/test/v1/checkin/seed             造连续打卡（自动触发奖励评估→生成可领取记录）
#   POST /warm-nest/v1/user/login                    mock 登录拿 userId（code=test_xxx → 固定 userId）
#   POST /warm-nest/v1/family/profile                改档案（测缺收货信息场景）
#   POST /warm-nest/v1/checkin/do                    真实打卡（触发给子女发带照片消息）

set -euo pipefail
HOST="${1:-http://8.137.189.95:8080}"
APP_HEADER="X-App-Id: seed"
CT="Content-Type: application/json"

# 登录拿 userId（mock：同 code 同 userId，幂等）
login() {
  curl -s -X POST "$HOST/warm-nest/v1/user/login" -H "$CT" -d "{\"code\":\"$1\"}" \
    | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['userId'])"
}
# 调用并打印结果摘要
post() { curl -s -X POST "$HOST$1" -H "$APP_HEADER" -H "$CT" -d "${2:-{}}"; }

echo "▶ 目标: $HOST"
echo "▶ 1/6 清空所有数据..."
post /warm-nest/test/v1/wipe-all >/dev/null
echo "  ✓ 已清空"

echo "▶ 2/6 初始化奖励规则（monthly_egg + continuous_egg_7）..."
post /warm-nest/test/v1/reward/seed-tasks >/dev/null
echo "  ✓ 已初始化"

echo "▶ 3/6 登录建号（mock，固定 code→固定 userId）..."
ELDER_FULL=$(login test_elder_full)         # 场景A 老人：完整数据
GUARD_FULL=$(login test_guardian_full)       # 场景A 子女：守护 elder_full
ELDER_NOADDR=$(login test_elder_noaddr)      # 场景B 老人：缺收货信息
ELDER_NEW=$(login test_elder_new)            # 场景C 老人：已绑定无打卡
GUARD_PENDING=$(login test_guardian_pending) # 场景D 子女：仅发邀请未被接受
echo "  elder_full=$ELDER_FULL guardian_full=$GUARD_FULL"
echo "  elder_noaddr=$ELDER_NOADDR elder_new=$ELDER_NEW guardian_pending=$GUARD_PENDING"

echo "▶ 4/6 建守护关系 + 老人档案（结构化地址）..."
# 场景A：完整老人，子女守护，档案带完整收货地址
post /warm-nest/test/v1/guardianship/seed "{\"guardianUserId\":\"$GUARD_FULL\",\"elderUserId\":\"$ELDER_FULL\",\"relation\":\"MOM\"}" >/dev/null
# 场景B：缺地址老人（先建关系+默认档案，下一步改档案清空收货人/电话）
post /warm-nest/test/v1/guardianship/seed "{\"guardianUserId\":\"$GUARD_FULL\",\"elderUserId\":\"$ELDER_NOADDR\",\"relation\":\"DAD\"}" >/dev/null
# 场景C：新老人，已绑定但无打卡
post /warm-nest/test/v1/guardianship/seed "{\"guardianUserId\":\"$GUARD_FULL\",\"elderUserId\":\"$ELDER_NEW\",\"relation\":\"GRANDMA\"}" >/dev/null
echo "  ✓ 3 条守护关系（guardian_full 守护 3 位老人，测一子女多老人）"

echo "▶ 5/6 造打卡数据（自动触发奖励评估→生成可领取记录）..."
# 场景A：连续 8 天打卡 → 命中 continuous_egg_7 → 生成 PENDING 领取记录（可测领取）
post /warm-nest/test/v1/checkin/seed "{\"elderUserId\":\"$ELDER_FULL\",\"days\":8}" >/dev/null
# 场景B：连续 8 天打卡 → 也有可领取记录（但档案缺收货人，测领取错误码）
post /warm-nest/test/v1/checkin/seed "{\"elderUserId\":\"$ELDER_NOADDR\",\"days\":8}" >/dev/null
# 场景B 改档案：清空收货人/电话（只留省市详细），测领取 10002002/10002003
post /warm-nest/v1/family/profile -H "$APP_HEADER" >/dev/null 2>&1 || true
curl -s -X POST "$HOST/warm-nest/v1/family/profile" -H "$APP_HEADER" -H "X-User-Id: $ELDER_NOADDR" -H "$CT" \
  -d "{\"elderUserId\":\"$ELDER_NOADDR\",\"realName\":\"缺地址老人\",\"remindTime\":\"09:00\",\"address\":{\"province\":\"上海市\",\"city\":\"上海市\",\"district\":\"浦东新区\",\"detail\":\"测试路1号\",\"receiverName\":\"\",\"receiverPhone\":\"\"}}" >/dev/null
echo "  ✓ elder_full/elder_noaddr 各 8 天打卡；elder_new 无打卡"

echo "▶ 6/6 真实打卡（触发给子女发带照片消息）+ 演示邀请流程..."
# 场景A 老人真实打卡今天（触发给 guardian_full 发 CHECK_IN 消息，带照片，测 message/list）
curl -s -X POST "$HOST/warm-nest/v1/checkin/do" -H "$APP_HEADER" -H "X-User-Id: $ELDER_FULL" -H "$CT" \
  -d "{\"photoUrl\":\"https://cdn.example.com/checkin/today.jpg\",\"weather\":\"晴 26°C\",\"city\":\"上海\"}" >/dev/null
# 场景D：guardian_pending 真实发起邀请（PENDING，未被接受）——测「邀请已发待接受」状态
INVITE=$(curl -s -X POST "$HOST/warm-nest/v1/invitation/create" -H "$APP_HEADER" -H "X-User-Id: $GUARD_PENDING" -H "$CT" \
  -d "{\"elderPhone\":\"13800138000\",\"relation\":\"DAD\",\"remindTime\":\"08:00\",\"city\":\"北京\"}" \
  | python3 -c "import sys,json;d=json.load(sys.stdin).get('data',{});print(d.get('inviteCode',''))" 2>/dev/null || echo "")
echo "  ✓ elder_full 今日已打卡(有消息)；guardian_pending 发起邀请 inviteCode=$INVITE"

cat <<EOF

════════════════════════════════════════════════════
✅ 全场景数据已重建。前端账号速查（用 code 调 /user/login）：
  场景A 老人(完整)   code=test_elder_full      userId=$ELDER_FULL
  场景A 子女(守护)   code=test_guardian_full   userId=$GUARD_FULL
  场景B 老人(缺地址) code=test_elder_noaddr    userId=$ELDER_NOADDR
  场景C 老人(无打卡) code=test_elder_new       userId=$ELDER_NEW
  场景D 子女(待接受) code=test_guardian_pending userId=$GUARD_PENDING
  场景D 邀请码 inviteCode=$INVITE
════════════════════════════════════════════════════
EOF
