## HTTP 알림 안내

선택한 사이트 이벤트가 발생하면 GooseForum이 콜백 URL로 비동기 POST 요청을 보냅니다. 알림은 전달되지 않을 수 있습니다. 전송에 실패해도 글쓰기, 답글, 가입, 신고는 정상적으로 처리됩니다.

### 지원하는 이벤트

- `topic.published`
- `topic.updated`
- `comment.created`
- `user.signup`
- `moderation.report.created`

### 요청 형식

요청 본문은 JSON입니다. 헤더에는 이벤트 이름, 전송 ID, 타임스탬프, 서명이 포함됩니다.

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

`baseUri`에 `topic.url`, `user.url`, `post.url` 같은 사이트 경로를 붙이면 절대 URL을 만들 수 있습니다. 새 연동에서는 `topic`, `user`, `post`, `reporter` 같은 객체를 사용하세요. 최상위에 중복된 필드와 같은 정보가 들어 있습니다. 답글 이벤트에는 내용 미리 보기, 작성자 프로필, 게시물 URL이 포함됩니다. 게시물 신고 이벤트에는 신고된 게시물 정보가 들어 있습니다.

### 서명 확인

`X-Goose-Signature`는 비밀 키, 타임스탬프, 받은 그대로의 요청 본문으로 계산합니다.

```text
sha256 = HMAC_SHA256(secret, timestamp + "." + rawBody)
```

받은 그대로의 본문으로 서명을 확인하세요. 본문을 파싱한 뒤 다시 JSON으로 만들면 서명이 달라질 수 있습니다.

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

### 전송 실패 처리

같은 콜백 URL로 보내는 데 세 번 연속 실패하면 해당 URL을 비활성화하고 중단 상태로 표시합니다. URL을 다시 활성화한 뒤 저장하면 실패 상태가 초기화됩니다.
