import {
  createGooseClient,
  updateDocumentMetadata,
  type AnyPagePayload,
  type GooseSiteApi,
  type GooseAdminApi,
  type PagePayload,
  type ThemePayload,
} from '@gooseforum/client'
import { authLocales, type AuthLocale } from '@gooseforum/react/i18n/auth'
import type { GooseFlashType } from '@gooseforum/react/runtime'

const flashStorageKey = 'goose:flash-messages'
const themeStorageKey = 'goose-site-theme'

export interface PageSource<TPage extends PagePayload> {
  api: GooseSiteApi
  admin: GooseAdminApi
  load(url: URL, signal?: AbortSignal): Promise<TPage>
}

function createDevelopmentPageSource<TPage extends PagePayload>(
  components?: readonly string[],
): PageSource<TPage> {
  const client = createGooseClient<TPage>({
    pages: components ? { components } : undefined,
  })

  return {
    api: client.api,
    admin: client.admin,
    load(url, signal) {
      const pageUrl = `${url.pathname}${url.search}`
      return client.pages.fetch(`/__goose_page${pageUrl}`, { signal })
    },
  }
}

export const sitePageSource = createDevelopmentPageSource<AnyPagePayload>()
export const adminPageSource = createDevelopmentPageSource<PagePayload>(['admin.shell'])

export function detectBrowserLocale(): AuthLocale {
  const queryLocale = normalizeLocale(new URL(window.location.href).searchParams.get('lang'))
  if (queryLocale) return queryLocale

  const cookie = document.cookie
    .split('; ')
    .find((item) => item.startsWith('lang='))
    ?.slice('lang='.length)
  const cookieLocale = normalizeLocale(cookie ? decodeURIComponent(cookie) : undefined)
  if (cookieLocale) return cookieLocale

  for (const language of navigator.languages || [navigator.language]) {
    const locale = normalizeLocale(language)
    if (locale) return locale
  }
  return 'zh'
}

export function applyBrowserLocale(locale: AuthLocale) {
  document.documentElement.lang = locale
  document.cookie = `lang=${locale}; path=/; max-age=31536000; samesite=lax`
}

export function queueBrowserFlash(message: string, type: GooseFlashType = 'info') {
  const text = message.trim()
  if (!text) return
  try {
    const stored = window.sessionStorage.getItem(flashStorageKey)
    const parsed = stored ? JSON.parse(stored) : []
    const messages = Array.isArray(parsed) ? parsed : []
    messages.push({ type, message: text })
    window.sessionStorage.setItem(flashStorageKey, JSON.stringify(messages.slice(-4)))
  } catch {
    // Storage can be unavailable in privacy modes. The destination page can
    // still load normally; it simply will not receive the queued message.
  }
}

export function detectBrowserTheme(): ThemePayload['current'] {
  const cookieTheme = document.cookie
    .split('; ')
    .find((item) => item.startsWith(`${themeStorageKey}=`))
    ?.slice(themeStorageKey.length + 1)
  if (cookieTheme === 'gf-light' || cookieTheme === 'gf-dark') return cookieTheme
  try {
    const stored = localStorage.getItem(themeStorageKey)
    if (stored === 'gf-light' || stored === 'gf-dark') return stored
  } catch {
    // Fall through to the default theme.
  }
  const documentTheme = document.documentElement.dataset.theme
  if (documentTheme === 'gf-light' || documentTheme === 'gf-dark') return documentTheme
  return 'gf-light'
}

export function applyBrowserTheme(theme: ThemePayload['current'], colors?: Record<string, string>) {
  document.documentElement.dataset.theme = theme
  document.documentElement.style.colorScheme = theme === 'gf-dark' ? 'dark' : 'light'
  document.querySelector('meta[name="theme-color"]')?.setAttribute('content', colors?.[theme] || (theme === 'gf-dark' ? '#101010' : '#fbfdff'))
  document.cookie = `${themeStorageKey}=${theme}; path=/; max-age=31536000; samesite=lax`
  try {
    localStorage.setItem(themeStorageKey, theme)
  } catch {
    // Ignore storage failures in restricted browsing modes.
  }
}

export function prepareDocument(payload: PagePayload, preferredTheme?: ThemePayload['current']) {
  updateDocumentMetadata(payload)
  document.documentElement.lang ||= 'zh-CN'
  applyTheme(payload.layout.theme, preferredTheme)
}

function applyTheme(theme: ThemePayload, preferredTheme = theme.current) {
  applyBrowserTheme(preferredTheme, theme.colors)

  let link = document.querySelector<HTMLLinkElement>('#goose-site-theme-link')
  if (!theme.enabled || !theme.href) {
    link?.remove()
    return
  }
  if (!link) {
    link = document.createElement('link')
    link.id = 'goose-site-theme-link'
    link.rel = 'stylesheet'
    document.head.appendChild(link)
  }
  link.href = theme.href
}

function normalizeLocale(value?: string | null): AuthLocale | undefined {
  const short = (value || '').trim().toLowerCase().split(/[-_,;]/)[0] as AuthLocale
  return authLocales.includes(short) ? short : undefined
}
