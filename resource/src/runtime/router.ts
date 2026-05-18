import { useNavigationState } from './navigation-state'
import { preloadPageComponent } from './page-registry'
import type { PagePayload } from '../types/payload'

const PREFETCH_TTL = 15_000
const pageCache = new Map<string, { promise: Promise<PagePayload>; expiresAt: number }>()
let prefetchedUrl = ''

export function installNavigation(onPage: (payload: PagePayload) => void) {
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
      const payload = await getPreparedPage(url)
      history.pushState({ goose: true, payload }, '', url)
      onPage(payload)
      window.scrollTo({ top: 0 })
    } catch {
      window.location.href = url.toString()
    } finally {
      navigation.setNavigating(false)
    }
  })

  document.addEventListener('pointerover', (event) => {
    const link = (event.target as Element | null)?.closest<HTMLAnchorElement>('a[href]')
    prefetchLink(link)
  })

  document.addEventListener('touchstart', (event) => {
    const link = (event.target as Element | null)?.closest<HTMLAnchorElement>('a[href]')
    prefetchLink(link)
  }, { passive: true })

  window.addEventListener('popstate', async (event) => {
    const cached = event.state?.payload as PagePayload | undefined
    if (cached) {
      navigation.setNavigating(true)
      preloadPageComponent(cached.component)
        .then(() => onPage(cached))
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

function prefetchLink(link: HTMLAnchorElement | null | undefined) {
  if (!link || link.target === '_blank' || link.hasAttribute('download')) return

  const url = new URL(link.href)
  if (url.origin !== window.location.origin || !isRoutablePath(url.pathname)) return
  if (url.href === window.location.href || url.href === prefetchedUrl) return

  prefetchedUrl = url.href
  getPreparedPage(url).catch(() => {
    pageCache.delete(url.href)
  })
}

async function getPreparedPage(url: URL): Promise<PagePayload> {
  const payload = await cachedFetchPage(url)
  await preloadPageComponent(payload.component)
  return payload
}

function cachedFetchPage(url: URL): Promise<PagePayload> {
  const cacheKey = url.href
  const cached = pageCache.get(cacheKey)
  if (cached && cached.expiresAt > Date.now()) {
    return cached.promise
  }

  const promise = fetchPage(url).catch((error) => {
    pageCache.delete(cacheKey)
    throw error
  })
  pageCache.set(cacheKey, {
    promise,
    expiresAt: Date.now() + PREFETCH_TTL,
  })
  return promise
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
