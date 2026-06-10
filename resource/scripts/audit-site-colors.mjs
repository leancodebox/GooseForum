#!/usr/bin/env node
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative, resolve, sep } from 'node:path'

const root = resolve(import.meta.dirname, '..')

const include = [
  'src/site',
  'src/runtime',
  'src/types',
  'src/styles/resource.css',
  'templates/layout/app.gohtml',
  'templates/pages',
  'templates/partials',
  'templates/view',
]

const exclude = [
  'src/admin',
  'templates/layout/admin.gohtml',
  'templates/pages/admin.gohtml',
]

const extensions = new Set(['.css', '.gohtml', '.ts', '.vue'])
const colorNames = [
  'slate',
  'gray',
  'zinc',
  'neutral',
  'stone',
  'red',
  'orange',
  'amber',
  'yellow',
  'lime',
  'green',
  'emerald',
  'teal',
  'cyan',
  'sky',
  'blue',
  'indigo',
  'violet',
  'purple',
  'fuchsia',
  'pink',
  'rose',
  'white',
  'black',
  'transparent',
  'current',
].join('|')

const utilityPrefixes = [
  'bg',
  'text',
  'border',
  'ring',
  'divide',
  'outline',
  'decoration',
  'accent',
  'caret',
  'from',
  'via',
  'to',
  'placeholder',
  'shadow',
].join('|')

const colorUtilityPattern = new RegExp(
  String.raw`(?:[A-Za-z0-9_\-\[\]=&*()'"./#%]+:)*(?:${utilityPrefixes})-(?:(?:${colorNames})(?:-\d{2,3})?(?:\/\d{1,3})?|\[[^\]\s"'` + '`' + String.raw`]+\](?:\/\d{1,3})?)`,
  'g',
)
const hexPattern = /#[0-9a-fA-F]{3,8}\b/g
const functionColorPattern = /\b(?:rgb|rgba|hsl|hsla|oklch|oklab)\([^)]*\)/g
const cssVarColorPattern = /var\(--[A-Za-z0-9_-]*(?:color|bg|border|text|ring|blue|gray|red|green|amber|rose|slate)[A-Za-z0-9_-]*\)/g
const intentionalUtilityAllowlist = new Set([
  'src/site/components/UserHoverCard.vue',
  'src/site/pages/UserPage.vue',
])
const intentionalHexAllowlist = new Set([
  '#030712',
  '#1b1b1b',
  '#059669',
  '#244bd6',
  '#315ef4',
  '#34A853',
  '#374151',
  '#4285F4',
  '#6b7280',
  '#dc2626',
  '#e5e7eb',
  '#EA4335',
  '#FBBC05',
  '#f9fafb',
  '#ffffff',
])
const intentionalFunctionAllowlist = [
  /^oklch\(/,
  /^rgb\(244 63 94/,
  /^rgba\(15, 23, 42, 0\.08\)$/,
  /^rgba\(37, 99, 235, 0\.(35|9)\)$/,
  /^rgba\(49, 94, 244, 0\.2\)$/,
]

const files = include.flatMap((item) => collectFiles(resolve(root, item)))
const results = {
  utilities: new Map(),
  baseUtilities: new Map(),
  hex: new Map(),
  colorFunctions: new Map(),
  cssVars: new Map(),
}

for (const file of files) {
  const text = readFileSync(file, 'utf8')
  for (const match of text.matchAll(colorUtilityPattern)) {
    const value = match[0]
    const rel = toPosix(relative(root, file))
    if (intentionalUtilityAllowlist.has(rel)) continue
    if (!isColorUtility(value)) continue

    add(results.utilities, value, file)
    add(results.baseUtilities, stripVariants(value), file)
  }
  collectMatches(results.hex, text.matchAll(hexPattern), file, isIntentionalHex)
  collectMatches(results.colorFunctions, text.matchAll(functionColorPattern), file, isIntentionalFunctionColor)
  collectMatches(results.cssVars, text.matchAll(cssVarColorPattern), file)
}

console.log(`Scanned ${files.length} non-admin resource files.\n`)
printSection('Base Tailwind color utilities', results.baseUtilities, 40)
printSuggestions(results.baseUtilities, 50)
printSection('Full Tailwind color utilities, including variants', results.utilities, 40)
printSection('Hard-coded hex colors', results.hex, 40)
printSection('CSS color functions', results.colorFunctions, 40)
printSection('Existing color-like CSS variables', results.cssVars, 40)

function collectFiles(path) {
  if (!exists(path) || isExcluded(path)) return []
  const stat = statSync(path)
  if (stat.isFile()) return shouldRead(path) ? [path] : []
  if (!stat.isDirectory()) return []

  return readdirSync(path).flatMap((name) => collectFiles(join(path, name)))
}

function exists(path) {
  try {
    statSync(path)
    return true
  } catch {
    return false
  }
}

function shouldRead(path) {
  return extensions.has(path.slice(path.lastIndexOf('.')))
}

function isExcluded(path) {
  const rel = toPosix(relative(root, path))
  return exclude.some((item) => rel === item || rel.startsWith(`${item}/`))
}

function collectMatches(target, matches, file, skip = () => false) {
  for (const match of matches) {
    if (skip(match[0])) continue
    add(target, match[0], file)
  }
}

function isIntentionalHex(value) {
  return intentionalHexAllowlist.has(value)
}

function isIntentionalFunctionColor(value) {
  return intentionalFunctionAllowlist.some((pattern) => pattern.test(value))
}

function isColorUtility(value) {
  const base = stripVariants(value)
  const parsed = parseColor(splitUtility(base)[1] ?? '')
  if (parsed?.color === 'neutral' && !parsed.shade) return false

  if (!base.includes('[')) return true

  return /\[(?:#|(?:rgb|rgba|hsl|hsla|oklch|oklab)\(|var\(--)/.test(base)
}

function add(target, value, file) {
  const rel = toPosix(relative(root, file))
  const entry = target.get(value) ?? { count: 0, files: new Set() }
  entry.count += 1
  entry.files.add(rel)
  target.set(value, entry)
}

function stripVariants(value) {
  let bracketDepth = 0
  let lastColon = -1

  for (let index = 0; index < value.length; index += 1) {
    const char = value[index]
    if (char === '[') bracketDepth += 1
    else if (char === ']') bracketDepth = Math.max(0, bracketDepth - 1)
    else if (char === ':' && bracketDepth === 0) lastColon = index
  }

  return lastColon === -1 ? value : value.slice(lastColon + 1)
}

function printSection(title, map, limit) {
  const rows = [...map.entries()]
    .sort((a, b) => b[1].count - a[1].count || a[0].localeCompare(b[0]))
    .slice(0, limit)

  console.log(`${title}`)
  console.log('-'.repeat(title.length))
  if (!rows.length) {
    console.log('(none)\n')
    return
  }

  const valueWidth = Math.max(...rows.map(([value]) => value.length), 8)
  for (const [value, entry] of rows) {
    const files = [...entry.files].slice(0, 4).join(', ')
    const suffix = entry.files.size > 4 ? `, +${entry.files.size - 4} more` : ''
    console.log(`${value.padEnd(valueWidth)}  ${String(entry.count).padStart(4)}  ${String(entry.files.size).padStart(3)} files  ${files}${suffix}`)
  }
  console.log('')
}

function printSuggestions(map, limit) {
  const rows = [...map.entries()]
    .map(([value, entry]) => [value, entry, suggestReplacement(value)])
    .filter(([, , suggestion]) => Boolean(suggestion))
    .sort((a, b) => b[1].count - a[1].count || a[0].localeCompare(b[0]))
    .slice(0, limit)

  const title = 'Suggested daisy-style replacements backed by --gf-color-*'
  console.log(title)
  console.log('-'.repeat(title.length))
  if (!rows.length) {
    console.log('(none)\n')
    return
  }

  const valueWidth = Math.max(...rows.map(([value]) => value.length), 8)
  const suggestionWidth = Math.max(...rows.map(([, , suggestion]) => suggestion.length), 10)
  for (const [value, entry, suggestion] of rows) {
    console.log(`${value.padEnd(valueWidth)}  ->  ${suggestion.padEnd(suggestionWidth)}  ${String(entry.count).padStart(4)} uses`)
  }
  console.log('')
}

function suggestReplacement(value) {
  const [prefix, colorWithShade] = splitUtility(value)
  if (!prefix || !colorWithShade) return undefined

  const parsed = parseColor(colorWithShade)
  if (!parsed) return undefined

  const { color, shade, opacity } = parsed
  const alpha = opacity ? `/${opacity}` : ''

  if (color === 'white') {
    if (prefix === 'bg') return `bg-base-100${alpha}`
    if (prefix === 'text') return `text-primary-content${alpha}`
    if (prefix === 'border' || prefix === 'ring' || prefix === 'divide') return `${prefix}-line${alpha}`
  }

  if (color === 'black') {
    if (prefix === 'bg') return `bg-neutral${alpha}`
    if (prefix === 'text') return `text-neutral-content${alpha}`
    if (prefix === 'border' || prefix === 'ring' || prefix === 'divide') return `${prefix}-neutral${alpha}`
  }

  if (['slate', 'gray', 'zinc', 'neutral', 'stone'].includes(color)) {
    return suggestNeutral(prefix, shade, alpha)
  }

  if (color === 'blue' || color === 'sky' || color === 'indigo' || color === 'violet') {
    return suggestSemantic(prefix, shade, alpha, 'primary', 'info')
  }

  if (color === 'red' || color === 'rose' || color === 'pink') {
    return suggestSemantic(prefix, shade, alpha, 'error', 'error')
  }

  if (color === 'amber' || color === 'yellow' || color === 'orange') {
    return suggestSemantic(prefix, shade, alpha, 'warning', 'warning')
  }

  if (color === 'green' || color === 'emerald' || color === 'teal' || color === 'cyan' || color === 'lime') {
    return suggestSemantic(prefix, shade, alpha, 'success', 'accent')
  }

  return undefined
}

function splitUtility(value) {
  const bracketIndex = value.indexOf('-[')
  if (bracketIndex !== -1) return [value.slice(0, bracketIndex), value.slice(bracketIndex + 1)]

  const index = value.indexOf('-')
  if (index === -1) return []
  return [value.slice(0, index), value.slice(index + 1)]
}

function parseColor(value) {
  if (value.startsWith('[')) return undefined

  const opacityParts = value.split('/')
  const colorParts = opacityParts[0].split('-')
  const shadeCandidate = colorParts.at(-1)
  const hasShade = /^\d{2,3}$/.test(shadeCandidate ?? '')

  return {
    color: hasShade ? colorParts.slice(0, -1).join('-') : colorParts.join('-'),
    shade: hasShade ? Number(shadeCandidate) : undefined,
    opacity: opacityParts[1],
  }
}

function suggestNeutral(prefix, shade, alpha) {
  if (prefix === 'bg') {
    if (!shade || shade <= 50) return `bg-base-200${alpha}`
    if (shade <= 100) return `bg-base-300${alpha}`
    if (shade >= 800) return `bg-neutral${alpha}`
    return `bg-base-200${alpha}`
  }

  if (prefix === 'text' || prefix === 'placeholder' || prefix === 'caret') {
    if (!shade || shade >= 800) return `${prefix}-base-content${alpha}`
    if (shade >= 600) return `${prefix}-base-content/75`
    return `${prefix}-base-content/55`
  }

  if (prefix === 'border' || prefix === 'ring' || prefix === 'divide' || prefix === 'outline') {
    return `${prefix}-line${alpha}`
  }

  if (prefix === 'shadow') return undefined
  return undefined
}

function suggestSemantic(prefix, shade, alpha, strongToken, softToken) {
  const token = shade && shade <= 100 ? softToken : strongToken

  if (prefix === 'bg') {
    if (shade && shade <= 100) return `bg-${token}/10`
    return `bg-${token}${alpha}`
  }

  if (prefix === 'text' || prefix === 'placeholder' || prefix === 'caret') {
    return `${prefix}-${strongToken}${alpha}`
  }

  if (prefix === 'border' || prefix === 'ring' || prefix === 'divide' || prefix === 'outline') {
    if (shade && shade <= 200) return `${prefix}-${strongToken}/20`
    return `${prefix}-${strongToken}${alpha}`
  }

  if (prefix === 'from' || prefix === 'via' || prefix === 'to' || prefix === 'accent' || prefix === 'decoration') {
    return `${prefix}-${strongToken}${alpha}`
  }

  return undefined
}

function toPosix(path) {
  return path.split(sep).join('/')
}
