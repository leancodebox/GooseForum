## HTTP notification guide

HTTP notifications send asynchronous POST requests to callback URLs when selected site events happen. This is best-effort delivery; failures never block posting, replying, registration, or reports.

### Supported events

- `article.published`
- `article.updated`
- `comment.created`
- `user.signup`
- `moderation.report.created`

### Request format

The request body is JSON. Headers include the event name, delivery ID, timestamp, and signature.

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
    "articleId": 123,
    "title": "Hello GooseForum",
    "userId": 1
  }
}
```

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
