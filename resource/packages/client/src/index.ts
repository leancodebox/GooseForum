import { readInitialPayload as readBrowserPayload } from './browser.js'
import { createHttpClient, type GooseClientOptions, type GooseRequestInit } from './http/client.js'
import { createPageClient } from './pages.js'

export * from './contracts/index.js'
export * from './http/client.js'
export * from './http/error.js'
export * from './pages.js'

export function createGooseClient(options: GooseClientOptions = {}) {
  const http = createHttpClient(options)
  const pages = createPageClient(http)
  return {
    pages,
    request<T>(path: string | URL, init?: GooseRequestInit) {
      return http.request<T>(path, init)
    },
    resolve(path: string | URL) {
      return http.resolve(path)
    },
    readInitialPayload(root?: ParentNode) {
      return readBrowserPayload(pages, root)
    },
  }
}

export type GooseClient = ReturnType<typeof createGooseClient>
