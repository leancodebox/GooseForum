import type { IntlMessageFormat as Formatter } from 'intl-messageformat'

let handle: Promise<void> | undefined
let MessageFormat: typeof Formatter | undefined
const formats = new Map<string, Map<string, Formatter>>()

export function prepareIcuFormat() {
  return handle ??= import('intl-messageformat').then(module => {
    MessageFormat = module.IntlMessageFormat
  }).catch(error => {
    handle = undefined
    throw error
  })
}

export function formatMessage(message: string, locale: string, params: Record<string, unknown> = {}): string {
  if (!/[{}']/.test(message)) return message
  if (!MessageFormat)
    return message.replace(/\{([A-Za-z0-9_]+)\}/g, (match, name: string) =>
      Object.hasOwn(params, name) ? String(params[name] ?? '') : match)
  try {
    let cache = formats.get(locale)
    if (!cache) {
      cache = new Map()
      formats.set(locale, cache)
    }
    let formatter = cache.get(message)
    if (!formatter) {
      formatter = new MessageFormat(message, locale)
      cache.set(message, formatter)
    }
    return String(formatter.format(params as Parameters<Formatter['format']>[0]))
  } catch {
    return message
  }
}
