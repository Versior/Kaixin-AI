# 画布数据结构

画布项目主要保存在浏览器本地。业务 JSON 和媒体 Blob 分开存储，避免长期保存大体积 base64。

## 存储位置

- 画布项目 JSON：`localForage`，数据库名 `linggan-sws`，storeName `app_state`，key 为 `linggan-sws:canvas_store`。
- 我的素材 JSON：`localForage`，数据库名 `linggan-sws`，storeName `app_state`，key 为 `linggan-sws:asset_store`。
- 图片 Blob：`localForage`，数据库名 `linggan-sws`，storeName `image_files`。
- 视频等媒体 Blob：`localForage`，数据库名 `linggan-sws`，storeName `media_files`。

节点、助手会话和素材只保存展示 URL、`storageKey` 和元信息。真实文件通过 `storageKey` 读取。

## CanvasProject

```ts
type CanvasProject = {
  id: string;
  title: string;
  createdAt: string;
  updatedAt: string;
  nodes: CanvasNodeData[];
  connections: CanvasConnection[];
  chatSessions: CanvasAssistantSession[];
  activeChatId: string | null;
  backgroundMode: "lines" | "dots" | "blank";
  viewport: { x: number; y: number; k: number };
};
```

字段说明：

- `id`：项目 ID，前端生成。
- `title`：项目名称。
- `createdAt` / `updatedAt`：ISO 时间字符串。
- `nodes`：节点列表。
- `connections`：节点连线。
- `chatSessions`：画布助手会话。
- `activeChatId`：当前助手会话 ID。
- `backgroundMode`：背景模式。
- `viewport`：视口平移和缩放。

## CanvasNodeData

```ts
type CanvasNodeData = {
  id: string;
  type: "image" | "text" | "config" | "video";
  title: string;
  position: { x: number; y: number };
  width: number;
  height: number;
  metadata?: CanvasNodeMetadata;
};
```

通用字段：

- `id`：节点 ID。
- `type`：节点类型。
- `title`：节点标题。
- `position`：画布世界坐标。
- `width` / `height`：画布世界尺寸。
- `metadata`：节点内容和业务状态。

## CanvasNodeMetadata

常见字段：

```ts
type CanvasNodeMetadata = {
  content?: string;
  prompt?: string;
  status?: "idle" | "success" | "loading" | "error";
  errorDetails?: string;
  storageKey?: string;
  url?: string;
  model?: string;
  size?: string;
  quality?: string;
  count?: number;
  references?: string[];
  generatedAt?: string;
};
```

不同节点会使用不同子集。

文本节点主要使用 `content`、`prompt`、`status`。图片节点主要使用 `url`、`storageKey`、`prompt`、`model`、`size`、`quality`、`count`。视频节点主要使用 `url`、`storageKey` 和视频参数。

## CanvasConnection

```ts
type CanvasConnection = {
  id: string;
  sourceNodeId: string;
  targetNodeId: string;
};
```

连线表达节点引用关系。生成配置节点会读取上游节点，生成结果也会通过连线保留来源。

## 媒体存储约定

图片和视频文件存储在 localForage 的独立 store 中。

- JSON 中保存 `storageKey`。
- 读取节点时通过 `storageKey` 恢复 Blob URL。
- 导出项目时需要把 JSON 和引用到的媒体文件一起打包。
- 导入项目时先恢复媒体，再恢复项目 JSON。

## 清理约定

删除节点或素材时，不应立即误删仍被其他节点引用的媒体。撤销场景尤其要小心：如果节点恢复但媒体已经删除，会导致图片丢失。

清理媒体前应确认没有项目、节点、助手会话或素材继续引用同一个 `storageKey`。

## 后端同步边界

当前画布项目主要保存在浏览器本地。后续如果接入后端同步，应保持以下边界：

- 项目 JSON 和媒体文件分开上传。
- 不直接把大体积 base64 写入项目 JSON。
- 服务端不要假设客户端 `storageKey` 永远有效。
- 导入导出格式要兼容当前 ZIP 包结构。
