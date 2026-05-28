# AGENTS.md

本文档约束本仓库内的自动化开发行为。先读源码，再修改；先验证，再汇报。

## 工作原则

- 有源码就先看源码，不要先碰线上页面猜问题。
- 不改无关文件，不顺手重构。
- 小补丁优先，只有明确收益时才拆新抽象。
- 如果已有用户改动，不回滚、不覆盖，只在必要范围追加。
- 涉及删除文件、数据、容器、镜像前，必须先列清单确认。
- 改完代码必须验证。后端跑 `go test ./...`，前端改动跑 `npm run build` 或对应更小验证。
- 线上修复必须构建镜像、部署容器，并验证接口、页面、容器状态和日志。

## 架构边界

系统是模块化单体：Next.js 前端 + Go API + GORM 数据库。默认单机部署，SQLite 存储在 `/app/data`。

核心链路：用户请求 → Next.js 页面/API 代理 → Go handler → service → repository → 数据库或模型上游。

多节点部署前必须先中心化数据库、队列和限流状态。不要直接把 SQLite 单机应用做 round-robin。

## 后端规范

- `handler/` 负责 HTTP 入参、认证上下文读取、调用 service、返回 `OK` / `Fail`。
- `service/` 负责业务规则、默认值、校验、扣费、退费、ID、时间、鉴权辅助。
- `repository/` 负责数据库访问和 GORM 查询。
- `model/` 只放数据结构、枚举和简单模型定义。
- 列表接口优先使用 `model.Query`、`Normalize`、分页和关键词筛选。
- 业务接口统一返回 `{ code, data, msg }`。
- 新增数据表或关键字段时同步更新 `docs/backend-database.md`。

## 前端规范

- 前端使用 Next.js App Router、React、TypeScript、Ant Design、Tailwind、Zustand。
- API 请求统一放在 `web/src/services/api/`。
- 全局状态放在 `web/src/stores/`，页面私有逻辑放在页面目录。
- 页面只有一个主业务组件时直接写在 `page.tsx`，不要为了命名再包一层 Manager。
- 页面文案保持中文。
- 组件不要多层透传全局配置；需要时直接从 store、hook 或常量读取。
- 浏览器本地业务数据默认用 `localforage`，不要用 `localStorage` 保存大 JSON、图片或生成记录。
- Ant Design 组件优先沿用项目现有写法。

## 生图链路规则

- `/api/v1/images/generations` 和 `/api/v1/images/edits` 走后端代理。
- 请求前按模型算力点预扣，失败或少返回图片时退点。
- 3 分钟最多 3 张，按图片张数计数，不按提交次数计数。
- 空图片文件必须在本系统拦截，不能发给上游。
- 生成日志只保存必要审计信息；用户侧历史不暴露原始 request / response。
- 全站任务展示排队状态；统计排行按成功产出的图片张数排名。

## 数据和隐私

- 用户密码只保存哈希。
- 上游 API Key 只存在后端私有配置，不进入前端和普通接口返回。
- 注册时记录 `registerIP`，同一 IP 只允许注册一个用户。
- 管理后台可以查看生成日志；用户端只能查看自己的云端历史。
- 公开错误信息不要泄露上游密钥、服务器路径和内部网络细节。

## 部署规则

- 镜像标签使用 `v0.1.0-kaixin.N` 递增。
- 构建后推送到 `ghcr.io/versior/kaixin-ai`。
- 线上容器名为 `infinite-canvas`。
- 线上数据目录挂载到 `/root/infinite-canvas/data:/app/data`。
- 部署后至少验证 `/api/health`、关键页面、相关接口、`docker ps` 和最近日志。

## 文档规则

- 功能变化同步更新 README 或 `docs/`。
- 数据库表字段变化同步更新 `docs/backend-database.md`。
- 配置结构变化同步更新 `docs/system-settings.md`。
- 行为约定变化同步更新 `docs/features.md` 或对应专题文档。
