import ts from 'typescript'

// Resolve translator aliases in the nearest function scope. A name bound in
// another component, or an unrelated occurrence of the same text, is no evidence.
export function references(source, file, namespaceForImport) {
  const ast = ts.createSourceFile(file, source, ts.ScriptTarget.Latest, true)
  const namespaces = new Set()
  for (const stmt of ast.statements) {
    if (!ts.isImportDeclaration(stmt)) continue
    const clause = stmt.importClause
    const bindings = clause?.namedBindings
    if (!bindings || !ts.isNamedImports(bindings)) continue
    if (bindings.elements.some(el => (clause.isTypeOnly || el.isTypeOnly) &&
        /(?:TextKey|SettingsText)$/.test((el.propertyName ?? el.name).text))) {
      const ns = namespaceForImport?.(stmt.moduleSpecifier.text)
      if (ns) namespaces.add(ns)
    }
  }
  const adminNs = namespaces.size === 1 ? [...namespaces][0] : undefined
  const hits = []
  let checker
  const tables = new Map()
  function tableValues(expr) {
    if (!ts.isElementAccessExpression(expr) || !ts.isIdentifier(expr.expression)) return
    checker ??= localChecker(ast)
    const symbol = checker.getSymbolAtLocation(expr.expression)
    const decl = symbol?.valueDeclaration
    if (!decl || !ts.isVariableDeclaration(decl) || !(decl.parent.flags & ts.NodeFlags.Const) ||
        !decl.initializer || !ts.isObjectLiteralExpression(decl.initializer) ||
        decl.parent.parent.modifiers?.some(m => m.kind === ts.SyntaxKind.ExportKeyword)) return
    if (tables.has(symbol)) return tables.get(symbol)
    const fields = decl.initializer.properties
    let safe = fields.length > 0 && fields.every(p => ts.isPropertyAssignment(p) &&
      !ts.isComputedPropertyName(p.name) && ts.isStringLiteralLike(p.initializer))
    // Only narrow private tables used by simple value reads. Aliases, writes,
    // and escapes could introduce keys outside the initializer.
    function check(node) {
      if (!safe) return
      if (ts.isIdentifier(node) && node.text === expr.expression.text && node !== decl.name &&
          (checker.getSymbolAtLocation(node) === symbol ||
           (ts.isShorthandPropertyAssignment(node.parent) && checker.getShorthandAssignmentValueSymbol(node.parent) === symbol))) {
        const access = node.parent
        const use = access.parent
        safe = ts.isElementAccessExpression(access) && access.expression === node &&
          ((ts.isBinaryExpression(use) && use.left === access && use.operatorToken.kind === ts.SyntaxKind.BarBarToken) ||
           (ts.isConditionalExpression(use) && use.condition === access))
      }
      ts.forEachChild(node, check)
    }
    check(ast)
    const values = safe ? fields.map(p => p.initializer.text) : undefined
    tables.set(symbol, values)
    return values
  }
  // Follow direct translator props into local function components. Collect all
  // call sites first: mixed namespaces or an unrecognised prop stay unresolved.
  const components = new Map()
  let props = new Map()
  let nextProps
  const spreadProps = new Set()
  for (const stmt of ast.statements) {
    if (ts.isFunctionDeclaration(stmt) && stmt.name) components.set(stmt.name.text, stmt)
  }
  function visit(node, scope, collecting) {
    if (ts.isFunctionLike(node) || ts.isBlock(node)) scope = new Map(scope)
    if (ts.isFunctionLike(node)) {
      for (const param of node.parameters) {
        if (ts.isIdentifier(param.name)) scope.delete(param.name.text)
        else for (const el of param.name.elements) {
          if (!el.name) continue
          scope.delete(el.name.text)
          const prop = (el.propertyName ?? el.name).text
          const sites = props.get(node)?.get(prop)
          if (!spreadProps.has(node) && sites?.size === 1 && !sites.has(undefined)) scope.set(el.name.text, [...sites][0])
          if (adminNs && prop === 'text') scope.set(el.name.text, 'admin/' + adminNs)
        }
      }
    }
    if (ts.isVariableDeclaration(node)) {
      const init = node.initializer
      if (ts.isObjectBindingPattern(node.name)) {
        for (const el of node.name.elements) scope.delete(el.name.text)
      }
      if (ts.isObjectBindingPattern(node.name) && init && ts.isCallExpression(init) &&
          init.expression.getText(ast) === 'useTranslation' && ts.isStringLiteralLike(init.arguments[0] ?? init)) {
        for (const el of node.name.elements)
          if ((el.propertyName ?? el.name).text === 't') scope.set(el.name.text, 'site/' + init.arguments[0].text)
      } else if (ts.isIdentifier(node.name)) scope.delete(node.name.text)
    }
    if (collecting && (ts.isJsxOpeningElement(node) || ts.isJsxSelfClosingElement(node))) {
      const target = components.get(node.tagName.getText(ast))
      if (target) {
        if (!nextProps.has(target)) nextProps.set(target, new Map())
        const bindings = nextProps.get(target)
        const supplied = new Set()
        for (const attr of node.attributes.properties) {
          if (ts.isJsxSpreadAttribute(attr)) { spreadProps.add(target); continue }
          const prop = attr.name.getText(ast)
          supplied.add(prop)
          if (!bindings.has(prop)) bindings.set(prop, new Set())
          const value = attr.initializer
          const expr = value && ts.isJsxExpression(value) && value.expression
          bindings.get(prop).add(expr && ts.isIdentifier(expr) ? scope.get(expr.text) : undefined)
        }
        const param = target.parameters[0]?.name
        if (param && ts.isObjectBindingPattern(param)) for (const el of param.elements) {
          const name = (el.propertyName ?? el.name).text
          if (supplied.has(name)) continue
          if (!bindings.has(name)) bindings.set(name, new Set())
          bindings.get(name).add(undefined)
        }
      }
    }
    if (!collecting && ts.isCallExpression(node) && ts.isIdentifier(node.expression) && scope.has(node.expression.text)) {
      const arg = node.arguments[0]
      let ns = scope.get(node.expression.text)
      const options = node.arguments[1]
      if (ns.startsWith('site/') && options && ts.isObjectLiteralExpression(options)) {
        const override = options.properties.find(p => p.name?.getText(ast) === 'ns')
        if (override) ns = ts.isPropertyAssignment(override) && ts.isStringLiteralLike(override.initializer)
          ? 'site/' + override.initializer.text : undefined
      }
      const at = file + ':' + (ast.getLineAndCharacterOfPosition(node.getStart(ast)).line + 1)
      function record(expr) {
        if (!expr || !ns) return
        if (ts.isStringLiteralLike(expr)) {
          const id = ns.startsWith('site/') && expr.text.includes(':') ? 'site/' + expr.text : ns + ':' + expr.text
          hits.push({ id, at })
        }
        else if (ts.isConditionalExpression(expr)) { record(expr.whenTrue); record(expr.whenFalse) }
        else if (ts.isBinaryExpression(expr) && expr.operatorToken.kind === ts.SyntaxKind.BarBarToken) {
          const values = tableValues(expr.left)
          if (values) for (const value of values) hits.push({ candidate: ns + ':' + value, at })
          else record(expr.left)
          record(expr.right)
        }
        else hits.push({ prefix: ns + ':' + (ts.isTemplateExpression(expr) ? expr.head.text : ''), at })
      }
      record(arg)
    }
    ts.forEachChild(node, child => visit(child, scope, collecting))
  }
  // Start with unknown prop bindings and propagate one local JSX hop per
  // pass. A mixed or unknown caller prevents a unique namespace downstream.
  for (let pass = 0; pass < components.size; pass++) {
    nextProps = new Map()
    visit(ast, new Map(), true)
    const changed = [...nextProps].some(([node, bindings]) => [...bindings].some(([name, values]) => {
      const old = props.get(node)?.get(name)
      return old?.size !== values.size || [...values].some(value => !old.has(value))
    }))
    props = nextProps
    if (!changed) break
  }
  visit(ast, new Map(), false)
  return hits
}

// Only actual TS/TSX string literals count as evidence of pre-existing inline
// copy. Comments, translation files and substring matches cannot establish it.
export function inlineStrings(source, file) {
  const ast = ts.createSourceFile(file, source, ts.ScriptTarget.Latest, true)
  const strings = new Map()
  function visit(node) {
    if (ts.isStringLiteralLike(node) && !strings.has(node.text))
      strings.set(node.text, ast.getLineAndCharacterOfPosition(node.getStart(ast)).line + 1)
    ts.forEachChild(node, visit)
  }
  visit(ast)
  return strings
}

// Resolve const aliases by TypeScript symbol, not variable name. Only files
// importing the admin loader need a checker; no dependency graph is loaded.
export function adminDictionaryReads(source, file, keys) {
  if (!source.includes('getAdminDictionary')) return []
  const ast = ts.createSourceFile(file, source, ts.ScriptTarget.Latest, true)
  const checker = localChecker(ast)
  const hits = []
  function pathOf(node, seen = new Set()) {
    if (ts.isParenthesizedExpression(node) || ts.isAsExpression(node)) return pathOf(node.expression, seen)
    if (ts.isIdentifier(node)) {
      const declaration = checker.getSymbolAtLocation(node)?.valueDeclaration
      if (!declaration || !ts.isVariableDeclaration(declaration) || !declaration.initializer ||
          !(declaration.parent.flags & ts.NodeFlags.Const) || seen.has(declaration)) return
      seen.add(declaration)
      return pathOf(declaration.initializer, seen)
    }
    if (ts.isCallExpression(node) && ts.isIdentifier(node.expression)) {
      const decl = checker.getSymbolAtLocation(node.expression)?.declarations?.[0]
      const ns = node.arguments[0]
      if (decl && ts.isImportSpecifier(decl) && (decl.propertyName ?? decl.name).text === 'getAdminDictionary' &&
          /(?:^|\/)translation-loader$/.test(decl.parent.parent.parent.moduleSpecifier.text) &&
          ns && ts.isStringLiteralLike(ns)) return 'admin/' + ns.text + ':'
      return
    }
    if (!ts.isPropertyAccessExpression(node) && !ts.isElementAccessExpression(node)) return
    const parent = pathOf(node.expression, seen)
    if (!parent || keys.has(parent)) return // e.g. .replace() on translated text
    const field = ts.isPropertyAccessExpression(node) ? node.name : node.argumentExpression
    if (ts.isElementAccessExpression(node) && !ts.isStringLiteralLike(field)) return
    return parent + (parent.endsWith(':') ? '' : '.') + field.text
  }
  function visit(node) {
    if (ts.isTypeNode(node)) return
    if (ts.isPropertyAccessExpression(node) || ts.isElementAccessExpression(node)) {
      const path = pathOf(node)
      const at = file + ':' + (ast.getLineAndCharacterOfPosition(node.getStart(ast)).line + 1)
      if (path && keys.has(path)) hits.push({ id: path, at })
      else if (ts.isElementAccessExpression(node) && !ts.isStringLiteralLike(node.argumentExpression)) {
        const base = pathOf(node.expression)
        if (base && !base.endsWith(':') && !keys.has(base)) hits.push({ prefix: base + '.', at })
      }
    }
    ts.forEachChild(node, visit)
  }
  visit(ast)
  return hits
}

function localChecker(ast) {
  const options = { noLib: true, noResolve: true, jsx: ts.JsxEmit.Preserve }
  const host = ts.createCompilerHost(options)
  host.getSourceFile = name => name === ast.fileName ? ast : undefined
  return ts.createProgram([ast.fileName], options, host).getTypeChecker()
}
