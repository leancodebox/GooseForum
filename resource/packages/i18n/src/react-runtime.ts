import i18next, { type I18nFormatModule, type Resource, type ThirdPartyModule } from 'i18next'
import { supportedLocales, type Locale } from '@gooseforum/client/i18n/locale'
import { fallbackLanguages, sourceLocale } from './index'
import { formatMessage } from './format'

const formatter: I18nFormatModule = {
  type: 'i18nFormat',
  addLookupKeys() {
    // ICU expresses plurals and selects inside the message.
  },
  parse(
    value: unknown,
    options: Record<string, unknown>,
    locale: string | undefined,
    _namespace: string | undefined,
    _key: string | undefined,
    info: { resolved?: { res?: unknown; usedLng?: string } },
  ) {
    if (typeof value !== 'string') return value
    const messageLocale = typeof info.resolved?.res === 'string' ? info.resolved.usedLng : undefined
    const params = options.replace
    return formatMessage(value, messageLocale || locale || sourceLocale,
      params && typeof params === 'object' ? { ...options, ...params } : options)
  },
}

export function createI18n(locale: Locale, resources: Resource, plugin: ThirdPartyModule) {
  const instance = i18next.createInstance().use(formatter).use(plugin)
  void instance.init({
    lng: locale,
    fallbackLng: language => [...(fallbackLanguages[language as Locale] ?? [sourceLocale])],
    supportedLngs: [...supportedLocales],
    load: 'languageOnly',
    defaultNS: 'site.auth',
    ns: [],
    resources,
    interpolation: { escapeValue: false },
    initAsync: false,
    react: { useSuspense: false, bindI18nStore: 'added' },
  })
  return instance
}
