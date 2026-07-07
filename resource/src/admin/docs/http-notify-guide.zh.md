## HTTP 通知使用说明

HTTP 通知会在选中的站内事件发生后，异步向回调地址发送 POST 请求。这个功能是尽力通知，失败不会影响用户发帖、回复或注册。

### 支持的事件

- `article.published`
- `article.updated`
- `comment.created`
- `user.signup`
- `moderation.report.created`

### 请求格式

请求体固定为 JSON，Header 会带上事件名、投递 ID、时间戳和签名。

```http
POST /your-webhook HTTP/1.1
Content-Type: application/json
X-Goose-Event: article.published
X-Goose-Delivery: 7c9f0b0fd4e2a111
X-Goose-Timestamp: 1710000000
X-Goose-Signature: sha256=...
```

```json
{
  "event": "article.published",
  "timestamp": 1710000000,
  "data": {
    "baseUri": "http://localhost:5234",
    "topic": {
      "id": 123,
      "title": "Hello GooseForum",
      "url": "/p/post/123",
      "description": "Topic summary",
      "firstImageUrl": "",
      "userId": 1,
      "user": {
        "id": 1,
        "username": "alice",
        "nickname": "Alice",
        "displayName": "Alice",
        "avatarUrl": "/static/pic/1.webp",
        "url": "/u/1"
      },
      "categoryIds": [2],
      "categories": [
        { "id": 2, "name": "Announcements", "slug": "announcements" }
      ]
    },
    "user": {
      "id": 1,
      "username": "alice",
      "nickname": "Alice",
      "displayName": "Alice",
      "avatarUrl": "/static/pic/1.webp",
      "url": "/u/1"
    }
  }
}
```

`baseUri` 可用于把 `article.url`、`user.url`、`comment.url` 这类站内路径拼成完整 URL；新接入建议优先使用 `article`、`user`、`comment`、`reporter` 等结构化对象，避免重复解析顶层镜像字段。评论事件会额外包含评论内容预览、评论人信息和评论 URL；举报评论时也会包含被举报评论信息。

### 签名校验

`X-Goose-Signature` 由 Secret、时间戳和原始请求体计算得出。

```text
sha256 = HMAC_SHA256(secret, timestamp + "." + rawBody)
```

验签时请使用原始 body，不要先解析再重新序列化。

### Node.js

```js
import crypto from 'node:crypto'

function verify(secret, timestamp, rawBody, signature) {
  const digest = crypto
    .createHmac('sha256', secret)
    .update(timestamp + '.' + rawBody)
    .digest('hex')

  return signature === 'sha256=' + digest
}
```

### Go

```go
func verify(secret string, timestamp string, rawBody []byte, signature string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(timestamp))
    mac.Write([]byte("."))
    mac.Write(rawBody)
    want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(want))
}
```

### 失败保护

同一个回调地址连续 3 次通知失败后，系统会自动关闭该地址，并标记为异常终止。重新启用并保存后会清空失败状态。
