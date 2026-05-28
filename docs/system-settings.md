# 系统配置

系统配置保存在 `settings` 表中，主要分为公开配置 `public` 和私有配置 `private`。

公开配置会返回给前端。私有配置只允许管理员读取，后端调用模型时也会使用私有配置。

## public.value

示例：

```json
{
  "modelChannel": {
    "availableModels": ["gpt-5.5", "gpt-image-2"],
    "modelCosts": [
      { "model": "gpt-5.5", "credits": 1 },
      { "model": "gpt-image-2", "credits": 10 }
    ],
    "defaultModel": "gpt-image-2",
    "defaultImageModel": "gpt-image-2",
    "defaultTextModel": "gpt-5.5",
    "systemPrompt": "",
    "allowCustomChannel": true
  },
  "auth": {
    "allowRegister": true,
    "linuxDo": {
      "enabled": false
    }
  },
  "announcement": {
    "enabled": false,
    "title": "",
    "content": "",
    "version": "",
    "oncePerVersion": true
  }
}
```

### modelChannel

- `availableModels`：前端允许选择的系统模型。
- `modelCosts`：模型算力点价格。后端请求前按模型匹配预扣。
- `defaultModel`：默认模型。
- `defaultImageModel`：默认图片模型。
- `defaultTextModel`：默认文本模型。
- `systemPrompt`：系统提示词。
- `allowCustomChannel`：是否允许用户在浏览器本地配置自定义渠道。

`modelCosts` 每项：

- `model`：模型名。
- `credits`：每次或每张图片消耗的算力点，具体由调用链路解释。

### auth

- `allowRegister`：是否允许账号密码注册。
- `linuxDo.enabled`：是否开启 Linux.do 登录。

注册开启时，系统仍会检查注册 IP。同一 IP 只能注册一个用户。

### announcement

网站公告弹窗配置：

- `enabled`：是否启用。
- `title`：公告标题。
- `content`：公告内容。
- `version`：公告版本。版本变化后可重新弹出。
- `oncePerVersion`：同一版本是否只弹一次。

## private.value

示例：

```json
{
  "channels": [
    {
      "id": "channel-1",
      "name": "默认渠道",
      "protocol": "openai",
      "baseUrl": "https://example.com/v1",
      "apiKey": "sk-xxx",
      "models": ["gpt-image-2", "gpt-5.5"],
      "enabled": true
    }
  ],
  "oauth": {
    "linuxDo": {
      "clientId": "",
      "clientSecret": "",
      "redirectUri": ""
    }
  }
}
```

### channels

模型渠道列表。后端代理 `/api/v1/*` 请求时按模型名选择启用渠道。

- `id`：渠道 ID。
- `name`：渠道名称。
- `protocol`：协议，目前按 OpenAI 兼容接口处理。
- `baseUrl`：上游 Base URL。
- `apiKey`：上游 API Key，只能存在私有配置里。
- `models`：该渠道可用模型。
- `enabled`：是否启用。

### oauth.linuxDo

Linux.do 登录配置：

- `clientId`
- `clientSecret`
- `redirectUri`

## 配置安全

- `private.value` 不能返回给普通前端接口。
- `apiKey` 不能进入前端、普通日志或错误消息。
- 更新配置后应验证模型列表、登录入口和公告弹窗是否符合预期。
