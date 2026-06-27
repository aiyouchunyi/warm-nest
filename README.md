<!-- BEGIN:AUTO 由 /custom_project init 自动生成，请勿手改此段（手改会被下次刷新覆盖）-->
# warm-nest

<TEAM: 一句话描述这个服务>

## 服务信息

| 字段 | 值 |
|------|-----|
| 服务名 | `warm-nest` |
| 仓库 | https://github.com/aiyouchunyi/warm-nest.git |
| Go 版本 | 1.23.0 |
| 配置文件 | app_dev.toml app_local.toml app_prod.toml app.toml |

## 开发流程

1. 从主干 `main` 切功能分支：`git checkout -b <MMDD>-<feature> main`
2. 本地开发 + `go test ./...` 通过
3. 推送功能分支，创建 MR/PR 到主干 `main`
4. Review 通过后合并 → 部署上线

## 分支命名

`<MMDD>-<topic>`，例如：
- `0615-kafka-upgrade`

## 测试

集成测试位于 `tests/`，单元测试随源码同包（`*_test.go`）

```bash
go test ./...                        # 全量
go test ./tests/...                  # 仅集成测试目录
go test -run TestName ./path/to/pkg  # 单个用例
```
<!-- END:AUTO 自动生成段结束，下方是团队完善内容 -->

## 部署 / 升级（手动单机）

当前部署在阿里云 ECS（Alibaba Cloud Linux 3，2C2G），服务跑 Docker、MySQL 宿主机自建。
**升级服务 = 改完代码后跑一条命令**：

```bash
bash deploy/deploy.sh
```

脚本自动完成：本地跨架构构建 amd64 镜像 → `docker save` 压缩 → `scp` 传输 → 远端 `docker load` → 停删旧容器 + 起新容器 → 轮询 `/ping` 健康检查。详见 [deploy/README.md](deploy/README.md)。

要点：
- **配置**：`deploy/deploy.conf`（IP/DSN/端口，含密码已 gitignore）；生产业务配置 `app_prod.toml`（已 gitignore）
- **架构**：本地 Mac arm64 → 服务器 amd64，脚本用 `docker buildx --platform linux/amd64` 跨架构构建
- **DB**：容器连宿主机 MySQL，host 用 `172.18.0.1`（本机 docker 网桥，非默认 172.17.0.1）
- **回滚**：改 `deploy.conf` 的 `IMAGE_TAG` 指向旧镜像重跑，或远端 `docker run` 旧 tag
- **前置**：阿里云安全组放行 8080（服务）；勿对公网放 3306
