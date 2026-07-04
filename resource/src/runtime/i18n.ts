import { createI18n } from 'vue-i18n'
import zh from '@/locales/zh'
import en from '@/locales/en'
import ja from '@/locales/ja'
import it from '@/locales/it'

export const supportedLocales = ['zh', 'en', 'ja', 'it'] as const
export type Locale = (typeof supportedLocales)[number]

export const fallbackLocale: Locale = 'zh'

export const messages = {
  zh,
  en,
  ja,
  it,
} as const

export function normalizeLocale(value?: string | null): Locale | undefined {
  const normalized = (value || '').trim().toLowerCase()
  const short = normalized.split(/[-_,;]/)[0] as Locale
  return supportedLocales.includes(short) ? short : undefined
}

function readCookie(name: string) {
  if (typeof document === 'undefined') return ''
  return document.cookie
    .split('; ')
    .find((item) => item.startsWith(`${name}=`))
    ?.split('=')
    .slice(1)
    .join('=') || ''
}

export function detectLocale(): Locale {
  const queryLocale = typeof window !== 'undefined'
    ? normalizeLocale(new URL(window.location.href).searchParams.get('lang'))
    : undefined
  if (queryLocale) return queryLocale

  const cookieLocale = normalizeLocale(decodeURIComponent(readCookie('lang')))
  if (cookieLocale) return cookieLocale

  if (typeof navigator !== 'undefined') {
    for (const language of navigator.languages || [navigator.language]) {
      const locale = normalizeLocale(language)
      if (locale) return locale
    }
  }

  return fallbackLocale
}

export function setLocaleCookie(locale: Locale) {
  document.cookie = `lang=${locale}; path=/; max-age=31536000; samesite=lax`
}

export const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale: detectLocale(),
  fallbackLocale,
  messages,
  missingWarn: false,
  fallbackWarn: false,
})

export function setLocale(locale: Locale) {
  i18n.global.locale.value = locale
  setLocaleCookie(locale)
  document.documentElement.lang = locale
}

export function currentLocale() {
  return i18n.global.locale.value as Locale
}
