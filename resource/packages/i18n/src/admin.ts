import type { Locale } from '@gooseforum/client/i18n/locale'
import type { InterpolationMap, ParseKeys } from 'i18next'
import type Resources from '../generated/resources'
import { loaders, namespaces } from '../generated/admin'
import { loaders as sharedLoaders, namespaces as sharedNamespaces } from '../generated/shared'
import { createTranslationStore, fallbackLanguages, trustLevelLabel, type Dictionary } from './index'
import { formatMessage } from './format'

export type { Dictionary } from './index'
export const adminLoaders = { ...loaders, ...sharedLoaders }
export const adminNamespaces = [...namespaces, ...sharedNamespaces]
export type AdminNamespace = (typeof adminNamespaces)[number]
export const adminTranslations = createTranslationStore<AdminNamespace>(adminLoaders)

type MessageAt<T, K extends string> = K extends keyof T ? T[K]
  : K extends `${infer Head}.${infer Tail}` ? Head extends keyof T ? MessageAt<T[Head], Tail> : never
  : never
type Params<N extends AdminNamespace, K extends string> = InterpolationMap<MessageAt<Resources[N], K>>
type AdminKey<N extends AdminNamespace> = Exclude<ParseKeys<N>, `${string}:${string}`>
export type AdminText<N extends AdminNamespace> = <K extends AdminKey<N>>(
  key: K,
  ...args: Record<never, never> extends Params<N, K> ? [params?: Params<N, K>] : [params: Params<N, K>]
) => string

export function adminText<N extends AdminNamespace>(locale: Locale, namespace: N): AdminText<N> {
  const languages = [locale, ...fallbackLanguages[locale]]
  const translate = (key: string, options: Record<string, unknown> = {}) => {
    for (const language of languages) {
      const dictionary = adminTranslations.read(language, namespace)
      const message = lookup(dictionary, key)
      if (message !== undefined) {
        const params = options.replace
        return formatMessage(message, language, params && typeof params === 'object' ? { ...options, ...params } : options)
      }
    }
    return typeof options.defaultValue === 'string' ? options.defaultValue : key
  }
  return translate as AdminText<N>
}

function lookup(dictionary: unknown, key: string): string | undefined {
  if (!dictionary || typeof dictionary !== 'object') return
  const record = dictionary as Record<string, unknown>
  if (Object.hasOwn(record, key) && typeof record[key] === 'string') return record[key]
  const head = key.indexOf('.')
  if (head !== -1 && Object.hasOwn(record, key.slice(0, head)))
    return lookup(record[key.slice(0, head)], key.slice(head + 1))
}

function overlay(base: Record<string, unknown>, values: Record<string, unknown>): Record<string, unknown> {
  const result = { ...base }
  for (const [key, value] of Object.entries(values)) {
    result[key] = value && typeof value === 'object'
      ? overlay((result[key] ?? {}) as Record<string, unknown>, value as Record<string, unknown>)
      : value
  }
  return result
}

export function getAdminDictionary<N extends AdminNamespace>(namespace: N, locale: Locale): Dictionary<N> {
  const resource = adminTranslations.read(locale, namespace)
  if (!resource) throw new Error('Admin translations were not prepared: ' + locale + '/' + namespace)
  const sources = [...fallbackLanguages[locale]].reverse().map(fallback =>
    adminTranslations.read(fallback, namespace))
  const dictionary = [...sources, resource].reduce<Record<string, unknown>>((result, values) =>
    values && typeof values === 'object' ? overlay(result, values) : result, {})
  return dictionary as Dictionary<N>
}

export function trustLevelName(level: number, locale: Locale): string {
  return trustLevelLabel(level, adminText(locale, 'shared.shell'))
}
