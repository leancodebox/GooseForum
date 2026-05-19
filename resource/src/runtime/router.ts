import { useNavigationState } from './navigation-state'
import { resolvePageComponent } from './page-registry'
import type { Component } from 'vue'
import type { PagePayload } from '@/types/payload'

export interface PreparedPage {
  payload: PagePayload
  component: Component
}

export function installNavigation(onPage: (page: PreparedPage) => void) {
  const navigation = useNavigationState()

  document.addEventListener('click', async (event) => {
    if (event.defaultPrevented || event.button !== 0) return
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return

    const link = (event.target as Element | null)?.closest<HTMLAnchorElement>('a[href]')
    if (!link || link.target === '_blank' || link.hasAttribute('download')) return

    const url = new URL(link.href)
    if (url.origin !== window.location.origin || !isRoutablePath(url.pathname)) return

    event.preventDefault()
    navigation.setNavigating(true)

    try {
      const page = await getPreparedPage(url)
      history.pushState({ goose: true, payload: page.payload }, '', url)
      onPage(page)
      window.scrollTo({ top: 0 })
    } catch {
      window.location.href = url.toString()
    } finally {
      navigation.setNavigating(false)
    }
  })

  window.addEventListener('popstate', async (event) => {
    const cached = event.state?.payload as PagePayload | undefined
    if (cached) {
      navigation.setNavigating(true)
      preparePayload(cached)
        .then(onPage)
        .finally(() => navigation.setNavigating(false))
      return
    }

    try {
      navigation.setNavigating(true)
      onPage(await getPreparedPage(new URL(window.location.href)))
    } catch {
      window.location.reload()
    } finally {
      navigation.setNavigating(false)
    }
  })
}

async function getPreparedPage(url: URL): Promise<PreparedPage> {
  return preparePayload(await fetchPage(url))
}

export async function fetchPage(url: URL): Promise<PagePayload> {
  const response = await fetch(url, {
    headers: {
      Accept: 'application/json',
      'X-Goose-Page': 'true',
    },
  })
  if (!response.ok && response.status !== 404) {
    throw new Error(`Page request failed: ${response.status}`)
  }
  return response.json() as Promise<PagePayload>
}

export async function preparePayload(payload: PagePayload): Promise<PreparedPage> {
  const component = await resolvePageComponent(payload.component)
  if (!component) {
    throw new Error(`Unknown page component: ${payload.component}`)
  }
  return { payload, component }
}

function isRoutablePath(pathname: string) {
  if (
    pathname.startsWith('/api') ||
    pathname.startsWith('/admin') ||
    pathname.startsWith('/assets') ||
    pathname.startsWith('/static') ||
    pathname.startsWith('/file')
  ) {
    return false
  }
  if (pathname === '/robots.txt' || pathname === '/sitemap.xml' || pathname === '/rss.xml') {
    return false
  }
  return !/\.[a-z0-9]+$/i.test(pathname)
}
