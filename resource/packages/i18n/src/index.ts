import type {} from './i18next'
import type { ParseKeysWithoutVariables, Resource, ResourceKey } from 'i18next'
import type { Locale } from '@gooseforum/client/i18n/locale'
import type Resources from '../generated/resources'
import { fallbackLanguages as fallback, icuNamespaces, needsFallback, sourceLocale } from '../generated/fallback'
import { prepareIcuFormat } from './format'

export { sourceLocale }
export const fallbackLanguages: Record<Locale, readonly Locale[]> = fallback

export type Namespace = keyof Resources
export type SimpleKey<N extends Namespace> = ParseKeysWithoutVariables<N>
export type Dictionary<N extends Namespace> = MessageStrings<Resources[N]>
type MessageStrings<T> = { [K in keyof T]: T[K] extends string ? string : MessageStrings<T[K]> }

type Listener<N> = (locale: Locale, namespace: N, dictionary: ResourceKey) => void
type Loader = () => Promise<{ default: ResourceKey }>

export function createTranslationStore<N extends Namespace>(loaders: Record<string, Loader | undefined>) {
  const resources: Resource = {}
  const pending = new Map<string, Promise<void>>()
  const listeners = new Set<Listener<N>>()

  function loadNamespace(locale: Locale, namespace: N): Promise<void> {
    if (resources[locale]?.[namespace]) return Promise.resolve()
    const id = locale + '/' + namespace
    const current = pending.get(id)
    if (current) return current
    const loader = loaders[id]
    if (!loader) return Promise.reject(new Error('Unknown translation namespace: ' + namespace))
    const formatter = icuNamespaces[locale]?.includes(namespace) ? prepareIcuFormat() : undefined
    const promise = Promise.all([loader(), formatter]).then(([{ default: dictionary }]) => {
      (resources[locale] ??= {})[namespace] = dictionary
      for (const listener of listeners) listener(locale, namespace, dictionary)
    }).finally(() => pending.delete(id))
    pending.set(id, promise)
    return promise
  }

  return {
    async prepare(locale: Locale, namespaces: readonly N[]) {
      const missing = needsFallback[locale]
      await Promise.all([...new Set(namespaces)].flatMap(namespace => {
        const languages: Locale[] = [locale]
        if (locale !== sourceLocale) languages.push(sourceLocale)
        if (locale !== 'en' && missing?.includes(namespace) && fallbackLanguages[locale].includes('en')) languages.push('en')
        return languages.map(language => loadNamespace(language, namespace))
      }))
    },
    snapshot: () => structuredClone(resources),
    loadedNamespaces: () => [...new Set(Object.values(resources).flatMap(value => Object.keys(value)))] as N[],
    subscribe(listener: Listener<N>) {
      listeners.add(listener)
      for (const [locale, namespaces] of Object.entries(resources))
        for (const [namespace, dictionary] of Object.entries(namespaces))
          listener(locale as Locale, namespace as N, dictionary)
      return () => { listeners.delete(listener) }
    },
    read(locale: Locale, namespace: N) { return resources[locale]?.[namespace] },
  }
}

// Only use for keys supplied by a checked runtime contract or user data.
// Calls with catalog keys use i18next's namespace, key and argument types.
export function runtimeText(
  t: (...args: never[]) => unknown,
  key: string,
  options?: Record<string, unknown>,
): string {
  return (t as unknown as (key: string, options?: Record<string, unknown>) => string)(key, options)
}

type TrustLevelKey = Extract<SimpleKey<'shared.shell'>, `trustLevels.${string}`>
export function trustLevelLabel(level: number, text: (key: TrustLevelKey) => string): string {
  switch (level) {
    case 0: return text('trustLevels.newUser')
    case 100: return text('trustLevels.member')
    case 200: return text('trustLevels.regular')
    case 1000: return text('trustLevels.communityLeader')
    default: return String(level)
  }
}
