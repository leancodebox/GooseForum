import {
  createGooseClient,
  GooseProtocolError,
  type AnyPagePayload,
  type PagePayload,
} from '@gooseforum/client'
import type { PageSource } from '@gooseforum/runtime/page-source'
export {
  applyBrowserLocale,
  applyBrowserTheme,
  detectBrowserLocale,
  detectBrowserTheme,
  prepareBrowserDocument as prepareDocument,
  queueBrowserFlash,
} from '@gooseforum/runtime/browser-host'

type BrowserPage = AnyPagePayload | PagePayload<unknown, 'admin.shell'>
const client = createGooseClient<BrowserPage>({ pages: { components: ['admin.shell'] } })

export const browserPageSource = {
  api: client.api,
  admin: client.admin,
  readInitial() {
    if (hasEmbeddedPagePayload()) return client.readInitialPayload()
    if (import.meta.env.DEV) return undefined
    throw new GooseProtocolError('Missing embedded GooseForum page payload')
  },
  load(url: URL, signal?: AbortSignal) {
    return client.pages.fetch(pageRequestPath(url.pathname + url.search), { signal })
  },
}

export const sitePageSource: PageSource = {
  api: client.api,
  async load(url, signal) {
    const page = await browserPageSource.load(url, signal)
    if (page.component === 'admin.shell') throw new GooseProtocolError('Expected a forum page')
    return page
  },
}

function hasEmbeddedPagePayload() {
  return Boolean(document.querySelector('#goose-payload'))
}

export function pageRequestPath(
  pageUrl: string,
  embedded = hasEmbeddedPagePayload(),
  development = import.meta.env.DEV,
) {
  return embedded || !development ? pageUrl : `/__goose_page${pageUrl}`
}
