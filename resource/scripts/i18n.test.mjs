import assert from 'node:assert/strict'
import { test } from 'node:test'
import { messageId, parseJson, requiresIcu, revision, translationStatus, validateEntry, validateNamespace } from './i18n-catalog.mjs'
import { formatMessage, prepareIcuFormat } from '../packages/i18n/src/format.ts'
import { parseBackendMessageCodes, validateServerMessages } from './check-server-message-i18n.mjs'

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
