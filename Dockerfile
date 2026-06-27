# 构建环境：公网官方 Go 镜像（与原内部镜像同版本 1.23.10-alpine）
FROM golang:1.23.10-alpine AS builder
# 设置工作目录
WORKDIR /app
# 先复制依赖描述 + vendor，让依赖层在业务代码变更时仍可命中缓存
COPY go.mod go.sum ./
COPY vendor ./vendor
# 再复制项目源代码
COPY . .
# 编译 Go 程序（CGO 关闭，产出静态二进制，可在 alpine/scratch 直接跑）
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -trimpath -o warm-nest .

# 运行环境：公网 alpine（替代拉不到的内部 baseimage）。补两样原 Ubuntu base 自带、本服务需要的依赖：
#   - ca-certificates：调微信等 HTTPS 外部 API 必需，缺了报 x509 证书错
#   - tzdata：奖励评估按 Asia/Shanghai 算月末/当日，缺了 time.LoadLocation 失败（代码有 FixedZone 兜底，装上更稳）
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
# 时区设为 Asia/Shanghai：Go 的 time.Local（日志时间戳/time.Now）随 TZ env 走北京时间，
# 再固化 /etc/localtime + /etc/timezone 让容器内 date 等系统工具也对齐（不再是 UTC）
ENV TZ=Asia/Shanghai
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone
# 将工作目录设置为 /app
WORKDIR /app
# 从 builder 阶段复制编译好的二进制文件 + 各环境配置
# 注：原内部 Dockerfile 还 COPY 了 .buildinfo（公司 CI 构建期生成），手动构建无此文件，已移除
COPY --from=builder /app/warm-nest /app/app.toml /app/app_dev.toml /app/app_prod.toml ./
# 运行应用程序
CMD ["./warm-nest"]

EXPOSE 8080