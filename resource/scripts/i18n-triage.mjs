import ts from 'typescript'

function add(index, key, value) {
  if (!index.has(key)) index.set(key, new Set())
  index.get(key).add(value)
}

// This second pass collects weaker evidence for keys missed by the translator
// scanner. A matching spelling in another namespace is only a review lead.
function frontendHints(pending, sources) {
  const spellings = new Map()
  const prefixes = new Map()
  const hints = new Map()
  for (const { id } of pending) {
    const colon = id.indexOf(':')
    const namespace = id.slice(0, colon)
    const key = id.slice(colon + 1)
    if (namespace === 'go') continue
    const leaf = key.split('.').at(-1)
    for (const value of [key, leaf, namespace.split('/')[1] + ':' + key]) add(spellings, value, id)
    for (let i = key.indexOf('.'); i >= 0; i = key.indexOf('.', i + 1)) add(prefixes, key.slice(0, i + 1), id)
  }
  for (const [file, source] of sources) {
    const ast = ts.createSourceFile(file, source, ts.ScriptTarget.Latest, true)
    function visit(node) {
      if (ts.isTypeNode(node) || ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) return
      let matches
      if (ts.isStringLiteralLike(node)) matches = spellings.get(node.text)
      else if (ts.isPropertyAccessExpression(node)) matches = spellings.get(node.name.text)
      else if (ts.isBindingElement(node) && ts.isObjectBindingPattern(node.parent)) {
        matches = spellings.get((node.propertyName ?? node.name).text)
      } else if (ts.isPropertyAssignment(node) || ts.isShorthandPropertyAssignment(node)) {
        matches = spellings.get(node.name.text)
      } else if (ts.isTemplateExpression(node)) matches = prefixes.get(node.head.text)
      if (matches) {
        const at = file + ':' + (ast.getLineAndCharacterOfPosition(node.getStart(ast)).line + 1)
        for (const id of matches) add(hints, id, at)
      }
      ts.forEachChild(node, visit)
    }
    visit(ast)
  }
  return hints
}

export function groupUnresolved(inventory, goSources, frontendSources = []) {
  const pending = inventory.filter(row => row.status === 'no recognized reference')
  const hints = frontendHints(pending, frontendSources)
  const shared = new Map()
  const go = new Map()
  const origin = (locale, entry) => JSON.stringify([locale, entry.file, entry.key])
  for (const row of inventory.filter(row => row.status !== 'no recognized reference')) {
    for (const [locale, entry] of Object.entries(row.translations)) {
      add(shared, origin(locale, entry), row.id)
    }
  }
  const goKeys = new Set(pending.filter(row => row.id.startsWith('go:')).map(row => row.id.slice(3)))
  if (goKeys.size) for (const [file, source] of goSources) {
    for (const [i, line] of source.split('\n').entries()) {
      for (const match of line.matchAll(/"([\w.-]+)"/g)) {
        if (goKeys.has(match[1])) add(go, 'go:' + match[1], file + ':' + (i + 1))
      }
    }
  }

  const groups = {}
  for (const row of pending) {
    const sharedDefinitions = []
    for (const [locale, entry] of Object.entries(row.translations)) {
      const ids = entry.file && shared.get(origin(locale, entry))
      if (ids) sharedDefinitions.push({ locale, file: entry.file, key: entry.key, alsoUsedBy: [...ids] })
    }
    const evidence = {
      sharedDefinition: sharedDefinitions,
      goSourceLiteral: [...(go.get(row.id) ?? [])],
      frontendSpelling: [...(hints.get(row.id) ?? [])],
    }
    const signals = Object.fromEntries(Object.entries(evidence).filter(([, hits]) => hits.length))
    const reason = Object.keys(signals)[0] ?? 'unusedCandidate'
    const namespace = row.id.split(':')[0]
    const group = groups[reason] ??= {}
    const members = group[namespace] ??= []
    members.push({ id: row.id, definition: row.translations.zh?.file, evidence: signals })
  }
  return groups
}
