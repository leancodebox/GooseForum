import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import {
  applyBrowserTheme,
  detectBrowserTheme,
  pageRequestPath,
  sitePageSource,
} from './browser-runtime'

describe('browser theme runtime', () => {
  beforeEach(() => {
    document.documentElement.dataset.theme = 'gf-light'
    localStorage.clear()
    document.cookie = 'goose-site-theme=; path=/; max-age=0'
    document.querySelector('#goose-payload')?.remove()
  })

  afterEach(() => vi.unstubAllEnvs())

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
  afterEach(() => vi.unstubAllEnvs())

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
