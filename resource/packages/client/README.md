# @gooseforum/client

Framework-agnostic client SDK for building GooseForum themes and browser applications. It is the supported boundary between the Go server and independently developed user interfaces; it has no dependency on Vue, routing libraries, UI components, or localization implementations.

## Quick start

```ts
import { createGooseClient, updateDocumentMetadata } from '@gooseforum/client'

const client = createGooseClient({
  baseURL: window.location.origin,
})

const initialPage = client.readInitialPayload()
const nextPage = await client.pages.fetch('/?sort=hot')
updateDocumentMetadata(nextPage)

await client.api.topics.like(42, 1)
```

`PagePayloadMap` is a discriminated map of every server-rendered page. Narrowing on `payload.component` also narrows `payload.props`, so a theme does not need to redefine page data contracts.

## Domain API

`client.api` exposes typed APIs grouped by responsibility:

- `accessGroups`, `posts`, `topics`, and `moderation`
- `notifications` and `chat`
- `users` and `auth`
- `themes` and `uploads`

The SDK owns HTTP methods, routes, query serialization, request shapes, result envelopes, and structured errors. Applications remain responsible for translating `messageCode` and `params` into user-facing text.

```ts
import { GooseClientError } from '@gooseforum/client'

try {
  await client.api.posts.create({ topicId: 42, content: 'Hello' })
} catch (error) {
  if (error instanceof GooseClientError) {
    translate(error.messageCode, error.params)
  }
}
```

Password login uses an ephemeral server public key. Browser themes can use the complete helper flow:

```ts
import { loginWithPassword } from '@gooseforum/client/browser'

await loginWithPassword(client.api.auth, {
  username,
  password,
  captchaId,
  captchaCode,
})
```

## Page protocol and extensions

The default client rejects unknown page components. A theme that installs a server extension can register its components without weakening validation for everything else:

```ts
const client = createGooseClient({
  pages: {
    components: ['plugin.dashboard'],
    validate(payload) {
      // Optional extension-specific runtime validation.
    },
  },
})
```

For forward-compatible shells that render their own unknown-page fallback, set `allowUnknownComponents: true` and choose the generic payload result explicitly:

```ts
import type { PagePayload } from '@gooseforum/client'

const client = createGooseClient<PagePayload>({
  pages: { allowUnknownComponents: true },
})
```

TypeScript integrations can preserve a typed extension payload with the client generic:

```ts
import type { AnyPagePayload, PagePayload } from '@gooseforum/client'

type ThemePage = AnyPagePayload | PagePayload<DashboardProps, 'plugin.dashboard'>

const client = createGooseClient<ThemePage>({
  pages: { components: ['plugin.dashboard'] },
})
```

The SDK always validates the page envelope, component policy, and protocol major version. Use `pages.validate` when a deployment needs deeper runtime validation of extension props.

## Compatibility policy

- Payload protocol `1.x` changes are additive. Removing or changing the meaning of existing fields requires a new protocol major version.
- SDK minor releases may add page components, fields, and domain methods. Existing methods and types remain source-compatible within the same SDK major version.
- A new server page must be added to `PagePayloadMap` and `pageComponents` in the same change.
- A public theme API route must be added to `siteApiRoutes`. Go tests verify that every catalogued route is registered by the server.
- Applications should handle `GooseProtocolError` by falling back to a full page load or an upgrade screen.

## Development

From `resource/`:

```bash
pnpm client:build
pnpm client:test
pnpm typecheck
```

The package is ESM-only and publishes compiled JavaScript plus declarations from `dist/`. `prepack` always rebuilds the package, preventing stale declarations from being published.
