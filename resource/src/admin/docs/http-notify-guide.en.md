## HTTP notification guide

HTTP notifications send asynchronous POST requests to callback URLs when selected site events happen. This is best-effort delivery; failures never block posting, replying, registration, or reports.

### Supported events

- `topic.published`
- `topic.updated`
- `comment.created`
- `user.signup`
- `moderation.report.created`

### Request format

The request body is JSON. Headers include the event name, delivery ID, timestamp, and signature.

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

`baseUri` can be used to turn site paths such as `topic.url`, `user.url`, and `post.url` into absolute URLs. New integrations should prefer structured objects such as `topic`, `user`, `post`, and `reporter`, avoiding duplicate top-level mirror fields. Comment events also include the content preview, commenter profile, and post URL; report events for posts include the reported post details.

### Signature verification

`X-Goose-Signature` is calculated from the Secret, timestamp, and raw request body.

```text
sha256 = HMAC_SHA256(secret, timestamp + "." + rawBody)
```

Use the raw body for verification. Do not parse and serialize it again first.

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

### Failure protection

If the same callback URL fails 3 times in a row, the system disables that URL and marks it abnormally stopped. Re-enable and save it to clear the failure state.
