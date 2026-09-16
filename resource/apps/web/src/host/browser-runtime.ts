import {
  createGooseClient,
  GooseProtocolError,
  type AnyPagePayload,
  type GooseAdminApi,
  type PagePayload,
} from '@gooseforum/client'
import type { PageSource as RuntimePageSource } from '@gooseforum/runtime/page-source'
export {
  applyBrowserLocale,
  applyBrowserTheme,
  detectBrowserLocale,
  detectBrowserTheme,
  prepareBrowserDocument as prepareDocument,
  queueBrowserFlash,
} from '@gooseforum/runtime/browser-host'

export interface PageSource<TPage extends PagePayload>
  extends RuntimePageSource<TPage> {
  admin: GooseAdminApi
}

function createBrowserPageSource<TPage extends PagePayload>(
  components?: readonly string[],
): PageSource<TPage> {
  const client = createGooseClient<TPage>({
    pages: components ? { components } : undefined,
  })

  return {
    api: client.api,
    admin: client.admin,
    readInitial() {
      if (hasEmbeddedPagePayload()) return client.readInitialPayload()
      if (import.meta.env.DEV) return undefined
      throw new GooseProtocolError('Missing embedded GooseForum page payload')
    },
    load(url, signal) {
      const pageUrl = `${url.pathname}${url.search}`
      return client.pages.fetch(pageRequestPath(pageUrl), { signal })
    },
  }
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

export const sitePageSource = createBrowserPageSource<AnyPagePayload>()
export const adminPageSource = createBrowserPageSource<PagePayload>(['admin.shell'])
