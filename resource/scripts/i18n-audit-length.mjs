import { readFile } from 'node:fs/promises'
import { join } from 'node:path'
import { parse, TYPE } from '@formatjs/icu-messageformat-parser'
import ts from 'typescript'
import { formatMessage, prepareIcuFormat } from '../packages/i18n/src/format.ts'
import { root, sourceLocale, translationStatus } from './i18n-catalog.mjs'

const segmenter = new Intl.Segmenter(undefined, { granularity: 'grapheme' })
const wide = /[\p{Script=Han}\p{Script=Hangul}\p{Script=Hiragana}\p{Script=Katakana}\p{Extended_Pictographic}\p{Regional_Indicator}\u3000-\u303f\uff01-\uff60]/u
const halfwidth = /^[\uff61-\uffdc]/u
const numberSamples = [0, 1, 2, 3, 5, 11, 21, 100, 1.5, 1000000]
const sampleDate = Date.UTC(2026, 0, 15, 12, 30)
const relatedNavigationKeys = {
  'admin.shell:registrationInvitations': [
    'admin.assets:newInvitation', 'admin.assets:createInvitation', 'admin.assets:revokeInvitation',
    'admin.assets:revokeInvitationConfirm', 'site.auth:invitationCode',
  ],
  'admin.shell:resources': ['admin.assets:files', 'admin.assets:filesHint'],
  'shared.shell:nav.drafts': ['site.drafts:title', 'site.drafts:emptyTitle', 'site.drafts:emptyHint'],
}

export function estimatedWidth(text) {
  return Math.max(...text.split(/\r?\n/).map(line => {
    let width = 0
    for (const { segment } of segmenter.segment(line)) {
      if (wide.test(segment) && !halfwidth.test(segment)) width += 2
      else if (/^\s+$/u.test(segment)) width += 0.5
      else if (!/^\p{Cf}+$/u.test(segment)) width++
    }
    return width
  }))
}

function samplesFor(messages) {
  const defaults = Object.create(null)
  const alternatives = new Map()
  function visit(elements) {
    for (const element of elements) {
      if (!('value' in element) || element.type === TYPE.literal) continue
      const { value: name, type } = element
      if (type === TYPE.date || type === TYPE.time) defaults[name] = sampleDate
      else if (type === TYPE.select || type === TYPE.plural) {
        const choices = alternatives.get(name) ?? new Set()
        const keys = Object.keys(element.options)
        const values = type === TYPE.select ? keys : [
          ...numberSamples.map(number => number + (element.offset ?? 0)),
          ...keys.filter(key => key.startsWith('=')).map(key => Number(key.slice(1))),
        ]
        for (const value of values) choices.add(value)
        alternatives.set(name, choices)
        defaults[name] = type === TYPE.plural ? 2 : 'other'
        for (const branch of Object.values(element.options)) visit(branch.value)
      } else if (type === TYPE.number) defaults[name] = 2
      else if (!(name in defaults)) defaults[name] = 'Sample'
    }
  }
  for (const text of messages) visit(parse(text))
  const selectors = [...alternatives].map(([name, choices]) => [name, [...new Set([defaults[name], ...choices])]])
  const combinations = selectors.reduce((total, [, choices]) => total * choices.length, 1)
  if (combinations <= 256) {
    return selectors.reduce((samples, [name, choices]) =>
      samples.flatMap(sample => choices.map(value => ({ ...sample, [name]: value }))), [defaults])
  }
  // Sample each selector independently when the full set of combinations is too large.
  return [defaults, ...selectors.flatMap(([name, choices]) =>
    choices.filter(value => value !== defaults[name]).map(value => ({ ...defaults, [name]: value })))]
}

async function adminNavigation() {
  const path = join(root, 'apps/web/src/admin/nav.tsx')
  const file = ts.createSourceFile(path, await readFile(path, 'utf8'), ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
  const labels = new Set()
  const groups = file.statements.filter(ts.isVariableStatement).flatMap(statement => [...statement.declarationList.declarations])
    .find(declaration => declaration.name.getText(file) === 'adminNavGroups')
  if (!groups?.initializer) throw Error('Cannot locate adminNavGroups for the length audit')
  function visit(node) {
    if (ts.isPropertyAssignment(node) && node.name.getText(file) === 'label' && ts.isStringLiteral(node.initializer))
      labels.add(node.initializer.text)
    ts.forEachChild(node, visit)
  }
  visit(groups.initializer)
  return labels
}

function navigationGroup(namespace, key, adminLabels) {
  if (namespace === 'admin.shell' && adminLabels.has(key)) return 'admin navigation'
  if (namespace === 'shared.shell' && (key.startsWith('nav.') || ['accessGroups', 'themePreview', 'resources'].includes(key)))
    return 'forum navigation'
}

function median(values) {
  const sorted = [...values].sort((a, b) => a - b)
  const middle = Math.floor(sorted.length / 2)
  return sorted.length % 2 ? sorted[middle] : (sorted[middle - 1] + sorted[middle]) / 2
}

function addValue(groups, key, value) {
  const values = groups.get(key) ?? []
  values.push(value)
  groups.set(key, values)
}

function relatedMessages(catalog, row) {
  if (!row.group) return []
  const key = row.key.replace(/^nav\./, '')
  const ids = new Set(relatedNavigationKeys[row.id] ?? [])
  for (const [namespace, entries] of catalog) {
    if (!namespace.startsWith(row.group === 'admin navigation' ? 'admin.' : 'site.')) continue
    for (const candidate of [key, key + 'Hint']) {
      if (entries[candidate]) ids.add(namespace + ':' + candidate)
    }
  }
  ids.delete(row.id)
  const related = []
  for (const id of ids) {
    const [namespace, candidate] = id.split(':')
    const entry = catalog.get(namespace)?.[candidate]
    if (entry && translationStatus(entry, row.locale) === 'ready')
      related.push({ id, source: entry.source, text: entry.translations[row.locale].text })
  }
  return related
}

export async function auditLength(catalog, { scope, locale, limit }) {
  const adminLabels = await adminNavigation()
  await prepareIcuFormat()
  const rows = []
  const baseline = new Map()
  const navigationWidths = new Map()

  for (const [namespace, entries] of catalog) {
    for (const [key, entry] of Object.entries(entries)) {
      const translations = Object.entries(entry.translations).filter(([lang]) => translationStatus(entry, lang) === 'ready')
      const samples = samplesFor([entry.source, ...translations.map(([, translation]) => translation.text)])
      const width = (text, lang) => Math.max(...samples.map(params => estimatedWidth(formatMessage(text, lang, params))))
      const sourceWidth = width(entry.source, sourceLocale)
      const bucket = sourceWidth <= 12 ? 'short' : sourceWidth <= 32 ? 'medium' : 'long'
      const group = navigationGroup(namespace, key, adminLabels)
      for (const [lang, translation] of translations) {
        const estimated = width(translation.text, lang)
        const factor = Math.log((estimated + 2) / (sourceWidth + 2))
        const bucketKey = lang + '/' + bucket
        const groupKey = group && group + '/' + lang
        addValue(baseline, bucketKey, factor)
        if (groupKey) addValue(navigationWidths, groupKey, estimated)
        rows.push({ id: namespace + ':' + key, key, locale: lang, group, bucketKey, groupKey,
          source: entry.source, text: translation.text, sourceWidth, estimated, factor })
      }
    }
  }

  const normalExpansion = new Map([...baseline].map(([key, values]) => [key, median(values)]))
  const normalNavWidth = new Map([...navigationWidths].map(([key, values]) => [key, median(values)]))
  const byKey = new Map()
  for (const row of rows) addValue(byKey, row.id, row)
  const normalize = row => row.factor - normalExpansion.get(row.bucketKey)
  const round = value => Math.round(value * 100) / 100
  const findings = []
  let checked = 0
  for (const row of rows) {
    if ((scope === 'navigation' && !row.group) || (locale && row.locale !== locale)) continue
    checked++
    const peers = byKey.get(row.id).filter(peer => peer.locale !== row.locale).map(normalize)
    const typicalWidth = (row.sourceWidth + 2) * Math.exp(normalExpansion.get(row.bucketKey) + (peers.length ? median(peers) : 0)) - 2
    const ratio = (row.estimated + 2) / (typicalWidth + 2)
    const navWidth = normalNavWidth.get(row.groupKey)
    const signals = []
    if (row.estimated - typicalWidth >= 5 && ratio >= 1.55) signals.push('longer than comparable translations')
    if (typicalWidth - row.estimated >= 5 && ratio <= 0.65) signals.push('shorter than comparable translations')
    const longNavigation = navWidth !== undefined && row.estimated - navWidth >= 7 && row.estimated >= navWidth * 1.55
    if (longNavigation) signals.push('long compared with other navigation labels')
    if (!signals.length) continue
    const score = Math.max(Math.abs(Math.log2(ratio)), longNavigation ? Math.log2(row.estimated / navWidth) : 0)
    findings.push({ id: row.id, key: row.key, locale: row.locale, group: row.group, source: row.source, text: row.text,
      sourceWidth: row.sourceWidth, width: row.estimated, expectedWidth: round(typicalWidth), ratio: round(ratio),
      ...(navWidth !== undefined && { navigationMedian: navWidth }), signals, score: round(score) })
  }
  findings.sort((a, b) => b.score - a.score || a.id.localeCompare(b.id) || a.locale.localeCompare(b.locale))
  return { checked, flagged: findings.length, findings: findings.slice(0, limit).map(({ key, ...row }) =>
    ({ ...row, related: relatedMessages(catalog, { ...row, key }) })) }
}
