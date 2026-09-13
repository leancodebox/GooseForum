#!/usr/bin/env node
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative, resolve, toNamespacedPath } from 'node:path'
import { parse } from 'vue/compiler-sfc'

const root = resolve(import.meta.dirname, '..')
const scanRoot = resolve(readOption('--scan-root') || join(root, 'src/site'))
const json = process.argv.includes('--json')
const failOnLegacy = process.argv.includes('--fail-on-legacy')
const failOnNativeControls = process.argv.includes('--fail-on-native-controls')
const allowDefaultButtons = process.argv.includes('--allow-default-buttons')

const nativeInteractiveTags = new Set(['a', 'button', 'input', 'select', 'textarea'])
const allowedNativeInputTypes = new Set(['file', 'hidden', 'range'])
const shadcnComponents = new Set([
  'AlertDialog', 'Avatar', 'Badge', 'Button', 'Card', 'Chart', 'Checkbox',
  'Command', 'Dialog', 'DropdownMenu', 'Input', 'InputGroup', 'Label',
  'Popover', 'RadioGroup', 'RangeCalendar', 'Select', 'Separator', 'Sheet',
  'Skeleton', 'Switch', 'Table', 'Tabs', 'Textarea', 'Tooltip',
])
const legacyClassPattern = /gf-(?:avatar|badge|button|composer-tool|icon-button|input|list-mode|locale-switch(?:-item)?|menu-surface|segmented(?:-item)?|tab|textarea)\b/
const overlayClassPattern = /\b(?:fixed|absolute)\b/

const files = collectFiles(scanRoot)
const inventory = {
  scannedFiles: files.length,
  nativeElements: {},
  components: {},
  legacyClasses: [],
  rawDialogs: [],
  overlays: [],
  globalEvents: [],
  nativeControlRisks: [],
  tabHeightRisks: [],
  buttonVariantRisks: [],
}

for (const file of files) {
  const source = readFileSync(file, 'utf8')
  const rel = toPosix(relative(root, file))
  const { descriptor, errors } = parse(source, { filename: file })

  if (errors.length) {
    console.error(`Unable to parse ${rel}`)
    console.error(errors.map(error => error.message).join('\n'))
    process.exitCode = 1
    continue
  }

  if (descriptor.template?.ast) walkTemplate(descriptor.template.ast, rel)
  scanScript(source, rel)
  scanTabHeightRisks(source, rel)
}

printReport()

function walkTemplate(node, rel) {
  if (!node || typeof node !== 'object') return

  if (node.tag) {
    const component = isComponent(node.tag)
    const attrs = readAttrs(node)
    const entry = component
      ? getGroup(inventory.components, node.tag)
      : getGroup(inventory.nativeElements, node.tag)
    entry.count += 1
    entry.locations.push(location(rel, node.loc.start.line, node.tag))

    if (!allowDefaultButtons && node.tag === 'Button' && !hasVariantProp(node)) {
      inventory.buttonVariantRisks.push(location(rel, node.loc.start.line, node.tag, 'missing variant (defaults to primary)'))
    }

    if (attrs.classes.some(value => legacyClassPattern.test(value))) {
      inventory.legacyClasses.push(location(rel, node.loc.start.line, node.tag, attrs.classes.join(' ')))
    }

    if (!component) {
      if (isNativeControlRisk(node.tag, attrs)) {
        inventory.nativeControlRisks.push(location(rel, node.loc.start.line, node.tag, attrs.type ? `type=${attrs.type}` : ''))
      }
      if (attrs.classes.some(value => overlayClassPattern.test(value))) {
        inventory.overlays.push(location(rel, node.loc.start.line, node.tag, attrs.classes.join(' ')))
      }
      if (attrs.role === 'dialog' || attrs['aria-modal'] === 'true') {
        inventory.rawDialogs.push(location(rel, node.loc.start.line, node.tag, attrs.classes.join(' ')))
      }
    }

  }

  for (const child of node.children ?? []) walkTemplate(child, rel)
  for (const branch of node.branches ?? []) walkTemplate(branch, rel)
}

function readAttrs(node) {
  const classes = []
  const attrs = {}

  for (const prop of node.props ?? []) {
    if (prop.type !== 6 && prop.type !== 7) continue

    if (prop.type === 6) {
      if (prop.name === 'class' && prop.value?.content) classes.push(prop.value.content)
      else attrs[prop.name] = prop.value?.content ?? ''
      continue
    }

    if (prop.name === 'bind' && String(prop.arg?.content) === 'class' && prop.exp?.content) {
      classes.push(prop.exp.content)
    }
    if (prop.name === 'bind' && prop.arg?.content) attrs[String(prop.arg.content)] = prop.exp?.content ?? ''
    if (prop.name === 'on') attrs[`@${prop.arg?.content ?? ''}`] = prop.exp?.content ?? ''
  }

  return { classes, ...attrs }
}

function scanScript(source, rel) {
  const pattern = /(?:window|document)\.addEventListener\(\s*['"]([^'"]+)['"]/g
  for (const match of source.matchAll(pattern)) {
    const line = source.slice(0, match.index).split('\n').length
    inventory.globalEvents.push(location(rel, line, match[1]))
  }
}

function scanTabHeightRisks(source, rel) {
  const listPattern = /<TabsList\b[^>]*>/g

  for (const listMatch of source.matchAll(listPattern)) {
    const opening = listMatch[0]
    const listClass = opening.match(/class="([^"]+)"/)?.[1] ?? ''
    const listHeight = getExplicitHeight(listClass)
    if (!listHeight) continue

    const listStart = listMatch.index + listMatch[0].length
    const listEnd = source.indexOf('</TabsList>', listStart)
    if (listEnd === -1) continue

    const listSource = source.slice(listStart, listEnd)
    const triggerPattern = /<TabsTrigger\b[^>]*>/g
    const listLine = source.slice(0, listMatch.index).split('\n').length

    for (const triggerMatch of listSource.matchAll(triggerPattern)) {
      const triggerClass = triggerMatch[0].match(/class="([^"]+)"/)?.[1] ?? ''
      const triggerHeight = getExplicitHeight(triggerClass)
      if (!triggerHeight || triggerHeight < listHeight) continue

      const absoluteOffset = listStart + triggerMatch.index
      const line = source.slice(0, absoluteOffset).split('\n').length
      inventory.tabHeightRisks.push(location(rel, line, 'TabsTrigger', `h-${triggerHeight} inside h-${listHeight}`))
    }

    if (!listSource.includes('<TabsTrigger')) {
      inventory.tabHeightRisks.push(location(rel, listLine, 'TabsList', `h-${listHeight} without triggers`))
    }
  }
}

function hasVariantProp(node) {
  return (node.props ?? []).some(prop => prop.name === 'variant' || prop.arg?.content === 'variant')
}

function getExplicitHeight(className) {
  const match = className.match(/\bh-(\d+(?:\.\d+)?)\b/)
  return match ? Number(match[1]) : null
}

function isNativeControlRisk(tag, attrs) {
  if (tag === 'button' || tag === 'select' || tag === 'textarea') return true
  if (tag !== 'input') return false
  return !allowedNativeInputTypes.has(attrs.type || 'text')
}

function printReport() {
  const nativeButtonCount = inventory.nativeElements.button?.count ?? 0
  if (failOnLegacy && (
    nativeButtonCount
    || inventory.legacyClasses.length
    || inventory.rawDialogs.length
    || inventory.tabHeightRisks.length
    || inventory.buttonVariantRisks.length
  )) {
    process.exitCode = 1
  }
  if (failOnNativeControls && (
    inventory.nativeControlRisks.length
    || inventory.legacyClasses.length
    || inventory.rawDialogs.length
    || inventory.tabHeightRisks.length
  )) {
    process.exitCode = 1
  }

  if (json) {
    console.log(JSON.stringify(inventory, null, 2))
    return
  }

  console.log(`Scanned ${inventory.scannedFiles} SFC files in ${toPosix(relative(root, scanRoot))}.\n`)

  printGrouped('shadcn-vue primitives', inventory.components, shadcnComponents)
  printGrouped('native interactive elements', inventory.nativeElements, nativeInteractiveTags)

  printFindings('legacy classes', inventory.legacyClasses)
  printFindings('raw dialogs', inventory.rawDialogs)
  printFindings('floating / fixed elements', inventory.overlays)
  printFindings('global event listeners', inventory.globalEvents)
  printFindings('native control risks', inventory.nativeControlRisks)
  printFindings('tab height risks', inventory.tabHeightRisks)
  printFindings('button variant risks', inventory.buttonVariantRisks)
}

function readOption(name) {
  const index = process.argv.indexOf(name)
  const value = index >= 0 ? process.argv[index + 1] : undefined
  return value && !value.startsWith('--') ? value : undefined
}

function printGrouped(title, groups, only) {
  console.log(title)
  console.log('-'.repeat(title.length))

  const rows = Object.entries(groups)
    .filter(([name]) => !only || only.has(name))
    .sort(([a], [b]) => a.localeCompare(b))

  if (!rows.length) {
    console.log('  none\n')
    return
  }

  for (const [name, entry] of rows) {
    console.log(`  ${name}: ${entry.count}`)
    for (const item of entry.locations.slice(0, 12)) {
      console.log(`    ${item}`)
    }
    if (entry.locations.length > 12) console.log(`    ... ${entry.locations.length - 12} more`)
  }
  console.log()
}

function printFindings(title, findings) {
  console.log(title)
  console.log('-'.repeat(title.length))

  if (!findings.length) {
    console.log('  none\n')
    return
  }

  for (const item of findings) console.log(`  ${item}`)
  console.log()
}

function getGroup(target, name) {
  return target[name] ??= { count: 0, locations: [] }
}

function location(file, line, tag, detail = '') {
  return `${file}:${line} ${tag}${detail ? ` ${detail}` : ''}`
}

function isComponent(tag) {
  return /^[A-Z]/.test(tag) || tag.includes('.')
}

function collectFiles(path) {
  const stat = statSync(path)
  if (stat.isFile()) return path.endsWith('.vue') ? [path] : []
  return readdirSync(path).flatMap(name => collectFiles(join(path, name)))
}

function toPosix(path) {
  return toNamespacedPath(path).replaceAll('\\', '/')
}
