## HTTP-уведомления

При выбранных событиях форум отправляет асинхронные POST-запросы на указанные URL. Доставка не гарантируется. Ошибки доставки не мешают публикации тем и ответов, регистрации или отправке жалоб.

### Поддерживаемые события

- `topic.published`
- `topic.updated`
- `comment.created`
- `user.signup`
- `moderation.report.created`

### Формат запроса

Тело запроса передаётся в формате JSON. Заголовки содержат имя события, идентификатор доставки, временную метку и подпись.

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

Используйте `baseUri`, чтобы преобразовать пути `topic.url`, `user.url` и `post.url` в абсолютные URL. Для новых интеграций используйте объекты `topic`, `user`, `post` и `reporter`. События комментариев также содержат фрагмент текста, профиль автора и ссылку на сообщение. События жалоб содержат сведения о сообщении, на которое подана жалоба.

### Проверка подписи

Подпись `X-Goose-Signature` вычисляется из секретного ключа, временной метки и исходного тела запроса.

```text
sha256 = HMAC_SHA256(secret, timestamp + "." + rawBody)
```

Проверяйте подпись по исходным байтам тела запроса до разбора JSON и повторной сериализации.

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

### Обработка ошибок

После трёх последовательных ошибок доставки на один адрес форум отключает его. Чтобы сбросить ошибку, включите адрес снова и сохраните настройки.
