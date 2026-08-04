# @gooseforum/client

Framework-agnostic page payload contracts and HTTP client for GooseForum themes and browser applications.

```ts
import { createGooseClient } from '@gooseforum/client'

const client = createGooseClient({
  baseURL: window.location.origin,
})

const page = await client.pages.fetch(window.location.href)
```

The package intentionally has no dependency on Vue, routing libraries, UI components, or localization implementations. API errors preserve `messageCode` and `params`; applications decide how to translate and display them.
