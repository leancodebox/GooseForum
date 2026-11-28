import assert from 'node:assert/strict'
import { test } from 'node:test'
import { messageId, parseJson, requiresIcu, revision, translationStatus, validateEntry, validateNamespace } from './i18n-catalog.mjs'
import { formatMessage, prepareIcuFormat } from '../packages/i18n/src/format.ts'
import { parseBackendMessageCodes, validateServerMessages } from './check-server-message-i18n.mjs'
import { auditLength, estimatedWidth } from './i18n-audit-length.mjs'

test('tracks missing, draft, approved and stale translations against the source and context', () => {
  const entry = { source: '{count, number} 条回复', context: 'Topic counter', translations: {} }
  assert.equal(translationStatus(entry, 'en'), 'missing')
  entry.translations.en = { text: '{count, plural, one {# reply} other {# replies}}', sourceRevision: revision(entry), reviewed: false }
  validateEntry('site.topic:replyCount', entry)
  assert.equal(translationStatus(entry, 'en'), 'draft')
  entry.translations.en.reviewed = true
  assert.equal(translationStatus(entry, 'en'), 'ready')
  entry.context = 'Replies since the last visit'
  assert.equal(translationStatus(entry, 'en'), 'stale')
  entry.source = '作者 {name}'
  validateEntry('site.topic:replyCount', entry)
})

test('rejects malformed messages, mismatched arguments and plural categories for the target language', () => {
  assert.throws(() => validateEntry('site.topic:replyCount', { source: '{count', translations: {} }), /replyCount/)
  const entry = { source: '{count, number} 条回复', translations: {} }
  entry.translations.en = { text: '{num} replies', sourceRevision: revision(entry), reviewed: false }
  assert.throws(() => validateEntry('site.topic:replyCount', entry), /arguments/)
  entry.translations.en.text = '{count, plural, few {# replies} other {# replies}}'
  assert.throws(() => validateEntry('site.topic:replyCount', entry), /invalid en plural category few/)
})

test('catches duplicate and ambiguous keys while preserving dotted API message codes', () => {
  assert.throws(() => parseJson('{"save":"a","save":"b"}', 'catalog'), /duplicate property save/)
  assert.throws(() => messageId('shared.shell:header.__proto__'), /Expected an ID/)
  assert.throws(() => messageId('shared.shell:header.toString.tooltip'), /Expected an ID/)
  const entry = { source: '限制', translations: {} }
  const keys = { 'upload.limit': entry, 'upload.limit.avatar': entry }
  assert.throws(() => validateNamespace('site.auth', keys), /conflicts with a parent key/)
  validateNamespace('shared.serverMessages', keys)
  const codes = parseBackendMessageCodes('const errorCode MessageCode = "upload.limit"')
  assert.deepEqual(validateServerMessages(codes, Object.keys(keys)), ['Unknown Go MessageCode upload.limit.avatar'])
})

test('loads the ICU formatter only for messages that need it and uses the fallback message language', async () => {
  assert.equal(requiresIcu('Hello {name}'), false)
  assert.equal(formatMessage('Hello {name}', 'en', { name: 'Ada' }), 'Hello Ada')
  assert.equal(requiresIcu("A quoted '{name}'"), true)
  const message = '{count, plural, one {# reply} other {# replies}}'
  assert.equal(requiresIcu(message), true)
  await prepareIcuFormat()
  assert.equal(formatMessage(message, 'en', { count: 1 }), '1 reply')
  assert.equal(formatMessage(message, 'en', { count: 2 }), '2 replies')
  assert.equal(formatMessage("A quoted '{name}'", 'en', { name: 'Ada' }), 'A quoted {name}')
})

test('estimates width by grapheme and uses the longest line', () => {
  assert.equal(estimatedWidth('日本語'), 6)
  assert.equal(estimatedWidth('ｶﾞ'), 1)
  assert.equal(estimatedWidth('e\u0301 👩‍💻'), 3.5)
  assert.equal(estimatedWidth('abc\n中文'), 4)
})

test('flags unusually long and short translations and samples ICU plural forms', async () => {
  function entry(english, french = 'Codes créés', source = '邀请码列表') {
    const result = { source, translations: {} }
    const sourceRevision = revision(result)
    result.translations.en = { text: english, sourceRevision, reviewed: true }
    result.translations.fr = { text: french, sourceRevision, reviewed: true }
    return result
  }
  const entries = Object.fromEntries(Array.from({ length: 8 }, (_, i) => ['label' + i, entry('Invite codes')]))
  for (const key of ['dashboard', 'authLogs', 'badges', 'posts']) entries[key] = entry('Invite codes')
  entries.registrationInvitations = entry('Invitation codes to use when signing up for this particular website')
  entries.users = entry('I')
  entries.roles = entry(
    '{mode, select, preview {{n, plural, =101 {Extraordinarily long choice with # items in the list} other {# roles}}} other {{n} roles}}',
    '{mode, select, preview {{n} rôles à vérifier} other {{n} rôles}}',
    '{mode, select, preview {待核对的角色 {n, number} 个} other {{n, number} 个角色}}',
  )
  const catalog = new Map([['admin.shell', entries], ['admin.assets', { registrationInvitations: entry('Invite codes') }]])
  const report = await auditLength(catalog, { scope: 'navigation', locale: 'en', limit: 10 })
  assert.equal(report.checked, 7)
  assert.equal(report.flagged, 3)
  const [long, short, plural] = ['registrationInvitations', 'users', 'roles']
    .map(key => report.findings.find(row => row.id === 'admin.shell:' + key))
  assert.ok(long.signals.includes('long compared with other navigation labels'))
  assert.equal(long.related[0].id, 'admin.assets:registrationInvitations')
  assert.ok(short.signals.includes('shorter than comparable translations'))
  assert.equal(plural.width, estimatedWidth('Extraordinarily long choice with 101 items in the list'))
})
