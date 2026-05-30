<p align="center">
  <img src="web/public/logo.svg" width="96" alt="灵感事务所 logo">
</p>

<h1 align="center">灵感事务所</h1>

> ⚠️ **本项目是基于 [basketikun/infinite-canvas](https://github.com/basketikun/infinite-canvas) 的二改版本。**
> 原项目 © [basketikun](https://github.com/basketikun)，感谢原作者的开源贡献。
>
> 二改仓库：[https://github.com/Versior/Kaixin-AI](https://github.com/Versior/Kaixin-AI)

---

灵感事务所是一套面向图片创作、素材沉淀和团队运营的 AI 工作台。它把灵感画布、生图工作台、视频创作台、素材库、提示词库、账号体系、算力点、后台管理和云端生图队列放在同一个系统里，适合持续生产图片内容、复用素材、管理用户和追踪生成记录。

## 二改内容

基于 `basketikun/infinite-canvas` 原版，本 fork 做了以下调整：

- **品牌重命名**：全局品牌名改为「灵感事务所」。
- **主页开放访问**：`/` 和 `/login` 无需登录即可浏览。
- **SMTP 邮箱验证码注册**：支持 QQ 邮箱等 SMTP 服务发送注册验证码。
- **主题闪烁修复**：修复浅色/深色主题在页面加载时的闪烁问题。
- **管理后台加载态**：修复从用户端跳转管理后台时白屏/无反应的问题。
- **注册赠送积分**：新用户注册自动获得 200 算力点。
- **管理后台入口**：登录后右上角头像下拉菜单可直达 `/admin`。
- **Docker Compose 部署优化**：`docker-compose.hk.yml` 适配云端部署，支持 GHCR 镜像拉取。
- **GitHub Actions CI/CD**：push 到 main 分支自动构建 Docker 镜像。

## 核心能力

- **灵感画布**：多项目、节点拖拽缩放、连线、小地图、撤销重做、导入导出。
- **图片创作**：文生图、图生图、参考图编辑、批量生图、本地历史和云端账号历史。
- **视频创作**：提示词和参考图生成视频，支持清晰度、尺寸、秒数配置。
- **素材沉淀**：我的素材、后台素材库、图片和视频媒体本地持久化。
- **提示词库**：同步多个 GitHub 图片提示词仓库，按分类展示案例。
- **账号与算力点**：账号密码注册、邮箱验证码注册、算力点扣费、失败退点、注册 IP 限制。
- **全站队列**：用户端展示全站生图任务、排队状态、统计排行和用户生图排行。
- **管理后台**：用户管理、算力点日志、模型渠道、提示词、素材、生成日志、系统公告、SMTP 配置和公开配置。

## 技术栈

- **前端**：Next.js App Router、React、TypeScript、Tailwind CSS、Ant Design、Zustand、TanStack Query。
- **后端**：Go、Gin、GORM。
- **数据库**：默认 SQLite，也支持 MySQL 和 PostgreSQL。
- **部署**：Docker / Docker Compose，GitHub Actions 自动构建镜像。

## 快速启动

### Docker Compose（推荐）

```bash
git clone https://github.com/Versior/Kaixin-AI.git
cd Kaixin-AI
cp .env.example .env
docker compose up -d
```

访问：`http://localhost:3000`

### 云端部署（华为云 / 香港节点参考）

```bash
git clone https://github.com/Versior/Kaixin-AI.git /tmp/kaixin-github
cd /tmp/kaixin-github
docker compose -f docker-compose.hk.yml up -d
```

> 构建镜像请通过 GitHub Actions 完成，**禁止在服务器上本地构建**。

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `ADMIN_USERNAME` | 管理员用户名 | `admin` |
| `ADMIN_PASSWORD` | 管理员密码 | `admin` |
| `JWT_SECRET` | JWT 签名密钥 | 随机生成 |
| `DATABASE_DSN` | 数据库路径 | `data/infinite-canvas.db` |
| `PUBLIC_IMAGE_BASE_URL` | 图片公网访问地址 | `http://localhost:3000` |
| `GIN_MODE` | Gin 运行模式 | `release` |

部署到公网前**务必修改管理员密码**。

## 常用页面

| 路径 | 说明 |
|------|------|
| `/` | 主页（公开） |
| `/login` | 登录/注册（公开） |
| `/image` | 生图工作台 |
| `/image/history` | 云端生图历史 |
| `/video` | 视频创作台 |
| `/canvas` | 灵感画布 |
| `/assets` | 我的素材 |
| `/prompts` | 提示词库 |
| `/admin` | 管理后台入口（需 admin 角色） |

## API 接口

用户侧通过后端代理访问 OpenAI 兼容接口：

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/images/generations` | 文生图 / 图生图 |
| POST | `/api/v1/images/edits` | 图片编辑 |
| POST | `/api/v1/chat/completions` | 聊天补全 |
| POST | `/api/v1/videos` | 视频生成 |
| GET | `/api/v1/images/tasks` | 生图任务列表 |
| GET | `/api/v1/images/stats` | 生图统计 |
| GET | `/api/v1/images/history` | 生图历史 |
| POST | `/api/auth/login` | 登录 |
| POST | `/api/auth/register` | 注册 |
| POST | `/api/auth/send-code` | 发送邮箱验证码 |

后端会按模型名选择已启用渠道，执行算力点预扣、上游请求、生成日志保存和失败退点。

## 部署建议

单机部署优先使用 Docker Compose 和 SQLite。多人公网使用建议至少做到：

- 使用强管理员密码。
- 配置稳定的模型渠道和算力点价格。
- 定期备份 `/app/data`（SQLite 数据文件）。
- 如果需要多节点或负载均衡，先迁移到 PostgreSQL/MySQL，并把队列和限流状态中心化。
- **所有镜像构建走 GitHub Actions，禁止在服务器上本地编译或构建。**

## 文档入口

- [功能说明](docs/features.md)
- [部署说明](docs/deployment.md)
- [后端数据库](docs/backend-database.md)
- [系统配置](docs/system-settings.md)
- [接口响应](docs/api-response.md)
- [画布数据结构](docs/canvas-data-structure.md)
- [画布节点手册](docs/canvas-node-manual.md)
- [画布快捷键](docs/canvas-shortcuts.md)
- [提示词源](docs/third-party-prompt-repositories.md)
- [待测试清单](docs/pending-test.md)
- [待办事项](docs/todo.md)

## 效果展示

<table width="100%">
  <tr>
    <td width="50%"><img src="https://i.ibb.co/TDFvGWDT/image.png" alt="image" border="0"></td>
    <td width="50%"><img src="https://i.ibb.co/zVwJq3YS/image.png" alt="image" border="0"></td>
  </tr>
  <tr>
    <td width="50%"><img src="https://i.ibb.co/PvY3qhhK/image.png" alt="image" border="0"></td>
    <td width="50%"><img src="https://i.ibb.co/7D04LwN/image.png" alt="image" border="0"></td>
  </tr>
</table>

## 原项目信息

| | |
|---|---|
| **原项目** | [basketikun/infinite-canvas](https://github.com/basketikun/infinite-canvas) |
| **原作者** | [basketikun](https://github.com/basketikun) |
| **二改仓库** | [Versior/Kaixin-AI](https://github.com/Versior/Kaixin-AI) |
| **协议** | 与原项目一致 |
