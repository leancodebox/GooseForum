## Guida alle notifiche HTTP

Quando si verifica uno degli eventi selezionati, GooseForum invia una richiesta POST asincrona all’URL di callback. L’invio non è garantito. Un errore di consegna non blocca pubblicazioni, risposte, registrazioni o segnalazioni.

### Eventi supportati

- `topic.published`
- `topic.updated`
- `comment.created`
- `user.signup`
- `moderation.report.created`

### Formato della richiesta

Il corpo della richiesta è in formato JSON. Le intestazioni riportano il nome dell’evento, l’ID della consegna, il timestamp e la firma.

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

Usa i percorsi `topic.url`, `user.url` e `post.url` insieme a `baseUri` per costruire gli URL assoluti. Nelle nuove integrazioni usa gli oggetti `topic`, `user`, `post` e `reporter`. Contengono gli stessi dati dei campi duplicati al primo livello. Gli eventi relativi alle risposte includono l’anteprima del testo, il profilo dell’autore e l’URL del post. Le segnalazioni includono anche i dati del post segnalato.

### Verifica della firma

La firma `X-Goose-Signature` è calcolata a partire dalla chiave segreta, dal timestamp e dal corpo originale della richiesta.

```text
sha256 = HMAC_SHA256(secret, timestamp + "." + rawBody)
```

Verifica la firma sul corpo così come l’hai ricevuto. Se lo converti in JSON e lo serializzi di nuovo, la firma potrebbe cambiare.

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

### Gestione degli errori di consegna

Se un URL di callback non risponde correttamente per tre invii consecutivi, GooseForum lo disattiva e lo contrassegna come interrotto. Riattiva l’URL e salva per azzerare lo stato di errore.
