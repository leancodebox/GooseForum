// pnpm audit:i18n [--base REV] [--history] [--all] [--groups] [--json]
import { execFileSync } from 'node:child_process'
import { existsSync, readFileSync } from 'node:fs'
import { resolve, dirname } from 'node:path'
import { parseArgs } from 'node:util'
import { admin, client, catalog, flatten, literals, root, server } from './i18n-catalog.mjs'
import { groupUnresolved } from './i18n-triage.mjs'
import { adminDictionaryReads, inlineStrings, references } from './i18n-references.mjs'

const { values: options } = parseArgs({ options: {
  base: { type: 'string' },
  groups: { type: 'boolean' }, history: { type: 'boolean' }, all: { type: 'boolean' }, json: { type: 'boolean' },
} })
const git = (...args) => execFileSync('git', args, { cwd: root, encoding: 'utf8', maxBuffer: 32 << 20 })
const files = git('ls-files', '-co', '--exclude-standard', '-z').split('\0').filter(f => f && existsSync(resolve(root, f)))
const read = file => readFileSync(resolve(root, file), 'utf8')
const { rows, locales } = await catalog()
const errors = []
const origins = new Map()
const fallback = Object.fromEntries(locales.map(locale => [locale, 0]))
const explicit = new Map()
const dynamic = new Map()
const lookupCandidates = new Map()
function remember(map, key, value) {
  if (!map.has(key)) map.set(key, [])
  map.get(key).push(value)
}
const placeholders = value => [...new Set(value.match(/\{[\w]+\}/g) ?? [])].sort().join(',')
for (const [id, translations] of rows) {
  for (const locale of locales) {
    const entry = translations[locale]
    if (!entry) { errors.push(`${id}: missing ${locale}`); continue }
    if (!entry.value.trim()) errors.push(`${id}: empty ${locale}`)
    if (translations.en && placeholders(entry.value) !== placeholders(translations.en.value))
      errors.push(`${id}: ${locale} placeholders [${placeholders(entry.value)}], en [${placeholders(translations.en.value)}]`)
    if (entry.file?.includes('/messages/en-') && locale !== 'en') fallback[locale]++
    if (locale === 'zh') remember(origins, entry.file + '#' + entry.key, id)
  }
}
const excluded = /(?:^|\/)(?:test|tests|e2e|__tests__|messages)\/|\.(?:spec|test)\.|_test\.go$/
const frontend = /^resource\/(?:packages|apps)\/[^/]+\/src\/.*\.tsx?$/
const backend = /^(app\/.*\.go|(?:app|resource)\/.*\.(?:gohtml|html))$/
function production(file) {
  return !excluded.test(file) && (frontend.test(file) || backend.test(file))
}
function adminHelpers(paths, readSource) {
  const helpers = new Map()
  for (const file of paths) {
    if (!file.startsWith(admin + '/') || !/i18n\.ts$/.test(file)) continue
    const namespace = readSource(file).match(/getAdminDictionary<[^>]+>\(["']([^"']+)/)?.[1]
    if (namespace) helpers.set(resolve(root, file.slice(0, -3)), namespace)
  }
  return helpers
}
function scan(file, source, helpers) {
  if (frontend.test(file)) {
    const hits = references(source, file, specifier => helpers.get(resolve(root, dirname(file), specifier)))
    if (file.startsWith(admin + '/')) hits.push(...adminDictionaryReads(source, file, rows))
    return hits
  }
  const pattern = file.endsWith('.go') ? /\bi18n\.T\([^,\n]+,\s*"([^"]+)"/g : /\b(?:t\s+\S+|call\s+\.T)\s+"([^"]+)"/g
  return [...source.matchAll(pattern)].map(m => ({ id: 'go:' + m[1], at: file + ':' + source.slice(0, m.index).split('\n').length }))
}
const currentHelpers = adminHelpers(files, read)
for (const file of files.filter(production)) {
  for (const hit of scan(file, read(file), currentHelpers)) {
    if (hit.id) remember(explicit, hit.id, hit.at)
    else if (hit.candidate) remember(lookupCandidates, hit.candidate, hit.at)
    else remember(dynamic, hit.prefix, hit.at)
  }
}
function usage(id) {
  const refs = explicit.get(id)
  if (refs?.length) return { status: 'referenced', at: refs }
  if (id.startsWith('site/serverMessages:')) return { status: 'API contract' } // checked against Go by check-server-message-i18n
  const patterns = [...dynamic.keys()].filter(prefix => id.startsWith(prefix))
  if (lookupCandidates.has(id)) return { status: 'dynamic candidate', at: lookupCandidates.get(id), patterns }
  if (patterns.length) return { status: 'dynamic candidate', patterns }
  return { status: 'no recognized reference' }
}
const inventory = [...rows].map(([id, translations]) => ({ id, ...usage(id), translations }))
function compareBaseline(baseName) {
  const base = git('rev-parse', '--verify', '--end-of-options', baseName + '^{commit}').trim()
  const baseFiles = new Set(git('ls-tree', '-r', '--name-only', '-z', base).split('\0'))
  function keysAt(rev, file) {
    const source = git('show', rev + ':' + file)
    return file.endsWith('.json') ? new Set(flatten(JSON.parse(source)).keys()) : new Set(literals(source, file).map(x => x.key))
  }
  const baseStrings = new Map()
  function findBaseLiteral(value, effective) {
    for (const { at = [] } of effective) {
      for (const location of at) {
        const file = location.slice(0, location.lastIndexOf(':'))
        if (!baseFiles.has(file) || !/\.tsx?$/.test(file)) continue
        if (!baseStrings.has(file)) baseStrings.set(file, inlineStrings(git('show', base + ':' + file), file))
        const line = baseStrings.get(file).get(value)
        if (line) return file + ':' + line
      }
    }
  }
  const additions = []
  for (const file of files.filter(path => path.startsWith(client + '/messages/zh-') || path.startsWith(admin + '/messages/zh-') || path === server + '/zh.json')) {
    const entries = file.endsWith('.json')
      ? [...flatten(JSON.parse(read(file)))].map(([key, value]) => ({ key, value }))
      : literals(read(file), file)
    const before = baseFiles.has(file) ? keysAt(base, file) : new Set()
    const added = entries.filter(({ key }) => !before.has(key))
    if (!added.length) continue
    const historic = new Map()
    if (options.history) {
      const commits = git('log', '--first-parent', '--diff-filter=AM', '--format=%H', base, '--', file).trim().split('\n').filter(Boolean)
      for (const rev of commits) {
        const old = keysAt(rev, file)
        for (const { key } of added) if (!historic.has(key) && old.has(key)) historic.set(key, rev)
        if (historic.size === added.length) break
      }
    }
    for (const { key, value } of added) {
      const effective = (origins.get(file + '#' + key) ?? []).map(id => ({ id, ...usage(id) }))
      additions.push({ file, key, effective, sameLiteralAtBase: findBaseLiteral(value, effective), ...(historic.has(key) ? { previouslyPresent: historic.get(key) } : {}) })
    }
  }
  // A shrinking inventory alone cannot identify regressions. Show surviving
  // keys whose previously observed call site disappeared in this diff.
  const changed = git('diff', '--no-renames', '--name-only', '-z', base, '--').split('\0')
    .filter(file => baseFiles.has(file) && production(file))
  const lost = new Map()
  if (changed.length) {
    const readBase = file => git('show', base + ':' + file)
    const helpers = adminHelpers(baseFiles, readBase)
    const unresolved = new Set(inventory.filter(row => row.status === 'no recognized reference').map(row => row.id))
    const previousDefinitions = new Map()
    function existedAtBase(id) {
      const origin = rows.get(id)?.zh
      if (!origin || !baseFiles.has(origin.file)) return false
      if (!previousDefinitions.has(origin.file)) previousDefinitions.set(origin.file, keysAt(base, origin.file))
      return previousDefinitions.get(origin.file).has(origin.key)
    }
    for (const file of changed) {
      for (const hit of scan(file, readBase(file), helpers)) {
        const candidates = hit.id ? [hit.id] : hit.candidate ? [hit.candidate] : hit.prefix && !hit.prefix.endsWith(':')
          ? [...unresolved].filter(id => id.startsWith(hit.prefix)) : []
        for (const id of candidates) {
          if (unresolved.has(id) && existedAtBase(id))
            remember(lost, id, { atBase: hit.at, match: hit.id ? 'literal' : hit.candidate ? 'lookup' : 'prefix' })
        }
      }
    }
  }
  return { base, additions, lostReferenceCandidates: Object.fromEntries(lost) }
}
const counts = {}
for (const { status } of inventory) counts[status] = (counts[status] ?? 0) + 1
const unknown = [...explicit.keys()].filter(id => !rows.has(id))
for (const id of unknown) errors.push(`Unknown translation: ${id} at ${explicit.get(id)[0]}`)
const result = { locales, keys: rows.size, counts, inheritedEnglish: fallback, errors }
const baseName = options.base ?? (options.history ? 'HEAD' : undefined)
if (baseName !== undefined) Object.assign(result, compareBaseline(baseName))
if (options.groups) {
  function* sources(pattern) {
    for (const file of files) if (production(file) && pattern.test(file)) yield [file, read(file)]
  }
  result.groups = groupUnresolved(inventory, sources(/\.go$/), sources(frontend))
}
if (options.json) console.log(JSON.stringify(options.all ? { ...result, inventory, dynamic: Object.fromEntries(dynamic) } : result, null, 2))
else {
  console.log('Effective keys:', rows.size, counts, '\nInherited English leaves:', fallback)
  if (options.groups) {
    for (const [reason, namespaces] of Object.entries(result.groups)) {
      console.log(reason, Object.values(namespaces).flat().length)
      for (const [ns, rows] of Object.entries(namespaces)) {
        console.log(' ', ns, rows.length, options.all ? rows.map(row => row.id).join(', ') : rows[0].id)
      }
    }
  }
  if (result.additions) {
    const { additions } = result
    console.log(`Added Chinese literal definitions vs ${baseName}: ${additions.length}`)
    console.log('Same text in baseline caller:', additions.filter(entry => entry.sameLiteralAtBase).length)
    if (options.history) console.log('Previously present at same dictionary path/key:', additions.filter(entry => entry.previouslyPresent).length)
    for (const entry of additions) console.log(`  ${entry.file}#${entry.key}: ${entry.effective.map(r => r.status + (r.at ? ' ' + r.at[0] : '')).join(', ') || 'not in effective runtime catalog'}${entry.previouslyPresent ? ' (previously present at ' + entry.previouslyPresent + ')' : ''}`)
    for (const [id, evidence] of Object.entries(result.lostReferenceCandidates)) console.log('Review lost reference:', id, evidence[0].match, baseName + ':' + evidence[0].atBase)
  }
  if (options.all && !options.groups) for (const row of inventory) console.log(row.status, row.id, row.at?.[0] ?? '')
  for (const error of errors) console.error(error)
}
if (errors.length) process.exitCode = 1
