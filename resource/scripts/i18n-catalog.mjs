import { createHash } from 'node:crypto'
import { mkdir, readFile, readdir, rename, rm, writeFile } from 'node:fs/promises'
import { join, resolve } from 'node:path'
import { parse as parseMessage, TYPE } from '@formatjs/icu-messageformat-parser'
import { getNodeValue, parseTree } from 'jsonc-parser'
import { supportedLocales } from '../packages/client/src/i18n/locale.ts'

export const root = resolve(import.meta.dirname, '..')
export const packageRoot = join(root, 'packages/i18n')
export const catalogDirectory = join(packageRoot, 'catalog')
export const generatedDirectory = join(packageRoot, 'generated')
export const sourceLocale = 'zh'
export const serverNamespace = 'shared.serverMessages'
export const locales = supportedLocales
export const targetLocales = locales.filter(locale => locale !== sourceLocale)
export const fallbackLanguages = { en: ['zh'], ja: ['en', 'zh'], it: ['en', 'zh'], ru: ['zh'], ko: ['en', 'zh'], fr: ['en', 'zh'], zh: [] }
export const requiredLocales = targetLocales

const segment = /^[A-Za-z0-9_-]+$/
const reserved = new Set(['prototype', ...Object.getOwnPropertyNames(Object.prototype)])
const isSegment = value => segment.test(value) && !reserved.has(value)
const isNamespace = value => /^(site|admin|shared)\.[a-zA-Z][\w-]*$/.test(value)
const isKey = value => value.split('.').every(isSegment)
const object = value => value !== null && typeof value === 'object' && !Array.isArray(value)

export function messageId(id) {
  const [namespace, key, ...extra] = id.split(':')
  if (!isNamespace(namespace) || !key || !isKey(key) || extra.length)
    throw Error('Expected an ID such as site.topic:replyCount')
  return { namespace, key }
}

export function revision(entry) {
  return fingerprint([entry.source, entry.context ?? ''])
}

function fingerprint(fields) {
  return createHash('sha256').update(JSON.stringify(fields)).digest('hex')
}

export function reviewRevision(translation) {
  return fingerprint([translation.sourceRevision, translation.text])
}

export function translationStatus(entry, locale) {
  const translation = entry.translations[locale]
  if (!translation) return 'missing'
  if (translation.sourceRevision !== revision(entry)) return 'stale'
  return translation.reviewed ? 'ready' : 'draft'
}

export function parseJson(input, file) {
  const errors = []
  const tree = parseTree(input, errors, { disallowComments: true })
  if (errors.length || !tree) throw Error(file + ': invalid JSON')
  function visit(node) {
    if (node.type === 'object') {
      const seen = new Set()
      for (const [keyNode, valueNode] of node.children.map(child => child.children)) {
        const key = keyNode.value
        if (seen.has(key)) throw Error(file + ': duplicate property ' + key)
        if (reserved.has(key)) throw Error(file + ': reserved property ' + key)
        seen.add(key)
        visit(valueNode)
      }
    } else if (node.type === 'array') node.children?.forEach(visit)
  }
  visit(tree)
  return getNodeValue(tree)
}

export function requiresIcu(message) {
  const nodes = parseMessage(message)
  const plain = nodes.map(node => node.type === TYPE.literal ? node.value
    : node.type === TYPE.argument && /^[A-Za-z0-9_]+$/.test(node.value) ? '{' + node.value + '}'
    : null)
  return plain.includes(null) || plain.join('') !== message
}

function messageArguments(text, locale) {
  if (typeof text !== 'string' || !text.trim()) throw Error('message must be a nonempty string')
  const values = new Map()
  function visit(elements) {
    for (const element of elements) {
      if (element.type === TYPE.tag) throw Error('markup requires an explicit component interface')
      if (!('value' in element) || element.type === TYPE.literal) continue
      const type = element.type === TYPE.number || element.type === TYPE.plural ? 'number'
        : element.type === TYPE.date || element.type === TYPE.time ? 'date'
        : element.type === TYPE.select ? 'select'
        : 'argument'
      const previous = values.get(element.value)
      if (previous && previous !== type && previous !== 'argument' && type !== 'argument')
        throw Error('incompatible formats for ' + element.value)
      if (!previous || previous === 'argument') values.set(element.value, type)
      if (element.type === TYPE.plural) {
        const categories = new Intl.PluralRules(locale, { type: element.pluralType }).resolvedOptions().pluralCategories
        for (const category of Object.keys(element.options))
          if (!category.startsWith('=') && !categories.includes(category))
            throw Error('invalid ' + locale + ' plural category ' + category)
      }
      if ('options' in element)
        for (const option of Object.values(element.options)) visit(option.value)
    }
  }
  visit(parseMessage(text))
  return values
}

export function validateEntry(id, entry) {
  if (!object(entry) || !object(entry.translations) ||
      Object.keys(entry).some(key => !['source', 'context', 'translations'].includes(key)) ||
      (entry.context !== undefined && typeof entry.context !== 'string')) throw Error(id + ': invalid entry')
  try {
    const source = messageArguments(entry.source, sourceLocale)
    const current = revision(entry)
    for (const [locale, translation] of Object.entries(entry.translations)) {
      if (!targetLocales.includes(locale) || !object(translation) ||
          Object.keys(translation).length !== 3 ||
          typeof translation.reviewed !== 'boolean' ||
          !/^[a-f0-9]{64}$/.test(translation.sourceRevision))
        throw Error('invalid translation metadata for ' + locale)
      const args = messageArguments(translation.text, locale)
      if (translation.sourceRevision === current &&
          (args.size !== source.size || [...args].some(([name, type]) => !source.has(name) ||
            (type !== 'argument' && type !== source.get(name)))))
        throw Error(locale + ' arguments do not match the source')
    }
  } catch (error) {
    throw Error(id + ': ' + error.message, { cause: error })
  }
}

export function validateNamespace(namespace, entries) {
  if (!isNamespace(namespace) || !object(entries)) throw Error(namespace + ': invalid namespace')
  const keys = new Set(Object.keys(entries))
  for (const [key, entry] of Object.entries(entries)) {
    const id = namespace + ':' + key
    messageId(id)
    const segments = key.split('.')
    for (let i = 1; namespace !== serverNamespace && i < segments.length; i++)
      if (keys.has(segments.slice(0, i).join('.'))) throw Error(id + ': conflicts with a parent key')
    validateEntry(id, entry)
  }
}

export async function readCatalog(directory = catalogDirectory) {
  const catalog = new Map()
  for (const group of ['site', 'admin', 'shared']) {
    const folder = join(directory, group)
    for (const file of (await readdir(folder)).sort()) {
      if (!file.endsWith('.json')) continue
      const namespace = group + '.' + file.slice(0, -5)
      const path = join(folder, file)
      const entries = parseJson(await readFile(path, 'utf8'), path)
      validateNamespace(namespace, entries)
      if (!Object.keys(entries).length) throw Error(namespace + ': empty namespace')
      catalog.set(namespace, entries)
    }
  }
  return catalog
}

export function catalogPath(namespace, directory = catalogDirectory) {
  const [group, name] = namespace.split('.')
  return join(directory, group, name + '.json')
}

export async function writeIfChanged(path, data) {
  try {
    if (await readFile(path, 'utf8') === data) return
  } catch (error) {
    if (error.code !== 'ENOENT') throw error
  }
  await mkdir(resolve(path, '..'), { recursive: true })
  const temp = path + '.' + process.pid + '.' + crypto.randomUUID() + '.tmp'
  try {
    await writeFile(temp, data, { flag: 'wx' })
    await rename(temp, path)
  } finally {
    await rm(temp, { force: true })
  }
}

export function serializeNamespace(entries) {
  return '{\n' + Object.keys(entries).sort().map(key => {
    const { source, context, translations } = entries[key]
    const values = [
      '    "source": ' + JSON.stringify(source),
      ...(context ? ['    "context": ' + JSON.stringify(context)] : []),
    ]
    const lines = targetLocales.filter(locale => translations[locale]).map(locale =>
      '      ' + JSON.stringify(locale) + ': ' + JSON.stringify(translations[locale]))
    values.push(lines.length ? '    "translations": {\n' + lines.join(',\n') + '\n    }' : '    "translations": {}')
    return '  ' + JSON.stringify(key) + ': {\n' + values.join(',\n') + '\n  }'
  }).join(',\n') + '\n}\n'
}
