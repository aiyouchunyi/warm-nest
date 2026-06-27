#!/usr/bin/env bash
# 在宿主机安装 warm-nest 日志归档（systemd 采集 + cron 清理 30 天）。幂等可重跑。
set -euo pipefail

LOG_DIR="/data/warm-nest/logs"
BIN="/usr/local/bin/warm-nest-log-archive.sh"
UNIT="/etc/systemd/system/warm-nest-log-archive.service"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "▶ 安装采集脚本 + unit"
install -m 0755 "$SCRIPT_DIR/warm-nest-log-archive.sh" "$BIN"
install -m 0644 "$SCRIPT_DIR/warm-nest-log-archive.service" "$UNIT"
mkdir -p "$LOG_DIR"

echo "▶ 启用 systemd 服务"
systemctl daemon-reload
systemctl enable --now warm-nest-log-archive.service

echo "▶ 安装 cron 清理（保留 30 天，每天 03:10 删更早文件）"
CRON_LINE="10 3 * * * find $LOG_DIR -name 'warm-nest-*.log' -mtime +30 -delete"
# 幂等：先去掉旧的同类行再加。
# 注意：crontab -l 在「无现有 crontab」时退出码非 0，配合 set -e 会中断脚本；
# grep 在无匹配时也返回非 0。故这段整体用 || true 兜底，且 crontab -l 失败按空 crontab 处理。
{ crontab -l 2>/dev/null || true; } | grep -vF "$LOG_DIR -name 'warm-nest-*.log'" > /tmp/wn-cron.tmp || true
echo "$CRON_LINE" >> /tmp/wn-cron.tmp
crontab /tmp/wn-cron.tmp
rm -f /tmp/wn-cron.tmp

echo "▶ 给 docker json.log 加上限，防撑爆磁盘（兜底，需 deploy.sh 重建容器后对新容器生效）"
echo "  注：当前运行容器不变，下次 deploy 起的新容器才带 --log-opt（见 deploy.sh 改动）"

echo "✓ 安装完成。状态："
systemctl --no-pager status warm-nest-log-archive.service | head -6
echo "--- 日志目录 ---"
ls -lh "$LOG_DIR" 2>&1
echo "--- cron ---"
crontab -l | grep warm-nest
