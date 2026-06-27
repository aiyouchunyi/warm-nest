#!/usr/bin/env bash
# 打包 warm-nest：编译 Linux 二进制 + 配置，输出到 dist/
# 用法: ./scripts/build.sh
set -euo pipefail

cd "$(dirname "$0")/.."
APP="warm-nest"
DIST="./dist"

rm -rf "$DIST"
mkdir -p "$DIST"

echo "[build] 编译 Linux amd64 二进制（vendor 模式）..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -mod=vendor -trimpath -o "$DIST/${APP}" .

echo "[build] 拷贝配置..."
cp app.toml app_dev.toml app_prod.toml "$DIST/" 2>/dev/null || cp app.toml "$DIST/"

echo "[build] 生成构建信息..."
{
  echo "app=$APP"
  echo "commit=$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
  echo "branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown)"
} > "$DIST/.buildinfo"

echo "[build] 打 tar 包..."
tar -czf "${APP}.tar.gz" -C "$DIST" .
echo "[build] 完成: $(pwd)/${APP}.tar.gz"
ls -lh "$DIST"
