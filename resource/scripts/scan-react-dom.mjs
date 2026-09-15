#!/usr/bin/env node
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { relative, resolve, join, toNamespacedPath } from 'node:path'
import ts from 'typescript'

const root = resolve(import.meta.dirname, '..')
const requestedRoots = readOptions('--scan-root')
const scanRoots = (requestedRoots.length
  ? requestedRoots
  : ['packages/theme-default/src', 'packages/ui/src', 'packages/runtime/src', 'apps/web/src']
).map(item => resolve(root, item))
const json = process.argv.includes('--json')
const failOnErrors = process.argv.includes('--fail-on-errors')
const maxFindings = Number(readOption('--max-findings') || 80)

const nativeTags = new Set([
  'a',
  'button',
  'dialog',
  'form',
  'iframe',
  'img',
  'input',
  'select',
  'textarea',
])
const focusableTags = new Set([
  'a',
  'button',
  'input',
  'select',
  'textarea',
  'Button',
  'TabsTrigger',
  'ToggleGroupItem',
])
const hardCodedColorPattern = /^(?:bg|text|border|ring|outline|fill|stroke)-(?:slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose|white|black)(?:-|\/|$)/

const inventory = {
  scannedFiles: 0,
  elements: new Map(),
  nativeElements: new Map(),
  components: new Map(),
  classes: new Map(),
  arbitraryClasses: new Map(),
}
const findings = []

for (const file of scanRoots.flatMap(collectFiles)) scanFile(file)

const errors = findings.filter(item => item.severity === 'error')
if (failOnErrors && errors.length) process.exitCode = 1
printReport()

function scanFile(file) {
  const source = readFileSync(file, 'utf8')
  const rel = toPosix(relative(root, file))
  const sourceFile = ts.createSourceFile(
    file,
    source,
    ts.ScriptTarget.Latest,
    true,
    ts.ScriptKind.TSX,
  )
  inventory.scannedFiles += 1
  for (const diagnostic of sourceFile.parseDiagnostics) {
    const line = sourceFile.getLineAndCharacterOfPosition(diagnostic.start || 0).line + 1
    addFinding('error', 'parse-error', `${rel}:${line}`, 'TSX', ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n'))
  }

  visitNested(sourceFile, [])

  function visitNested(node, ancestors) {
    if (ts.isJsxElement(node)) {
      const element = analyzeElement(node.openingElement, node.children, ancestors)
      visitNested(node.openingElement.attributes, ancestors)
      for (const child of node.children) visitNested(child, [...ancestors, element])
      return
    }
    if (ts.isJsxSelfClosingElement(node)) {
      analyzeElement(node, [], ancestors)
      visitNested(node.attributes, ancestors)
      return
    }
    ts.forEachChild(node, child => visitNested(child, ancestors))
  }

  function analyzeElement(opening, children, ancestors) {
    const tag = opening.tagName.getText(sourceFile)
    const line = sourceFile.getLineAndCharacterOfPosition(opening.getStart(sourceFile)).line + 1
    const attributes = readAttributes(opening.attributes.properties, sourceFile)
    const classGroups = attributes.className?.strings ?? []
    const classes = classGroups.flatMap(value => splitClasses(value))
    const native = /^[a-z]/.test(tag)
    const location = `${rel}:${line}`
    const element = { tag, native, classes, location }

    increment(inventory.elements, tag, location)
    increment(native ? inventory.nativeElements : inventory.components, tag, location)
    for (const className of classes) {
      increment(inventory.classes, className, location)
      if (className.includes('[')) increment(inventory.arbitraryClasses, className, location)
    }

    for (const value of classGroups) scanClassGroup(value, location, tag)
    scanSemantics({ tag, native, attributes, children, ancestors, classes, location })
    return element
  }
}

function scanClassGroup(value, location, tag) {
  const classes = splitClasses(value)
  const seen = new Set()
  for (const className of classes) {
    if (seen.has(className)) {
      addFinding('error', 'duplicate-class', location, tag, `duplicate “${className}”`)
    }
    seen.add(className)

    const { base } = splitVariant(className)
    if (hardCodedColorPattern.test(base) && !isIntentionalColor(location, className)) {
      addFinding('warning', 'hard-coded-color', location, tag, className)
    }
    if (/^z-\[/.test(base)) {
      addFinding('warning', 'arbitrary-z-index', location, tag, className)
    }
  }

  const conflicts = new Map()
  for (const className of classes) {
    const { variant, base } = splitVariant(className)
    const group = conflictGroup(base)
    if (!group) continue
    const key = `${variant}|${group}`
    const previous = conflicts.get(key)
    if (previous && previous !== className) {
      addFinding(
        'error',
        'conflicting-class',
        location,
        tag,
        `“${previous}” conflicts with “${className}”`,
      )
    } else {
      conflicts.set(key, className)
    }
  }
}

function scanSemantics({
  tag,
  native,
  attributes,
  children,
  ancestors,
  classes,
  location,
}) {
  if (tag === 'img' && !attributes.alt && !attributes.__spread) {
    addFinding('error', 'image-alt', location, tag, 'missing alt attribute')
  }
  if (tag === 'iframe' && !attributes.title && !attributes.__spread) {
    addFinding('error', 'iframe-title', location, tag, 'missing title attribute')
  }
  if (
    tag === 'a' &&
    attributes.target?.value === '_blank' &&
    (!attributes.rel ||
      (attributes.rel.strings.length === 1 &&
        !/\b(?:noopener|noreferrer)\b/.test(attributes.rel.strings[0])))
  ) {
    addFinding('error', 'unsafe-blank-target', location, tag, 'target="_blank" without noopener or noreferrer')
  }
  if (
    native &&
    !focusableTags.has(tag) &&
    attributes.onClick &&
    !attributes.role &&
    !attributes.tabIndex &&
    !attributes['data-event-delegation']
  ) {
    addFinding('warning', 'clickable-noninteractive', location, tag, 'onClick without role or tabIndex')
  }
  if (attributes.role?.value === 'dialog' && tag !== 'dialog' && tag !== 'DialogContent') {
    addFinding('warning', 'raw-dialog', location, tag, 'prefer the shared Dialog component')
  }
  if (
    (tag === 'button' || tag === 'Button') &&
    !attributes['aria-label'] &&
    !attributes.title &&
    !attributes.__spread &&
    !hasAccessibleChildren(children)
  ) {
    addFinding('warning', 'button-name', location, tag, 'no static accessible name detected')
  }

  const clipsFocus = ancestors
    .slice(-2)
    .some(parent => parent.classes.includes('overflow-hidden'))
  if (
    clipsFocus &&
    focusableTags.has(tag) &&
    classes.some(className => /(?:^|:)focus-visible:(?:ring|outline)-/.test(className))
  ) {
    addFinding('warning', 'focus-clip-risk', location, tag, 'focus style is inside an overflow-hidden ancestor')
  }
}

function readAttributes(properties, sourceFile) {
  const attributes = {}
  for (const property of properties) {
    if (ts.isJsxSpreadAttribute(property)) {
      attributes.__spread = { value: '', strings: [] }
      continue
    }
    if (!ts.isJsxAttribute(property)) continue
    const name = property.name.getText(sourceFile)
    if (!property.initializer) {
      attributes[name] = { value: '', strings: [] }
      continue
    }
    if (ts.isStringLiteral(property.initializer)) {
      attributes[name] = {
        value: property.initializer.text,
        strings: [property.initializer.text],
      }
      continue
    }
    const expression = property.initializer.expression
    const strings = expression ? collectStrings(expression) : []
    attributes[name] = {
      value: strings.length === 1 ? strings[0] : '',
      strings,
    }
  }
  return attributes
}

function collectStrings(node) {
  const values = []
  visit(node)
  return values

  function visit(current) {
    if (ts.isStringLiteralLike(current) || ts.isNoSubstitutionTemplateLiteral(current)) {
      values.push(current.text)
      return
    }
    if (ts.isTemplateExpression(current)) {
      values.push(current.head.text)
      for (const span of current.templateSpans) {
        visit(span.expression)
        values.push(span.literal.text)
      }
      return
    }
    ts.forEachChild(current, visit)
  }
}

function hasAccessibleChildren(children) {
  return children.some(child => {
    if (ts.isJsxText(child)) return Boolean(child.text.trim())
    if (ts.isJsxElement(child) || ts.isJsxSelfClosingElement(child)) return true
    if (ts.isJsxExpression(child)) return Boolean(child.expression)
    return false
  })
}

function splitClasses(value) {
  return value.trim().split(/\s+/).filter(Boolean)
}

function splitVariant(className) {
  let bracketDepth = 0
  let lastColon = -1
  for (let index = 0; index < className.length; index += 1) {
    const character = className[index]
    if (character === '[' || character === '(') bracketDepth += 1
    else if (character === ']' || character === ')') bracketDepth = Math.max(0, bracketDepth - 1)
    else if (character === ':' && bracketDepth === 0) lastColon = index
  }
  return lastColon === -1
    ? { variant: '', base: className }
    : { variant: className.slice(0, lastColon), base: className.slice(lastColon + 1) }
}

function conflictGroup(base) {
  if (/^(?:block|inline|inline-block|flex|inline-flex|grid|inline-grid|hidden|contents|table)$/.test(base)) return 'display'
  if (/^(?:static|fixed|absolute|relative|sticky)$/.test(base)) return 'position'
  if (/^ring(?:-(?:0|1|2|4|8))?$/.test(base)) return 'ring-width'
  if (/^resize(?:-none|-x|-y)?$/.test(base)) return 'resize'
  return ''
}

function isIntentionalColor(location, className) {
  const file = location.slice(0, location.lastIndexOf(':'))
  if (['dialog', 'drawer', 'sheet'].some(name => file === `packages/ui/src/components/${name}.tsx`) && className === 'bg-black/10') return true
  if (file === 'packages/ui/src/components/slider.tsx' && className === 'bg-white') return true
  return file === 'packages/theme-default/src/site/pages/topic-page.tsx' && className === 'bg-black/90'
}

function addFinding(severity, code, location, tag, detail) {
  findings.push({ severity, code, location, tag, detail })
}

function printReport() {
  const report = {
    scannedFiles: inventory.scannedFiles,
    elementCount: sum(inventory.elements),
    classCount: sum(inventory.classes),
    uniqueClasses: inventory.classes.size,
    findings,
    elements: mapToRows(inventory.elements),
    nativeElements: mapToRows(inventory.nativeElements),
    components: mapToRows(inventory.components),
    classes: mapToRows(inventory.classes),
    arbitraryClasses: mapToRows(inventory.arbitraryClasses),
  }

  if (json) {
    console.log(JSON.stringify(report, null, 2))
    return
  }

  console.log(`Scanned ${report.scannedFiles} TSX files.`)
  console.log(`Found ${report.elementCount} JSX elements and ${report.classCount} class usages (${report.uniqueClasses} unique).\n`)
  printInventory('native DOM elements', inventory.nativeElements, nativeTags, 20)
  printInventory('shared / custom components', inventory.components, null, 20)
  printInventory('most used classes', inventory.classes, null, 25)
  printInventory('arbitrary-value classes', inventory.arbitraryClasses, null, 20)

  console.log('findings')
  console.log('--------')
  if (!findings.length) {
    console.log('  none')
    return
  }
  for (const item of findings.slice(0, maxFindings)) {
    console.log(`  ${item.severity.toUpperCase()} ${item.code} ${item.location} <${item.tag}> ${item.detail}`)
  }
  if (findings.length > maxFindings) console.log(`  ... ${findings.length - maxFindings} more; use --json for the full report`)
  console.log(`\n${errors.length} error(s), ${findings.length - errors.length} warning(s)`)
}

function printInventory(title, map, allowlist, limit) {
  console.log(title)
  console.log('-'.repeat(title.length))
  const rows = mapToRows(map)
    .filter(item => !allowlist || allowlist.has(item.name))
    .slice(0, limit)
  if (!rows.length) console.log('  none')
  else for (const row of rows) console.log(`  ${row.name}: ${row.count}`)
  console.log()
}

function mapToRows(map) {
  return [...map.entries()]
    .map(([name, entry]) => ({ name, count: entry.count, locations: [...entry.locations] }))
    .sort((left, right) => right.count - left.count || left.name.localeCompare(right.name))
}

function increment(map, name, location) {
  const entry = map.get(name) || { count: 0, locations: new Set() }
  entry.count += 1
  entry.locations.add(location)
  map.set(name, entry)
}

function sum(map) {
  return [...map.values()].reduce((total, entry) => total + entry.count, 0)
}

function readOption(name) {
  const index = process.argv.indexOf(name)
  const value = index >= 0 ? process.argv[index + 1] : undefined
  return value && !value.startsWith('--') ? value : undefined
}

function readOptions(name) {
  return process.argv.flatMap((value, index) =>
    value === name && process.argv[index + 1] && !process.argv[index + 1].startsWith('--')
      ? [process.argv[index + 1]]
      : [],
  )
}

function collectFiles(path) {
  const stat = statSync(path)
  if (stat.isFile()) return path.endsWith('.tsx') && !path.endsWith('.test.tsx') ? [path] : []
  return readdirSync(path).flatMap(name => collectFiles(join(path, name)))
}

function toPosix(path) {
  return toNamespacedPath(path).replaceAll('\\', '/')
}
