## Guide des notifications HTTP

Lorsqu’un événement sélectionné se produit, GooseForum envoie une requête POST asynchrone à l’URL de rappel. La livraison peut échouer. Un échec de livraison ne bloque ni les publications, ni les réponses, ni les inscriptions, ni les signalements.

### Événements pris en charge

- `topic.published`
- `topic.updated`
- `comment.created`
- `user.signup`
- `moderation.report.created`

### Format des requêtes

Le corps de la requête est au format JSON. Les en-têtes contiennent le nom de l’événement, l’identifiant de livraison, l’horodatage et la signature.

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

Combinez `baseUri` avec des chemins comme `topic.url`, `user.url` et `post.url` pour construire les URL absolues. Dans une nouvelle intégration, utilisez les objets `topic`, `user`, `post` et `reporter`. Ils contiennent les mêmes données que les champs en double au premier niveau. Les événements liés aux réponses incluent un aperçu du contenu, le profil de l’auteur et l’URL de la publication. Les signalements incluent les détails de la publication signalée.

### Vérification de la signature

La signature `X-Goose-Signature` est calculée à partir de la clé secrète, de l’horodatage et du corps brut de la requête.

```text
sha256 = HMAC_SHA256(secret, timestamp + "." + rawBody)
```

Utilisez le corps tel que vous l’avez reçu pour vérifier la signature. L’analyse puis la resérialisation du JSON peuvent modifier la signature.

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

### Gestion des échecs de livraison

Après trois échecs consécutifs sur la même URL de rappel, GooseForum la désactive et indique que la livraison est interrompue. Réactivez l’URL et enregistrez-la pour réinitialiser son état.
