<p align="center">
  <img src="web/public/logo.svg" width="96" alt="灵感事务所 logo">
</p>

<h1 align="center">灵感事务所</h1>

灵感事务所是一套面向图片创作、素材沉淀和团队运营的 AI 工作台。它把灵感画布、生图工作台、视频创作台、素材库、提示词库、账号体系、算力点、后台管理和云端生图队列放在同一个系统里，适合持续生产图片内容、复用素材、管理用户和追踪生成记录。

项目仍在快速迭代。数据库结构、前端存储结构和接口字段会随着业务直接调整。正式公网多人使用前，请先做好数据库备份、上游渠道限额和管理员账号安全配置。

## 核心能力

- 灵感画布：多项目、节点拖拽缩放、连线、小地图、撤销重做、导入导出。
- 图片创作：文生图、图生图、参考图编辑、批量生图、本地历史和云端账号历史。
- 视频创作：提示词和参考图生成视频，支持清晰度、尺寸、秒数配置。
- 素材沉淀：我的素材、后台素材库、图片和视频媒体本地持久化。
- 提示词库：同步多个 GitHub 图片提示词仓库，按分类展示案例。
- 账号与算力点：账号密码注册、算力点扣费、失败退点、注册 IP 限制。
- 全站队列：用户端展示全站生图任务、排队状态、统计排行和用户生图排行。
- 管理后台：用户、算力点日志、模型渠道、提示词、素材、生成日志、系统公告和公开配置。

## 技术栈

- 前端：Next.js App Router、React、TypeScript、Tailwind CSS、Ant Design、Zustand、TanStack Query。
- 后端：Go、Gin、GORM。
- 数据库：默认 SQLite，也支持 MySQL 和 PostgreSQL。
- 部署：Docker / Docker Compose。

## 快速启动

```bash
git clone git@github.com:Versior/Kaixin-AI.git
cd Kaixin-AI
cp .env.example .env
docker compose up -d --build
```

访问：`http://localhost:3000`

默认管理员账号通常为 `admin`，密码由环境变量或部署配置决定。部署到公网前请立刻修改管理员密码。

## 常用页面

- `/image`：生图工作台。
- `/image/history`：当前账号的云端生图历史。
- `/video`：视频创作台。
- `/canvas`：灵感画布。
- `/assets`：我的素材。
- `/prompts`：提示词库。
- `/admin`：管理后台入口。

## AI 接口路径

用户侧通过后端代理访问 OpenAI 兼容接口：

- `POST /api/v1/images/generations`
- `POST /api/v1/images/edits`
- `POST /api/v1/chat/completions`
- `POST /api/v1/videos`
- `GET /api/v1/images/tasks`
- `GET /api/v1/images/stats`
- `GET /api/v1/images/history`

后端会按模型名选择已启用渠道，执行算力点预扣、上游请求、生成日志保存和失败退点。

## 部署建议

单机部署优先使用 Docker Compose 和 SQLite。多人公网使用建议至少做到：

- 使用强管理员密码。
- 配置稳定的模型渠道和算力点价格。
- 定期备份 `/app/data`。
- 如果需要多节点或负载均衡，先迁移到 PostgreSQL/MySQL，并把队列和限流状态中心化。

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
