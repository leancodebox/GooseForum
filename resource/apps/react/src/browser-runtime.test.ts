import { beforeEach, describe, expect, it } from 'vitest'
import { applyBrowserTheme, detectBrowserTheme } from './browser-runtime'

describe('browser theme runtime', () => {
  beforeEach(() => {
    document.documentElement.dataset.theme = 'gf-light'
    localStorage.clear()
    document.cookie = 'goose-site-theme=; path=/; max-age=0'
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
