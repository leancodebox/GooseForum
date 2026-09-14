import { createI18n } from 'vue-i18n'
import zh from '@/locales/zh'
import en from '@/locales/en'
import ja from '@/locales/ja'
import it from '@/locales/it'
import { authResources, normalizeLocale, oidcConsentResources, supportedLocales, type Locale } from '@gooseforum/client/i18n'

export { normalizeLocale, supportedLocales, type Locale } from '@gooseforum/client/i18n'

export const fallbackLocale: Locale = 'zh'

export const messages = {
  zh: withSharedAuth(zh, 'zh'),
  en: withSharedAuth(en, 'en'),
  ja: withSharedAuth(ja, 'ja'),
  it: withSharedAuth(it, 'it'),
} as const

function withSharedAuth<T extends { auth: Record<string, unknown> }>(messages: T, locale: Locale) {
  return {
    ...messages,
    auth: {
      ...messages.auth,
      ...authResources[locale],
      validation: authResources[locale].validation,
    },
    oidcConsent: oidcConsentResources[locale],
  }
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
