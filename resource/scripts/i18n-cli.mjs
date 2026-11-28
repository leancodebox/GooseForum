import { spawnSync } from 'node:child_process'
import { watch } from 'node:fs'
import { mkdir, readFile, rm } from 'node:fs/promises'
import { dirname, join, resolve } from 'node:path'
import { setTimeout as delay } from 'node:timers/promises'
import { parseArgs } from 'node:util'
import {
  catalogDirectory, catalogPath, messageId, parseJson, readCatalog, requiredLocales,
  reviewRevision, revision, root, serializeNamespace, targetLocales, translationStatus, validateNamespace, writeIfChanged,
} from './i18n-catalog.mjs'
import { generate } from './i18n-generate.mjs'
import { checkServerCodes } from './check-server-message-i18n.mjs'

const help = 'Usage (from resource/):\n' +
  '  pnpm i18n add <namespace:key> --source <text> [--context <text>]\n' +
  '  pnpm i18n update <namespace:key> [--source <text>] [--context <text>]\n' +
  '  pnpm i18n remove <namespace:key>\n' +
  '  pnpm i18n rename <namespace:key> <new-namespace:key>\n' +
  '  pnpm i18n status [<namespace:key>] [--locale <locale>] [--json]\n' +
  '  pnpm i18n audit-length [--scope all|navigation] [--locale <locale>] [--limit <count>] [--json]\n' +
  '  pnpm i18n translate --locale <locale> [--json]\n' +
  '  pnpm i18n translate <namespace:key> --locale <locale> --text <text> [--revision <hash>]\n' +
  '  pnpm i18n translate <namespace:key> --locale <locale> --clear\n' +
  '  pnpm i18n translate --import <path to JSON results>\n' +
  '  pnpm i18n review <namespace:key> --locale <locale> --revision <review hash>\n' +
  '  pnpm i18n review --import <path to JSON approval records>\n' +
  '  pnpm i18n generate | watch | check [--strict]\n'

function statusRows(catalog) {
  return [...catalog].flatMap(([namespace, entries]) =>
    Object.entries(entries).flatMap(([key, entry]) =>
      targetLocales.map(locale => ({
        id: namespace + ':' + key, locale, status: translationStatus(entry, locale),
        source: entry.source, ...(entry.context && { context: entry.context }),
        sourceRevision: revision(entry),
        ...(entry.translations[locale] && {
          previous: entry.translations[locale].text,
          reviewRevision: reviewRevision(entry.translations[locale]),
        }),
      }))))
}

async function commit(catalog, changed) {
  for (const [namespace, entries] of changed) validateNamespace(namespace, entries)
  for (const [namespace, entries] of changed) {
    const path = catalogPath(namespace)
    if (Object.keys(entries).length) {
      await writeIfChanged(path, serializeNamespace(entries))
      catalog.set(namespace, entries)
    } else {
      await rm(path)
      catalog.delete(namespace)
    }
  }
}

function requireLocale(locale) {
  if (!targetLocales.includes(locale)) throw Error('Expected --locale ' + targetLocales.join('|'))
  return locale
}

function namespaceEntries(catalog, namespace) {
  return { ...catalog.get(namespace) }
}

function translationChange(catalog, changed, item, review = false) {
  const { namespace, key } = messageId(item.id)
  const entries = changed.get(namespace) ?? namespaceEntries(catalog, namespace)
  const entry = entries[key]
  if (!entry) throw Error('Unknown key: ' + item.id)
  const locale = requireLocale(item.locale)
  const current = revision(entry)
  const previous = entry.translations[locale]
  if (review) {
    if (!previous || previous.sourceRevision !== current || reviewRevision(previous) !== item.reviewRevision)
      throw Error(item.id + ': the draft is missing or has changed. Review the current draft.')
  } else if (item.sourceRevision !== current) {
    throw Error(item.id + ': source has changed. Export current work before translating.')
  }
  const translation = review ? { ...previous, reviewed: true }
    : { text: item.text, sourceRevision: current, reviewed: false }
  if (review && previous.reviewed) return
  if (!review && previous?.text === item.text && previous.sourceRevision === current) return
  entries[key] = { ...entry, translations: { ...entry.translations, [locale]: translation } }
  changed.set(namespace, entries)
}

async function verifyRemoval(catalog, changed) {
  const proposed = new Map(catalog)
  for (const [namespace, entries] of changed) {
    validateNamespace(namespace, entries)
    if (Object.keys(entries).length) proposed.set(namespace, entries)
    else proposed.delete(namespace)
  }
  await checkServerCodes(proposed)
  try {
    await generate(proposed)
    const result = spawnSync('pnpm', ['typecheck:react'], {
      cwd: root, stdio: 'inherit', env: { ...process.env, GOOSE_I18N_CHECK_CANDIDATE: '1' },
    })
    if (result.error) throw result.error
    if (result.status !== 0) throw Error('The new catalog failed type checking. Update the affected call sites first.')
  } catch (error) {
    await generate(catalog)
    throw error
  }
}

async function withCatalogLock(work) {
  const lock = join(dirname(catalogDirectory), '.catalog-lock')
  const deadline = Date.now() + 60_000
  while (true) {
    try {
      await mkdir(lock)
      break
    } catch (error) {
      if (error.code !== 'EEXIST') throw error
      if (Date.now() >= deadline) throw Error('Another i18n command is using the catalog.')
      await delay(100)
    }
  }
  try {
    return await work()
  } finally {
    await rm(lock, { recursive: true, force: true })
  }
}

async function watchCatalog() {
  await withCatalogLock(generate)
  let pending = Promise.resolve()
  let timer
  watch(catalogDirectory, { recursive: true }, (_event, filename) => {
    if (!filename?.endsWith('.json')) return
    clearTimeout(timer)
    timer = setTimeout(() => {
      pending = pending.then(() => withCatalogLock(generate)).catch(error => console.error(error.message))
    }, 80)
  })
}

async function main() {
  process.chdir(root)
  const [command, ...args] = process.argv.slice(2)
  if (!command || command === '--help' || command === 'help') return console.log(help)
  const { positionals, values } = parseArgs({
    args, allowPositionals: true,
    options: {
      source: { type: 'string' }, context: { type: 'string' }, text: { type: 'string' },
      locale: { type: 'string' }, revision: { type: 'string' }, import: { type: 'string' },
      json: { type: 'boolean' }, strict: { type: 'boolean' }, clear: { type: 'boolean' },
      scope: { type: 'string' }, limit: { type: 'string' },
    },
  })
  const accepted = {
    add: ['source', 'context'], update: ['source', 'context'], remove: [], rename: [],
    translate: ['locale', 'text', 'revision', 'import', 'json', 'clear'],
    review: ['locale', 'revision', 'import'], generate: [], check: ['strict'], status: ['locale', 'json'], watch: [],
    'audit-length': ['scope', 'locale', 'limit', 'json'],
  }
  if (!Object.hasOwn(accepted, command) || Object.keys(values).some(option => !accepted[command].includes(option)))
    throw Error(help)
  const [id, destination] = positionals
  const expected = { add: 1, update: 1, remove: 1, rename: 2, review: 1 }
  const takesOptionalId = command === 'translate' || command === 'status' || (command === 'review' && values.import !== undefined)
  if (takesOptionalId ? positionals.length > 1 : positionals.length !== (expected[command] ?? 0))
    throw Error(help)

  if (command === 'watch') return watchCatalog()
  if (command === 'generate') {
    if (process.env.GOOSE_I18N_CHECK_CANDIDATE !== '1') await withCatalogLock(generate)
    return
  }
  if (command === 'status' || (command === 'translate' && !id && !values.import)) {
    const catalog = await withCatalogLock(readCatalog)
    const locale = values.locale === undefined && command === 'status' ? undefined : requireLocale(values.locale)
    if (id) {
      const { namespace, key } = messageId(id)
      if (!catalog.get(namespace)?.[key]) throw Error('Unknown key: ' + id)
    }
    if (command === 'translate' && (values.clear || values.text !== undefined || values.revision !== undefined))
      throw Error('Specify an ID when writing or clearing a translation')
    const rows = statusRows(catalog).filter(row =>
      (!id || row.id === id) &&
      (!locale || locale === row.locale) &&
      (command === 'status' || row.status === 'missing' || row.status === 'stale'))
    if (values.json || id || command === 'translate') return console.log(JSON.stringify(rows, null, 2))
    for (const lang of locale ? [locale] : targetLocales) {
      const counts = { missing: 0, stale: 0, draft: 0, ready: 0 }
      for (const row of rows) if (row.locale === lang) counts[row.status]++
      console.log(lang, counts)
    }
    return
  }
  if (command === 'check') {
    const catalog = await withCatalogLock(readCatalog)
    await checkServerCodes(catalog)
    const required = values.strict ? targetLocales : requiredLocales
    const rows = statusRows(catalog).filter(row => required.includes(row.locale) && row.status !== 'ready')
    if (rows.length) {
      for (const row of rows.slice(0, 25)) console.error(row.locale + ' ' + row.id + ' ' + row.status)
      throw Error(rows.length + ' required translations need attention. Run pnpm i18n status for totals.')
    }
    console.log('Catalog validated: ' + [...catalog.values()].reduce((count, entries) => count + Object.keys(entries).length, 0) +
      ' messages across ' + catalog.size + ' namespaces.')
    return
  }
  if (command === 'audit-length') {
    const scope = values.scope ?? 'all'
    const limit = Number(values.limit ?? 50)
    if (scope !== 'all' && scope !== 'navigation') throw Error('Expected --scope all|navigation')
    if (!Number.isSafeInteger(limit) || limit < 1) throw Error('Expected a positive integer for --limit')
    const locale = values.locale === undefined ? undefined : requireLocale(values.locale)
    const catalog = await withCatalogLock(readCatalog)
    const { auditLength } = await import('./i18n-audit-length.mjs')
    const report = await auditLength(catalog, { scope, locale, limit })
    if (values.json) return console.log(JSON.stringify(report, null, 2))
    console.log(`Flagged ${report.flagged} of ${report.checked} translations (showing ${report.findings.length}). Widths are estimates.`)
    for (const row of report.findings) {
      console.log(`\n${row.id} [${row.locale}] ${row.width} units (expected ${row.expectedWidth}, source ${row.sourceWidth})`)
      console.log(`  ${row.signals.join('; ')}${row.navigationMedian === undefined ? '' : ` (navigation median ${row.navigationMedian})`}`)
      console.log(`  ${JSON.stringify(row.source)} → ${JSON.stringify(row.text)}`)
      for (const related of row.related) console.log(`  ${related.id}: ${JSON.stringify(related.text)}`)
    }
    return
  }

  await withCatalogLock(async () => {
    const currentCatalog = await readCatalog()
    const changed = new Map()
    if (command === 'translate' || command === 'review') {
      if (values.import !== undefined) {
        if (id || Object.keys(values).length !== 1) throw Error('--import is used on its own')
        const items = parseJson(await readFile(resolve(values.import), 'utf8'), values.import)
        if (!Array.isArray(items) || items.length === 0) throw Error('Expected a nonempty array')
        const review = command === 'review'
        const previous = new Set()
        for (const item of items) {
          const fields = review ? ['id', 'locale', 'reviewRevision'] : ['id', 'locale', 'sourceRevision', 'text']
          if (!item || fields.some(field => typeof item[field] !== 'string'))
            throw Error('Each record requires ' + fields.join(', '))
          const key = item.id + '/' + item.locale
          if (previous.has(key)) throw Error('Duplicate translation: ' + key)
          previous.add(key)
          translationChange(currentCatalog, changed, item, review)
        }
      } else {
        const { namespace, key } = messageId(id)
        const entries = namespaceEntries(currentCatalog, namespace)
        const entry = entries[key]
        if (!entry) throw Error('Unknown key: ' + id)
        if (values.clear) {
          if (values.text !== undefined || values.revision !== undefined || values.json)
            throw Error('--clear requires only an ID and --locale')
          const locale = requireLocale(values.locale)
          const translations = { ...entry.translations }
          delete translations[locale]
          entries[key] = { ...entry, translations }
          changed.set(namespace, entries)
        } else {
          if (command === 'review' && !values.revision) throw Error('--revision is required for review')
          if (command === 'translate' && (values.text === undefined || values.json))
            throw Error('Specify --text when writing a translation')
          translationChange(currentCatalog, changed, {
            id, locale: values.locale, text: values.text,
            sourceRevision: values.revision ?? revision(entry), reviewRevision: values.revision,
          }, command === 'review')
        }
      }
    } else {
      const { namespace, key } = messageId(id)
      const entries = namespaceEntries(currentCatalog, namespace)
      const entry = entries[key]
      if (command === 'add') {
        if (entry) throw Error('Duplicate key: ' + id)
        entries[key] = { source: values.source, context: values.context, translations: {} }
      } else {
        if (!entry) throw Error('Unknown key: ' + id)
        if (command === 'update') {
          if (values.source === undefined && values.context === undefined) throw Error('Specify --source or --context')
          entries[key] = { ...entry, source: values.source ?? entry.source, context: values.context ?? entry.context }
        } else {
          delete entries[key]
          if (command === 'rename') {
            const target = messageId(destination)
            if (id === destination) throw Error('The new ID must be different')
            const targetEntries = target.namespace === namespace ? entries : namespaceEntries(currentCatalog, target.namespace)
            if (targetEntries[target.key]) throw Error('Duplicate key: ' + destination)
            targetEntries[target.key] = entry
            changed.set(target.namespace, targetEntries)
          }
        }
      }
      changed.set(namespace, entries)
      if (command === 'remove' || command === 'rename') await verifyRemoval(currentCatalog, changed)
    }
    await commit(currentCatalog, changed)
    if (changed.size) await generate(await readCatalog())
  })
}

main().catch(error => {
  console.error(error.message)
  process.exitCode = 1
})
