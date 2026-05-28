import { useNavigationState } from './navigation-state'
import { resolvePageComponent } from './page-registry'
import type { Component } from 'vue'
import type { PagePayload } from '@/types/payload'
import { createRouter, createWebHistory, isNavigationFailure, type Router } from 'vue-router'

export interface PreparedPage {
  payload: PagePayload
  component: Component
}

export function installNavigation(initialPage: PreparedPage, routeComponent: Component, onPage: (page: PreparedPage) => void): Router {
  const navigation = useNavigationState()
  let initialNavigation = true
  let loadedPage = initialPage

  const router = createRouter({
    history: createWebHistory(),
    routes: [
      {
        path: '/:pathMatch(.*)*',
        component: routeComponent,
      },
    ],
  })

  router.beforeEach(async (to) => {
    if (initialNavigation) {
      initialNavigation = false
      loadedPage = initialPage
      return true
    }

    const url = new URL(to.fullPath, window.location.origin)

    navigation.setNavigating(true)
    try {
      loadedPage = await getPreparedPage(url)
      return true
    } catch {
      window.location.href = url.toString()
      return false
    }
  })

  router.afterEach((to, _from, failure) => {
    if (failure && isNavigationFailure(failure)) {
      navigation.setNavigating(false)
      return
    }
    const url = new URL(to.fullPath, window.location.origin)
    onPage(loadedPage)
    scrollAfterNavigation(url)
    navigation.setNavigating(false)
  })

  document.addEventListener('click', async (event) => {
    if (event.defaultPrevented || event.button !== 0) return
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return

    const link = (event.target as Element | null)?.closest<HTMLAnchorElement>('a[href]')
    if (!link || link.target === '_blank' || link.hasAttribute('download')) return

    const url = new URL(link.href)
    if (url.origin !== window.location.origin || !isRoutablePath(url.pathname)) return

    event.preventDefault()
    try {
      await router.push(`${url.pathname}${url.search}${url.hash}`)
    } catch {
      window.location.href = url.toString()
    }
  })

  return router
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

function scrollAfterNavigation(url: URL) {
  if (!url.hash) {
    requestAnimationFrame(() => {
      window.scrollTo({ top: 0 })
    })
    return
  }

  requestAnimationFrame(() => {
    const target = document.getElementById(decodeURIComponent(url.hash.slice(1)))
    if (target) {
      target.scrollIntoView({ block: 'start' })
      return
    }
    window.scrollTo({ top: 0 })
  })
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
