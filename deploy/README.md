# warm-nest 部署

本地构建镜像 → scp 传输 → 远端加载 → 替换容器 → 健康检查，一条命令完成。

## 一键部署

```bash
# 在项目根目录执行
bash deploy/deploy.sh
```

执行流程（deploy.sh）：

| 步 | 动作 |
|----|------|
| 0 | 预检本地 docker/buildx + SSH 连通性 |
| 1 | `docker buildx --platform linux/amd64` 跨架构构建（本地 arm64 → 服务器 amd64） |
| 2 | `docker save` + gzip 导出镜像 |
| 3 | `scp` 传输到远端 `/root/warm-nest-deploy` |
| 4 | 远端 `docker load` 加载镜像 |
| 5 | 停删旧容器 → 起新容器（端口/卷/env 注入） |
| 6 | 轮询 `/ping` 健康检查（30s），通过才算成功，并打印容器状态+日志 |

## 配置与密钥

- **`deploy/deploy.conf`**（含敏感值，已 gitignore，勿提交）：`SSH_HOST` / `DB_DSN`（host 用 `172.18.0.1` 本机 docker 网桥）/ `JWT_SECRET_KEY` / `WECHAT_SECRET` / `IMAGE_TAG`（改 tag 便于回滚）。
- **`app_prod.toml`**（无密码，已进仓库）：mode=prod 时加载的非敏感业务配置（web/jwt expire/storage/wechat 占位）。
- **密钥注入方式**：DB 密码 / JWT key / 微信 secret **不落任何文件**。部署时脚本把 deploy.conf 里的值写到远端 `/root/warm-nest-deploy/warm-nest.env`（权限 600），容器用 `docker --env-file` 注入。框架 env 覆盖规则：`DATABASE_MYSQL` / `JWT_SECRET_KEY` / `WECHAT_SECRET`（大写下划线，覆盖 toml 同名项）。

## 前置条件（一次性）

- 本地：`docker` + `docker buildx`，且能 `ssh root@<IP>` 连通
- 远端：已装 docker；已装 MySQL，建库 `warm-nest`(utf8mb4) + 账号 `warmnest`；MySQL `bind-address=0.0.0.0` 监听网桥
- 阿里云安全组：放行 **8080**（服务端口，对外）；**勿对公网放 3306**

## 待办（拿到值后填 app_prod.toml 重新跑 deploy.sh 即可）

- `[wechat]` AppID / AppSecret / 两个模板ID（当前留空，调微信的接口会失败）
- `[storage].baseUrl`：ICP 备案有域名后改成 https 域名

## 回滚

旧镜像仍在远端：`ssh root@<IP> "docker images warm-nest"` 看历史 tag，
`docker rm -f warm-nest && docker run ... warm-nest:<旧tag>`（或改 deploy.conf 的 IMAGE_TAG 重跑）。

## 排障

```bash
ssh root@<IP> "docker logs --tail 50 warm-nest"   # 看服务日志
ssh root@<IP> "docker ps -a | grep warm-nest"      # 看容器状态
curl http://<IP>:8080/ping                          # 外部探活
```
