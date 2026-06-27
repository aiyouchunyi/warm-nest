#!/usr/bin/env bash
# 查看 warm-nest 服务状态
set -euo pipefail

cd "$(dirname "$0")/.."
APP="warm-nest"
PID_FILE="./runtime/${APP}.pid"
LOG_FILE="./runtime/${APP}.log"

if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
  PID="$(cat "$PID_FILE")"
  echo "[status] 运行中  PID=$PID"
  ps -o pid,etime,rss,command -p "$PID" 2>/dev/null | tail -n +1 || true
else
  echo "[status] 未运行"
fi

if [ -f "$LOG_FILE" ]; then
  echo "[status] 日志尾部 ($LOG_FILE):"
  tail -n 15 "$LOG_FILE" || true
fi
