#!/usr/bin/env bash
# warm-nest 容器日志归档：docker logs -f 持续跟随 → 剥 ANSI 色码 → 按天写文件。
# 文件名即当天日期 warm-nest-YYYYMMDD.log（按天滚动天然无 copytruncate 竞态）。
# 由 systemd 常驻拉起；容器被 deploy 删除重建时本服务自动重连（--retry 轮询）。
set -uo pipefail

CONTAINER="${CONTAINER_NAME:-warm-nest}"
LOG_DIR="${LOG_DIR:-/data/warm-nest/logs}"

mkdir -p "$LOG_DIR"

# 容器可能正在 deploy 重建：循环重连，不退出（systemd 也会兜底重启）。
while true; do
  if ! docker inspect "$CONTAINER" >/dev/null 2>&1; then
    sleep 3
    continue
  fi
  # --since 0 不取历史（避免重连时重复落旧行）；-f 持续跟随当前容器生命周期。
  # awk 按当前日期分流写文件：strftime 每行算当天，跨零点自动切到新文件。
  # 剥 ANSI 色码后落盘（grep 检索友好）。
  docker logs -f --since "$(date '+%Y-%m-%dT%H:%M:%S')" "$CONTAINER" 2>&1 \
    | sed -ur 's/\x1b\[[0-9;]*m//g' \
    | awk -v dir="$LOG_DIR" '{
        f = dir "/warm-nest-" strftime("%Y%m%d") ".log";
        print >> f;
        fflush(f);
      }'
  # docker logs -f 在容器停止/删除时会返回，循环重连下一个容器实例。
  sleep 3
done
