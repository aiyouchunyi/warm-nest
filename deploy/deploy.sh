#!/usr/bin/env bash
# warm-nest 一键部署：本地构建镜像 → 传输 → 远端加载 → 替换容器 → 健康检查
#
# 用法：在项目根目录执行
#   bash deploy/deploy.sh
#
# 前置：本地已能 ssh 连通 SSH_HOST；远端已装 docker + 自建 MySQL（库 warm-nest / 账号 warmnest）
# 配置：见 deploy/deploy.conf

set -euo pipefail

# ── 定位项目根 + 读配置 ──
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CONF="$SCRIPT_DIR/deploy.conf"

if [[ ! -f "$CONF" ]]; then
  echo "❌ 找不到配置文件 $CONF" >&2
  exit 1
fi
# shellcheck disable=SC1090
source "$CONF"

SSH="ssh -p ${SSH_PORT} ${SSH_USER}@${SSH_HOST}"
TAR_NAME="${IMAGE_NAME}-${IMAGE_TAG}.tar.gz"
LOCAL_TAR="/tmp/${TAR_NAME}"
IMAGE_REF="${IMAGE_NAME}:${IMAGE_TAG}"

log()  { echo -e "\n\033[1;36m▶ $*\033[0m"; }
ok()   { echo -e "\033[1;32m✓ $*\033[0m"; }
fail() { echo -e "\033[1;31m✗ $*\033[0m" >&2; exit 1; }

cd "$PROJECT_ROOT"

# ── 0. 预检 ──
log "0/6 预检本地环境"
command -v docker >/dev/null || fail "本地未装 docker"
docker buildx version >/dev/null 2>&1 || fail "缺 docker buildx（跨架构构建需要）"
$SSH "command -v docker >/dev/null" || fail "远端未装 docker 或 SSH 不通"
ok "本地 docker/buildx 就绪，SSH 可达 ${SSH_HOST}"

# ── 1. 本地构建镜像（跨架构 → linux/amd64）──
# 本地 Mac 多为 arm64，服务器是 amd64，必须 --platform 指定，否则远端 exec format error
log "1/6 构建镜像 ${IMAGE_REF}（${PLATFORM}）"
docker buildx build \
  --platform "${PLATFORM}" \
  -t "${IMAGE_REF}" \
  --load \
  . || fail "镜像构建失败"
ok "镜像构建完成"

# ── 2. 导出并压缩镜像 ──
log "2/6 导出镜像到 ${LOCAL_TAR}"
docker save "${IMAGE_REF}" | gzip > "${LOCAL_TAR}" || fail "docker save 失败"
ok "镜像已导出（$(du -h "${LOCAL_TAR}" | cut -f1)）"

# ── 3. scp 传输到远端 ──
log "3/6 传输镜像到 ${SSH_HOST}:${REMOTE_TMP}"
$SSH "mkdir -p ${REMOTE_TMP}" || fail "远端创建目录失败"
scp -P "${SSH_PORT}" "${LOCAL_TAR}" "${SSH_USER}@${SSH_HOST}:${REMOTE_TMP}/${TAR_NAME}" || fail "scp 传输失败"
ok "传输完成"

# ── 4. 远端加载镜像 ──
log "4/6 远端加载镜像"
$SSH "gunzip -c ${REMOTE_TMP}/${TAR_NAME} | docker load" || fail "远端 docker load 失败"
ok "远端镜像加载完成"

# ── 5. 替换容器（停旧 → 删旧 → 起新）──
# 敏感值（DB 密码/JWT/微信 secret）写进主机 env 文件（权限 600），docker --env-file 注入。
# 不用命令行 -e：那样密码会出现在 docker inspect / ps，env-file 更干净。
log "5/6 替换容器 ${CONTAINER_NAME}"
ENV_FILE="${REMOTE_TMP}/warm-nest.env"
$SSH bash -s <<REMOTE
set -e
mkdir -p "${HOST_IMAGE_DIR}" "${REMOTE_TMP}"
# 生成主机 env 文件（敏感值；权限 600）。env-file 每行 KEY=VALUE，值不要加引号。
umask 077
cat > "${ENV_FILE}" <<ENVEOF
SERVER_MODE=${SERVER_MODE}
SERVER_IAMENABLE=${IAM_ENABLE}
DATABASE_MYSQL=${DB_DSN}
JWT_SECRET_KEY=${JWT_SECRET_KEY}
WECHAT_SECRET=${WECHAT_SECRET}
WECHAT_OFFICIAL_SECRET=${WECHAT_OFFICIAL_SECRET}
WECHAT_OFFICIAL_CALLBACK_TOKEN=${WECHAT_OFFICIAL_CALLBACK_TOKEN}
OSS_ENDPOINT=${OSS_ENDPOINT}
OSS_PUBLIC_ENDPOINT=${OSS_PUBLIC_ENDPOINT}
OSS_BUCKET=${OSS_BUCKET}
OSS_REGION=${OSS_REGION}
OSS_ACCESS_KEY_ID=${OSS_ACCESS_KEY_ID}
OSS_ACCESS_KEY_SECRET=${OSS_ACCESS_KEY_SECRET}
ENVEOF
# 停 + 删旧容器（不存在也不报错）
docker rm -f "${CONTAINER_NAME}" 2>/dev/null || true
# 起新容器
docker run -d \
  --name "${CONTAINER_NAME}" \
  --restart always \
  --memory "${MEM_LIMIT}" \
  -p "${HOST_PORT}:${CONTAINER_PORT}" \
  --env-file "${ENV_FILE}" \
  -v "${HOST_IMAGE_DIR}:${CONTAINER_IMAGE_DIR}" \
  --log-opt max-size=50m --log-opt max-file=3 \
  "${IMAGE_REF}"
REMOTE
ok "新容器已启动（敏感值经 --env-file 注入）"

# ── 6. 健康检查（轮询 /ping，最多 30s）──
log "6/6 健康检查 http://127.0.0.1:${HOST_PORT}/ping"
HEALTHY=0
for i in $(seq 1 15); do
  if $SSH "curl -sf -m 3 http://127.0.0.1:${HOST_PORT}/ping >/dev/null 2>&1"; then
    HEALTHY=1; break
  fi
  sleep 2
done

if [[ "$HEALTHY" == "1" ]]; then
  ok "服务健康（/ping 通）"
  echo
  $SSH "echo '── 容器状态 ──'; docker ps --filter name=${CONTAINER_NAME} --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'; echo '── 最近日志 ──'; docker logs --tail 15 ${CONTAINER_NAME}"
  echo
  ok "部署完成 → http://${SSH_HOST}:${HOST_PORT}/ping"
  # 清理远端 tar
  $SSH "rm -f ${REMOTE_TMP}/${TAR_NAME}" || true
  rm -f "${LOCAL_TAR}" || true
else
  fail "健康检查失败！服务未在 30s 内就绪。排查：$SSH \"docker logs --tail 50 ${CONTAINER_NAME}\""
fi
