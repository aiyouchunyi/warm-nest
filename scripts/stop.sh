#!/usr/bin/env bash
# 停止 warm-nest 后台服务
set -euo pipefail

cd "$(dirname "$0")/.."
APP="warm-nest"
PID_FILE="./runtime/${APP}.pid"

if [ ! -f "$PID_FILE" ]; then
  echo "[stop] 未找到 PID 文件，服务可能未启动"
  exit 0
fi

PID="$(cat "$PID_FILE")"
if ! kill -0 "$PID" 2>/dev/null; then
  echo "[stop] 进程 $PID 不存在，清理 PID 文件"
  rm -f "$PID_FILE"
  exit 0
fi

echo "[stop] 停止 $APP (PID $PID) ..."
kill "$PID"
# 等待最多 10s 优雅退出，超时强杀
for _ in $(seq 1 10); do
  kill -0 "$PID" 2>/dev/null || break
  sleep 1
done
if kill -0 "$PID" 2>/dev/null; then
  echo "[stop] 优雅退出超时，强制 kill -9"
  kill -9 "$PID" 2>/dev/null || true
fi
rm -f "$PID_FILE"
echo "[stop] 已停止"
