import { StrictMode, type ReactNode } from 'react'
import { HeaderHost } from '@gooseforum/ui/components/header-host'
import { createRoot } from 'react-dom/client'
import { flushSync } from 'react-dom'
import { commitHistory, historyIndex, isAdminPath, rememberScroll, restoreScroll, type AreaNavigation } from '@gooseforum/runtime'
import { browserPageSource, prepareDocument } from './host/browser-runtime'
import { BrowserBootstrapError } from './host/bootstrap'
import { SiteStatusNotice } from './host/site-status'

const mount = document.getElementById('goose-app')
if (!mount) throw new Error('Missing #goose-app mount element')
const root = createRoot(mount)
const deployment = document.querySelector<HTMLMetaElement>('meta[name="goose-deployment"]')?.content
let activeAdmin = isAdminPath(location.pathname)
let pending: AbortController | undefined
let activeStyle = document.querySelector<HTMLLinkElement>('#goose-layout-style')

function render(content: ReactNode) {
  flushSync(() => root.render(<StrictMode><HeaderHost>{content}</HeaderHost><div id="goose-site-status">
    <SiteStatusNotice api={browserPageSource.api} initialDeployment={deployment} admin={activeAdmin} />
  </div></StrictMode>))
}

function navigateDocument(url: URL, mode: 'push' | 'replace' | 'initial' | 'pop') {
  if (mode === 'push') location.assign(url)
  else location.replace(url)
}

const navigateArea: AreaNavigation = (url, mode, leavingIndex) => {
  pending?.abort()
  if (url.origin !== location.origin || isAdminPath(url.pathname) === activeAdmin) return false
  rememberScroll(leavingIndex)
  void openArea(url, mode)
  return true
}

async function prepareArea(url: URL, initial: boolean, signal: AbortSignal) {
  const embedded = initial ? browserPageSource.readInitial() : undefined
  const page = embedded ?? await browserPageSource.load(url, signal)
  signal.throwIfAborted()
  // Go can answer an admin URL with the login page or the forum's 404 page.
  if (page.component === 'admin.shell') {
    const entry = await import('./admin/entry')
    const allowed = entry.adminDestination(page, url)
    if (!isAdminPath(allowed.pathname)) return { redirect: allowed }
    url.href = allowed.href
    return { stylesheet: entry.stylesheet, page, prepare: () => entry.prepare(page, url) }
  }
  const destination = new URL(page.url, url)
  if (destination.pathname !== url.pathname) {
    url.pathname = destination.pathname
    url.search = destination.search
  }
  const entry = await import('./site/entry')
  return { stylesheet: entry.stylesheet, page, prepare: () => entry.prepare(page, url) }
}

async function fadeArea(direction: 'out' | 'in', signal: AbortSignal) {
  if (signal.aborted || document.hidden || !window.matchMedia('(prefers-reduced-motion: no-preference)').matches) return
  const leaving = direction === 'out'
  const animations = Array.from(document.querySelectorAll('[data-area-fade]'))
    .filter(element => element.getClientRects().length > 0)
    .map(element => element.animate(
      { opacity: leaving ? [1, 0] : [0, 1] },
      { duration: leaving ? 90 : 140, easing: leaving ? 'ease-in' : 'ease-out' },
    ))
  function cancel() {
    for (const animation of animations) animation.cancel()
  }
  signal.addEventListener('abort', cancel, { once: true })
  try {
    await Promise.allSettled(animations.map(animation => animation.finished))
  } finally {
    signal.removeEventListener('abort', cancel)
  }
}

async function openArea(url: URL, mode: 'initial' | 'push' | 'replace' | 'pop') {
  const request = new AbortController()
  pending = request
  let nextStyle: HTMLLinkElement | undefined
  try {
    const area = await prepareArea(url, mode === 'initial', request.signal)
    if (request.signal.aborted) return
    if (area.redirect) {
      navigateDocument(area.redirect, mode)
      return
    }
    const { page, stylesheet } = area
    const admin = page.component === 'admin.shell'
    const href = new URL(stylesheet, import.meta.url).href
    const sheet = activeStyle?.href === href && activeStyle.sheet ? activeStyle : document.createElement('link')
    let styleReady: Promise<void> | undefined
    if (sheet !== activeStyle) {
      nextStyle = sheet
      sheet.rel = 'stylesheet'
      sheet.crossOrigin = 'anonymous'
      sheet.href = href
      sheet.media = 'print'
      styleReady = new Promise<void>((resolve, reject) => {
        sheet.onload = () => resolve()
        sheet.onerror = () => reject(new Error('Failed to load ' + sheet.href))
        // Keep the cascade order chosen by the initial document.
        if (activeStyle) activeStyle.before(sheet)
        else document.head.append(sheet)
      })
    }
    const [content] = await Promise.all([area.prepare(), styleReady])
    if (request.signal.aborted) return
    const changingArea = mode !== 'initial' && admin !== activeAdmin
    if (changingArea) {
      await fadeArea('out', request.signal)
      if (request.signal.aborted) return
    }
    const index = historyIndex()
    if (mode === 'push' || mode === 'replace') commitHistory(url, mode)
    else if (url.href !== location.href) commitHistory(url, 'replace')
    activeAdmin = admin
    document.documentElement.toggleAttribute('data-goose-admin', admin)
    document.documentElement.toggleAttribute('data-goose-resource', !admin)
    document.documentElement.classList.toggle('goose-admin-document', admin)
    document.body.classList.toggle('bg-base-200', !admin)
    document.body.classList.toggle('text-base-content', !admin)
    document.body.classList.add('antialiased')
    if (admin) mount?.removeAttribute('aria-live')
    else mount?.setAttribute('aria-live', 'polite')
    sheet.media = 'all'
    if (activeStyle !== sheet) {
      activeStyle?.remove()
      sheet.id = 'goose-layout-style'
    }
    activeStyle = sheet
    prepareDocument(page)
    let ready = false
    function contentReady() {
      if (ready) return
      ready = true
      document.getElementById('goose-splash')?.remove()
      if (mode === 'initial') return
      restoreScroll(mode === 'pop' ? index : undefined, url.hash)
    }
    render(content(navigateArea, contentReady))
    if (admin) contentReady()
    if (changingArea) await fadeArea('in', request.signal)
  } catch (error) {
    if (request.signal.aborted) return
    if (mode !== 'initial') {
      // Let Go handle redirects, expired sessions and deployment changes.
      navigateDocument(url, mode)
      return
    }
    render(<BrowserBootstrapError error={error} onReady={() => document.getElementById('goose-splash')?.remove()} onRetry={() => location.reload()} />)
  } finally {
    if (nextStyle !== activeStyle) nextStyle?.remove()
    if (pending === request) pending = undefined
  }
}

render(null)
void openArea(new URL(location.href), 'initial')
