# 部署说明

项目推荐使用 Docker 部署。单机、小团队和演示环境可以使用 SQLite；公网多人长期使用建议迁移到 MySQL 或 PostgreSQL，并做好数据备份。

## Docker Compose

```bash
git clone git@github.com:basketikun/infinite-canvas.git
cd infinite-canvas
cp .env.example .env
docker compose up -d --build
```

默认访问：`http://localhost:3000`

## 数据目录

容器内数据目录通常是：

```text
/app/data
```

线上建议挂载到宿主机目录，例如：

```bash
-v /root/infinite-canvas/data:/app/data
```

SQLite 数据库、上传媒体、提示词缓存等都应放在持久化目录里。

## 管理员账号

管理员账号通常为：

```text
admin
```

密码由环境变量或部署配置决定。公网部署后必须修改默认密码。

## 模型渠道配置

部署完成后进入后台：

```text
/admin/settings
```

至少配置：

- 可用模型。
- 默认图片模型。
- 默认文本模型。
- 模型算力点价格。
- 私有模型渠道 Base URL 和 API Key。

如果不配置渠道，云端生图、文本和视频请求无法正常转发到上游。

## Render 部署

Render 一键部署仍可用于演示：

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/basketikun/infinite-canvas)

免费版限制：

- 空闲后会休眠。
- 本地文件不是稳定持久化存储。
- SQLite 数据可能因重启或重新部署丢失。

正式使用建议选择带持久化磁盘的实例，或使用外部数据库。

## 线上验证

部署后至少检查：

```bash
curl -i http://服务器地址/api/health
curl -i http://服务器地址/image
curl -i http://服务器地址/admin
```

同时查看容器状态和日志：

```bash
docker ps
docker logs --tail 120 infinite-canvas
```

日志中不应出现 `panic`、`fatal` 或连续业务错误。
