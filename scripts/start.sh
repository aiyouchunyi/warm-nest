#!/usr/bin/env bash
# 启动 warm-nest 后台服务
# 用法: ./scripts/start.sh [mode]
#   mode 缺省 local（加载 app_local.toml，鉴权 widget 放行）；
#   传 dev/prod 则经 SERVER_MODE 环境变量生效（加载 app_{mode}.toml，token 真校验）。
set -euo pipefail

cd "$(dirname "$0")/.."
APP="warm-nest"
MODE="${1:-}"
BIN="./bin/${APP}"
PID_FILE="./runtime/${APP}.pid"
LOG_FILE="./runtime/${APP}.log"

mkdir -p ./runtime ./bin

# 已在运行则拒绝重复启动
if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
  echo "[start] 已在运行 (PID $(cat "$PID_FILE"))，请先 ./scripts/stop.sh"
  exit 1
fi

# 先构建
echo "[start] 构建 $APP ..."
go build -o "$BIN" . || { echo "[start] 构建失败"; exit 1; }

echo "[start] 启动 $APP ${MODE:+(mode=$MODE)} ..."
if [ -n "$MODE" ]; then
  SERVER_MODE="$MODE" nohup "$BIN" >"$LOG_FILE" 2>&1 &
else
  nohup "$BIN" >"$LOG_FILE" 2>&1 &
fi
echo $! >"$PID_FILE"
sleep 1

if kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
  echo "[start] 已启动 PID=$(cat "$PID_FILE")  日志: $LOG_FILE"
else
  echo "[start] 启动失败，查看日志: $LOG_FILE"
  tail -n 30 "$LOG_FILE" || true
  rm -f "$PID_FILE"
  exit 1
fi
