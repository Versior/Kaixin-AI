# 后端数据库

后端使用 GORM 管理数据库连接和表结构迁移。默认数据库是 SQLite，路径通常在容器 `/app/data` 下。也可以通过配置切换到 MySQL 或 PostgreSQL。

启动时执行 `AutoMigrate`。本文档只记录当前代码实际使用的数据表和关键字段。

## users

用户表。保存账号、角色、算力点、第三方登录标识和注册 IP。

关键字段：

- `id`：用户 ID。
- `username`：用户名，唯一。
- `password`：密码哈希。
- `email`：邮箱。
- `display_name`：昵称。
- `avatar_url`：头像。
- `role`：角色，`user` 或 `admin`。
- `credits`：算力点余额。
- `aff_code`：邀请码。
- `aff_count`：邀请人数冗余统计。
- `inviter_id`：邀请人 ID。
- `github_id`：GitHub 用户 ID。
- `linux_do_id`：Linux.do 用户 ID。
- `wechat_id`：微信用户 ID。
- `register_ip`：注册 IP。同一 IP 只允许注册一个用户。
- `status`：用户状态，`active` 或 `ban`。
- `last_login_at`：最近登录时间。
- `extra`：扩展 JSON，第三方资料按平台命名空间保存。
- `created_at` / `updated_at`：创建和更新时间。

## credit_logs

算力点流水表。记录管理员调整、模型调用扣费、失败退费等变化。

关键字段：

- `id`：流水 ID。
- `user_id`：用户 ID。
- `username`：用户名快照。
- `type`：流水类型。
- `amount`：变化数量，扣费为负数，充值或退费为正数。
- `balance`：变化后的余额。
- `remark`：备注。
- `created_at`：创建时间。

## prompts

提示词表。保存公开提示词、同步自 GitHub 的提示词案例和后台手工维护内容。

关键字段：

- `id`：提示词 ID。
- `title`：标题。
- `cover_url`：封面。
- `prompt`：提示词正文。
- `tags`：标签数组 JSON。
- `category`：分类。
- `preview`：Markdown 预览内容。
- `created_at` / `updated_at`：创建和更新时间。

`github_url` 只用于接口返回，不作为核心数据库字段依赖。

## assets

后台素材表。

关键字段：

- `id`：素材 ID。
- `title`：标题。
- `type`：素材类型，如图片或视频。
- `url`：资源地址。
- `cover_url`：封面地址。
- `tags`：标签数组 JSON。
- `description`：说明。
- `created_at` / `updated_at`：创建和更新时间。

## settings

系统配置表。当前主要使用两行：`public` 和 `private`。

- `public`：公开配置，前端可以读取。
- `private`：私有配置，只给后端和管理员使用。

字段：

- `id`：配置 key。
- `value`：配置 JSON。
- `updated_at`：更新时间。

详细结构见 `docs/system-settings.md`。

## generation_logs

生成日志表。记录用户通过后端代理发起的模型调用审计信息。

关键字段：

- `id`：日志 ID。
- `user_id`：用户 ID。
- `username`：用户名快照。
- `type`：生成类型，如 `image`、`text`、`video`。
- `model`：模型名。
- `status`：状态，如 `success`、`failed`、`rate_limited`。
- `image_count`：成功产出的图片张数。
- `prompt`：提示词摘要或必要内容。
- `error_message`：错误信息。
- `request`：请求摘要。
- `response`：响应摘要。
- `created_at`：创建时间。

用户端历史接口只返回当前用户自己的安全字段。管理员后台可查看全站日志。

## generation_tasks

全站生图任务表。记录进入异步队列的图片生成任务。

关键字段：

- `id`：任务 ID。
- `user_id`：用户 ID。
- `username`：用户名快照。
- `model`：模型名。
- `batch_count`：本任务请求生成的图片张数。
- `status`：`queued`、`running`、`succeeded`、`failed` 等状态。
- `error_message`：失败原因。
- `created_at` / `updated_at`：创建和更新时间。

全站任务面板读取该表和内存队列状态。统计排行主要按成功图片产出统计。

## 数据一致性约定

- 算力点扣费和退费必须写入 `credit_logs`。
- 上游失败、空返回或少返回图片时，需要退还未产出的图片算力点。
- `generation_logs` 用于审计和历史，`generation_tasks` 用于任务状态，两者口径不同。
- 公开接口不能暴露上游 API Key、完整私有请求头或服务器内部路径。
