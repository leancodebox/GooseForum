import { computed, ref } from 'vue'
import type { ThemePayload } from '@/types/payload'

export type SiteTheme = 'gf-light' | 'gf-dark'

const STORAGE_KEY = 'goose-site-theme'
const COOKIE_KEY = 'goose-site-theme'
const themes: SiteTheme[] = ['gf-light', 'gf-dark']
const THEME_LINK_ID = 'goose-site-theme-link'
const THEME_PREVIEW_STYLE_ID = 'goose-site-theme-preview'
const themeColors: Record<SiteTheme, string> = {
  'gf-light': '#ffffff',
  'gf-dark': '#101010',
}
const currentTheme = ref<SiteTheme>(resolveInitialTheme())

export function useSiteTheme() {
  const isDark = computed(() => currentTheme.value === 'gf-dark')

  return {
    theme: currentTheme,
    isDark,
    toggleTheme,
  }
}

export function applyStoredTheme() {
  applyTheme(currentTheme.value)
}

export function applySiteThemePayload(theme?: ThemePayload) {
  for (const name of themes) {
    const color = theme?.enabled ? theme.colors?.[name] : undefined
    themeColors[name] = normalizeThemeColor(color) || (name === 'gf-dark' ? '#101010' : '#ffffff')
  }
  applySiteThemeLink(theme?.enabled ? theme.href : '')
  applyTheme(currentTheme.value)
}

export function applySiteThemeCss(css: string) {
  let el = document.getElementById(THEME_PREVIEW_STYLE_ID) as HTMLStyleElement | null
  if (!css) {
    el?.remove()
    return
  }
  if (!el) {
    el = document.createElement('style')
    el.id = THEME_PREVIEW_STYLE_ID
  }
  document.head.appendChild(el)
  el.textContent = css
}

function applySiteThemeLink(href?: string) {
  let el = document.getElementById(THEME_LINK_ID) as HTMLLinkElement | null
  if (!href) {
    el?.remove()
    return
  }
  if (!el) {
    el = document.createElement('link')
    el.id = THEME_LINK_ID
    el.rel = 'stylesheet'
  }
  if (el.href !== new URL(href, window.location.origin).href) {
    el.href = href
  }
  document.head.appendChild(el)
}

export function toggleTheme() {
  setTheme(currentTheme.value === 'gf-dark' ? 'gf-light' : 'gf-dark')
}

export function setTheme(theme: SiteTheme) {
  currentTheme.value = theme
  applyTheme(theme)
  writeThemeCookie(theme)
  try {
    window.localStorage.setItem(STORAGE_KEY, theme)
  } catch {
    // Ignore storage failures in private or restricted browsing contexts.
  }
}

function resolveInitialTheme(): SiteTheme {
  const documentTheme = document.documentElement.dataset.theme || null
  if (isSiteTheme(documentTheme)) return documentTheme

  try {
    const cookieTheme = readThemeCookie()
    if (isSiteTheme(cookieTheme)) return cookieTheme
  } catch {
    // Fall through to local storage compatibility.
  }

  try {
    const stored = window.localStorage.getItem(STORAGE_KEY)
    if (isSiteTheme(stored)) {
      writeThemeCookie(stored)
      return stored
    }
  } catch {
    // Fall through to light theme.
  }

  return 'gf-light'
}

function applyTheme(theme: SiteTheme) {
  document.documentElement.dataset.theme = theme
  document.querySelector('meta[name="theme-color"]')?.setAttribute('content', themeColors[theme])
}

function isSiteTheme(value: string | null): value is SiteTheme {
  return themes.includes(value as SiteTheme)
}

function readThemeCookie() {
  return document.cookie
    .split('; ')
    .find((item) => item.startsWith(`${COOKIE_KEY}=`))
    ?.split('=')
    .slice(1)
    .join('=') || ''
}

function writeThemeCookie(theme: SiteTheme) {
  document.cookie = `${COOKIE_KEY}=${theme}; path=/; max-age=31536000; samesite=lax`
}

function normalizeThemeColor(value?: string) {
  if (!value) return ''
  return /^#[0-9a-fA-F]{6}$/.test(value) ? value : ''
}
