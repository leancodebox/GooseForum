import { readInitialPayload as readBrowserPayload } from './browser.js'
import { createHttpClient, type GooseClientOptions, type GooseRequestInit } from './http/client.js'
import { createPageClient, type PageClientOptions } from './pages.js'
import { createSiteApi } from './api/site.js'
import type { AnyPagePayload } from './contracts/page-components.js'
import type { PagePayload } from './contracts/payload.js'

export * from './contracts/index.js'
export * from './http/client.js'
export * from './http/error.js'
export * from './pages.js'
export * from './api/index.js'
export * from './browser-auth.js'
export * from './document.js'

export interface CreateGooseClientOptions extends GooseClientOptions {
  pages?: PageClientOptions
}

export function createGooseClient<TPage extends PagePayload = AnyPagePayload>(options: CreateGooseClientOptions = {}) {
  const http = createHttpClient(options)
  const pages = createPageClient<TPage>(http, options.pages)
  const api = createSiteApi(http)
  return {
    pages,
    api,
    request<T>(path: string | URL, init?: GooseRequestInit) {
      return http.request<T>(path, init)
    },
    requestWithMeta<T>(path: string | URL, init?: GooseRequestInit) {
      return http.requestWithMeta<T>(path, init)
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
