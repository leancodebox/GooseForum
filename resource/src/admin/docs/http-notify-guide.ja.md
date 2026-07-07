## HTTP 通知の使い方

HTTP 通知は、選択したサイト内イベントが発生したときに、コールバック URL へ非同期で POST リクエストを送信します。この機能はベストエフォートです。送信に失敗しても、投稿、返信、登録、通報の処理には影響しません。

### 対応イベント

- `topic.published`
- `topic.updated`
- `comment.created`
- `user.signup`
- `moderation.report.created`

### リクエスト形式

リクエスト本文は JSON です。Header にはイベント名、配信 ID、タイムスタンプ、署名が含まれます。

```http
POST /your-webhook HTTP/1.1
Content-Type: application/json
X-Goose-Event: topic.published
X-Goose-Delivery: 7c9f0b0fd4e2a111
X-Goose-Timestamp: 1710000000
X-Goose-Signature: sha256=...
```

```json
{
  "event": "topic.published",
  "timestamp": 1710000000,
  "data": {
    "baseUri": "http://localhost:5234",
    "topic": {
      "id": 123,
      "title": "Hello GooseForum",
      "url": "/p/post/123",
      "description": "Article summary",
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

`baseUri` は `topic.url`、`user.url`、`post.url` などのサイト内パスを絶対 URL にするために使えます。新しい連携では、重複するトップレベルのミラーフィールドではなく、`topic`、`user`、`post`、`reporter` などの構造化オブジェクトを優先して使ってください。コメントイベントには内容プレビュー、投稿者情報、post URL も含まれます。post への通報イベントには、通報対象の post 情報も含まれます。

### 署名検証

`X-Goose-Signature` は Secret、タイムスタンプ、元のリクエスト本文から計算されます。

```text
sha256 = HMAC_SHA256(secret, timestamp + "." + rawBody)
```

検証には元の body を使ってください。先に JSON を解析して再シリアライズしないでください。

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

### 失敗保護

同じコールバック URL への通知が連続 3 回失敗すると、システムはその URL を自動的に無効化し、「異常終了」として表示します。再度有効化して保存すると、失敗状態はクリアされます。
