import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import {
  applyBrowserLocale,
  detectBrowserLocale,
  applyBrowserTheme,
  detectBrowserTheme,
  prepareDocument,
  pageRequestPath,
  sitePageSource,
} from './browser-runtime'

import type { PagePayload } from '@gooseforum/client'
import { detectBrowserThemePreference, saveBrowserThemePreference } from '@gooseforum/runtime/browser-host'

describe('browser locale resolution', () => {
  afterEach(() => {
    vi.restoreAllMocks()
    delete document.documentElement.dataset.defaultLocale
    history.replaceState(null, '', '/')
    document.cookie = 'lang=; path=/; max-age=0'
  })

  it('skips unsupported preferences and uses the configured site fallback', () => {
    document.documentElement.dataset.defaultLocale = 'it'
    const languages = vi.spyOn(navigator, 'languages', 'get').mockReturnValue(['fr-FR'])
    expect(detectBrowserLocale()).toBe('it')
    languages.mockReturnValue(['fr-FR', 'ja-JP'])
    expect(detectBrowserLocale()).toBe('ja')
    document.cookie = 'lang=en; path=/'
    history.replaceState(null, '', '/?lang=ko')
    expect(detectBrowserLocale()).toBe('en')
    history.replaceState(null, '', '/?lang=it')
    expect(detectBrowserLocale()).toBe('it')
    applyBrowserLocale(detectBrowserLocale('browser'), 'browser')
    expect(detectBrowserLocale()).toBe('ja')
  })
})

describe('browser theme runtime', () => {
  beforeEach(() => {
    document.documentElement.dataset.theme = 'gf-light'
    localStorage.clear()
    document.cookie = 'goose-site-theme=; path=/; max-age=0'
    document.querySelector('#goose-payload')?.remove()
  })

  afterEach(() => { vi.unstubAllEnvs(); vi.unstubAllGlobals() })

  it('defaults to the system theme without persisting the resolved color', () => {
    vi.stubGlobal('matchMedia', vi.fn(() => ({ matches: true })))
    expect(detectBrowserThemePreference()).toBe('system')
    expect(detectBrowserTheme()).toBe('gf-dark')
    applyBrowserTheme('gf-dark', undefined, false)
    expect(localStorage.getItem('goose-site-theme')).toBeNull()
  })

  it('can return from a manual choice to following the system', () => {
    applyBrowserTheme('gf-light')
    saveBrowserThemePreference('system')
    vi.stubGlobal('matchMedia', vi.fn(() => ({ matches: true })))
    expect(detectBrowserThemePreference()).toBe('system')
    expect(detectBrowserTheme()).toBe('gf-dark')
    expect(localStorage.getItem('goose-site-theme')).toBe('system')
    expect(document.cookie).toContain('goose-site-theme=system')
  })

  it('prepares a page using the system theme without overwriting the saved preference', () => {
    saveBrowserThemePreference('system')
    vi.stubGlobal('matchMedia', vi.fn(() => ({ matches: true })))
    prepareDocument({
      meta: { title: 'Test' },
      layout: { theme: { enabled: false, current: 'gf-light', themeColor: '#fbfdff' } },
    } as PagePayload)
    expect(document.documentElement.dataset.theme).toBe('gf-dark')
    expect(localStorage.getItem('goose-site-theme')).toBe('system')
    expect(document.cookie).toContain('goose-site-theme=system')
  })

  it('keeps a newly enabled site theme after the base stylesheet', () => {
    const root = document.createElement('div')
    root.id = 'goose-app'
    document.body.append(root)
    try {
      prepareDocument({
        meta: { title: 'Test' },
        layout: { theme: { enabled: true, href: '/site-theme.css?v=2', current: 'gf-light' } },
      } as PagePayload)
      expect(root.previousElementSibling?.id).toBe('goose-site-theme-link')
    } finally {
      document.getElementById('goose-site-theme-link')?.remove()
      root.remove()
    }
  })

  it('prefers a stored user theme over the static development HTML default', () => {
    localStorage.setItem('goose-site-theme', 'gf-dark')

    expect(detectBrowserTheme()).toBe('gf-dark')
  })

  it('applies the theme consistently to the document, cookie, and storage', () => {
    applyBrowserTheme('gf-dark')

    expect(document.documentElement.dataset.theme).toBe('gf-dark')
    expect(document.documentElement.style.colorScheme).toBe('dark')
    expect(localStorage.getItem('goose-site-theme')).toBe('gf-dark')
    expect(document.cookie).toContain('goose-site-theme=gf-dark')
  })
})

describe('browser page host', () => {
  beforeEach(() => document.querySelector('#goose-payload')?.remove())
  afterEach(() => { vi.unstubAllEnvs(); vi.unstubAllGlobals() })

  it('uses the Vite payload proxy only for the standalone development host', () => {
    expect(pageRequestPath('/categories?page=2', false, true)).toBe(
      '/__goose_page/categories?page=2',
    )
    expect(pageRequestPath('/categories?page=2', true, true)).toBe(
      '/categories?page=2',
    )
    expect(pageRequestPath('/categories?page=2', false, false)).toBe(
      '/categories?page=2',
    )
  })

  it('requires an embedded payload from the production Go host', () => {
    vi.stubEnv('DEV', false)
    expect(() => sitePageSource.readInitial?.()).toThrow(
      'Missing embedded GooseForum page payload',
    )
  })
})
