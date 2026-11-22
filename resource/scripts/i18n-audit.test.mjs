import assert from 'node:assert/strict'
import test from 'node:test'
import { groupUnresolved } from './i18n-triage.mjs'
import { flatten, literals } from './i18n-catalog.mjs'
import { adminDictionaryReads, references } from './i18n-references.mjs'

test('retains nested key paths and preserves dotted API codes', () => {
  const source = 'export default { first: { label: "一" }, second: { label: "二" }, "api.ok": "好" }'
  assert.deepEqual(literals(source, 'zh.ts').map(x => x.key), ['first.label', 'second.label', 'api.ok'])
  assert.deepEqual([...flatten({ first: { label: 'a' }, 'api.ok': 'b' }).keys()], ['first.label', 'api.ok'])
})

test('resolves scoped translator aliases, explicit namespaces and conditional keys', () => {
  const code = `function Page() {
    const { t: tr } = useTranslation('settings')
    tr(flag ? 'on' : 'off')
    tr('retry', { ns: 'topic' })
    tr('home:next')
    tr(\`authLogActions.\${action}\`)
    function other(tr) { tr('ignored') }
  }
  function NoTranslator() { tr('unused') }`
  const hits = references(code, 'page.tsx')
  assert.deepEqual(hits.map(h => h.id ?? h.prefix), ['site/settings:on', 'site/settings:off', 'site/topic:retry', 'site/home:next', 'site/settings:authLogActions.'])
})

test('records admin translator props in their own section only', () => {
  const code = "import type { UserTextKey } from './users-i18n'; function Panel({ text }: { text: (key: UserTextKey) => string }) { text('trustLevel'); log('trustLevel') }"
  assert.deepEqual(references(code, 'user.tsx', path => path === './users-i18n' ? 'users' : undefined), [{ id: 'admin/users:trustLevel', at: 'user.tsx:1' }])
})

test('reports a restored key, distinguishes inherited English and fails on a broken placeholder', async t => {
  const { mkdtempSync, mkdirSync, writeFileSync, symlinkSync, cpSync, rmSync } = await import('node:fs')
  const { tmpdir } = await import('node:os')
  const { join, dirname } = await import('node:path')
  const { execFileSync, spawnSync } = await import('node:child_process')
  const dir = mkdtempSync(join(tmpdir(), 'goose-i18n-audit-'))
  t.after(() => rmSync(dir, { recursive: true, force: true }))
  const write = (file, contents) => {
    mkdirSync(dirname(join(dir, file)), { recursive: true })
    writeFileSync(join(dir, file), contents)
  }
  const git = (...args) => execFileSync('git', args, { cwd: dir })
  const site = 'resource/packages/client/src/i18n'
  const dictionaries = 'resource/apps/web/src/admin/messages'
  const zh = dictionaries + '/zh-users.ts'
  write(site + '/locale.ts', "export const supportedLocales = ['zh', 'en', 'ja']")
  write('resource/packages/runtime/src/i18n/loader.ts',
    'export const gooseNamespaces = []; export function prepareGooseTranslations() {} export const cachedGooseResources = () => ({zh:{},en:{},ja:{}})')
  write('resource/apps/web/src/admin/translation-resources.ts',
    "export const translationLoaders = Object.fromEntries(['zh','en','ja'].map(l => [l+'/users', () => import('./messages/'+l+'-users')]))")
  for (const locale of ['zh', 'en', 'ja']) write('app/bundles/i18n/locales/' + locale + '.json', '{}')
  write(dictionaries + '/en-users.ts', 'export default { old: "Hello {name}", retired: "Retired" }')
  write(dictionaries + '/ja-users.ts', 'import en from "./en-users"; export default { ...en }')
  const chinese = 'export default { old: "你好，{name}", retired: "过期" }'
  write(zh, chinese)
  write('resource/.gitignore', 'node_modules/')
  cpSync(import.meta.dirname, join(dir, 'resource/scripts'), { recursive: true })
  symlinkSync(join(import.meta.dirname, '../node_modules'), join(dir, 'resource/node_modules'), 'dir')
  git('init', '-q')
  const commit = message => {
    git('add', '.')
    git('-c', 'user.email=audit@example.com', '-c', 'user.name=Audit Test', '-c', 'commit.gpgsign=false', 'commit', '-qm', message)
  }
  const current = spawnSync(process.execPath, ['resource/scripts/check-i18n-usage.mjs', '--groups', '--json'], { cwd: dir, encoding: 'utf8' })
  assert.equal(current.status, 0, current.stderr) // No commits exist yet.
  assert.equal(JSON.parse(current.stdout).groups.unusedCandidate['admin/users'].length, 2)
  commit('initial translation')
  const previous = git('rev-parse', 'HEAD').toString().trim()
  const caller = 'resource/apps/web/src/admin/pages/caller.ts'
  write(caller, "import { getAdminDictionary } from '../translation-loader'; getAdminDictionary('users', 'zh').retired")
  write(zh, 'export default { retired: "过期" }')
  commit('remove translation')
  rmSync(join(dir, caller))
  write(zh, chinese)
  const run = () => spawnSync(process.execPath, ['resource/scripts/check-i18n-usage.mjs', '--base', 'HEAD', '--history', '--groups', '--json'], { cwd: dir, encoding: 'utf8' })
  const clean = run()
  assert.equal(clean.status, 0, clean.stderr)
  const report = JSON.parse(clean.stdout)
  assert.equal(report.additions[0].previouslyPresent, previous)
  assert.equal(report.additions[0].effective[0].status, 'no recognized reference')
  assert.equal(report.inheritedEnglish.ja, 2)
  assert.ok(report.groups.unusedCandidate['admin/users'].some(row => row.id === 'admin/users:retired'))
  assert.deepEqual(report.lostReferenceCandidates, { 'admin/users:retired': [{ atBase: caller + ':1', match: 'literal' }] })
  write(zh, 'export default { old: "你好，{user}", retired: "过期" }')
  const broken = run()
  assert.equal(broken.status, 1)
  assert.match(JSON.parse(broken.stdout).errors.join('\n'), /placeholders/)
})

test('recognizes local JSX translator props but leaves mixed namespaces unresolved', () => {
  const code = `function Child({ translate: t }) { t('childLabel') }
  function Mixed({ t }) { t('ambiguous') }
  function Parent() {
    const { t } = useTranslation('topic')
    const { t: other } = useTranslation('home')
    return <><Child translate={t}/><Mixed t={t}/><Mixed t={other}/></>
  }`
  assert.deepEqual(references(code, 'page.tsx').map(hit => hit.id), ['site/topic:childLabel'])
})

test('admin dictionary reads follow const bindings; computed keys stay candidates', () => {
  const file = 'audit.ts'
  const code = `import { getAdminDictionary as get } from './translation-loader'
    const dict = get('audit', locale)
    const labels = dict.authLogs
    labels['title'].replace('a','b')
    dict.actions[row.action]
    function unrelated(labels) { return labels.retry }
    type OnlyType = typeof labels.retry
    let mutable = labels; mutable.retry
    function fake(get) { return get('audit', locale).authLogs.retry }
  `
  const keys = new Set(['admin/audit:authLogs.title', 'admin/audit:authLogs.retry', 'admin/audit:actions.login'])
  assert.deepEqual(adminDictionaryReads(code, file, keys).map(h => h.id ?? h.prefix),
    ['admin/audit:authLogs.title', 'admin/audit:actions.'])
})

test('mixed type imports and closed lookup tables preserve exact candidate keys', () => {
  const code = `import { format, type AuditTextKey as Key } from './audit-i18n'
    const actions = { 1: 'edit', 2: 'remove' }
    const changed = { 1: 'previous' }
    changed[1] = external
    function Page({text}: {text: (k: Key) => string}) {
      text(actions[id] || 'unknown')
      text(changed[id] || 'fallback')
    }`
  const hits = references(code, 'page.tsx', () => 'audit')
  assert.deepEqual(hits.map(hit => hit.id ?? hit.candidate ?? hit.prefix),
    ['admin/audit:edit', 'admin/audit:remove', 'admin/audit:unknown', 'admin/audit:', 'admin/audit:fallback'])
  assert.deepEqual(hits.filter(h => h.candidate).map(h => h.candidate), ['admin/audit:edit', 'admin/audit:remove'])
})

test('only current source literals and shared definitions inform unresolved groups', () => {
  const row = (id, key) => ({ id, status: 'no recognized reference',
    translations: { zh: { file: 'zh.ts', key, value: '相同' }, en: { value: 'same' } } })
  const rows = [row('site/demo:shared', 'shared'), row('site/demo:same', 'separate'), row('go:label', 'label'),
    { ...row('site/other:same', 'shared'), status: 'referenced', at: ['page.ts:1'] }]
  const groups = groupUnresolved(rows, [['app/page.go', 'return "label"']])
  assert.deepEqual(groups.sharedDefinition['site/demo'][0].evidence.sharedDefinition[0].alsoUsedBy, ['site/other:same'])
  assert.deepEqual(groups.goSourceLiteral.go[0].evidence.goSourceLiteral, ['app/page.go:1'])
  assert.equal(groups.unusedCandidate['site/demo'][0].id, 'site/demo:same')
})

test('forwards translator props through multiple local components without resolving mixed callers', () => {
  const source = `
    function Leaf({ t }) { t('found') }
    function Mixed({ t }) { t('ambiguous') }
    function Middle({ tr }) { return <><Leaf t={tr}/><Mixed t={tr}/></> }
    function Page() {
      const {t} = useTranslation('topic')
      const {t: other} = useTranslation('home')
      return <><Middle tr={t}/><Mixed t={other}/></>
    }`
  assert.deepEqual(references(source, 'page.tsx').map(hit => hit.id), ['site/topic:found'])
})

test('secondary spelling evidence protects unresolved translator calls and dynamic prefixes', () => {
  const keys = ['retry', 'tabs.account', 'status', 'absent']
  const rows = keys.map(key => ({ id: 'site/page:' + key, status: 'no recognized reference',
    translations: { zh: { file: 'zh.ts', key, value: key } } }))
  const groups = groupUnresolved(rows, [], [['page.tsx',
    "import retry from 'absent'; type T = 'absent'; t('retry'); labels.status; tr(`tabs.${id}`); /* absent */"]])
  assert.deepEqual(groups.frontendSpelling['site/page'].map(row => row.id), rows.slice(0, 3).map(row => row.id))
  assert.deepEqual(groups.unusedCandidate['site/page'].map(row => row.id), ['site/page:absent'])
})

test('omitted translator props and shadowed bindings do not establish a namespace', () => {
  const code = `function Child({ t = fallback }) { t('ambiguous') }
    function Page() {
      const { t } = useTranslation('topic')
      { const { t } = unrelated; t('shadowed') }
      return <><Child t={t}/><Child/></>
    }`
  assert.deepEqual(references(code, 'page.tsx'), [])
})
