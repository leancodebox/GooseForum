// Run the production dictionary loaders in an isolated audit process. Mark literal
// values before spreads/legacy merges so the final leaves retain their origin.
import { registerHooks } from 'node:module'
import { existsSync, readFileSync } from 'node:fs'
import { resolve, relative } from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'
import ts from 'typescript'

export const root = fileURLToPath(new URL('../../', import.meta.url))
export const client = 'resource/packages/client/src/i18n'
export const admin = 'resource/apps/web/src/admin'
export const server = 'app/bundles/i18n/locales'

export function literals(source, file) {
  const ast = ts.createSourceFile(file, source, ts.ScriptTarget.Latest, true)
  const result = []
  function walk(node, path) {
    if (ts.isPropertyAssignment(node)) {
      const keys = [...path, node.name.text]
      if (ts.isStringLiteralLike(node.initializer)) {
        result.push({ key: keys.join('.'), value: node.initializer.text,
          start: node.initializer.getStart(ast), end: node.initializer.end })
      } else walk(node.initializer, keys)
    } else ts.forEachChild(node, child => walk(child, path))
  }
  walk(ast, [])
  return result
}

export function flatten(object, prefix = '', leaves = new Map()) {
  for (const [key, value] of Object.entries(object)) {
    const path = prefix ? prefix + '.' + key : key
    if (typeof value === 'string') leaves.set(path, value)
    else if (value && typeof value === 'object') flatten(value, path, leaves)
    else throw Error('Unexpected translation value at ' + path)
  }
  return leaves
}

export async function catalog() {
  const origins = new Map()
  const url = path => pathToFileURL(resolve(root, path)).href
  const hook = registerHooks({
    resolve(specifier, context, next) {
      if (specifier.startsWith('@gooseforum/client/i18n/'))
        return { url: url(client + '/' + specifier.split('/i18n/')[1] + '.ts'), shortCircuit: true }
      if (specifier.startsWith('.') && context.parentURL?.startsWith(url('resource') + '/')) {
        const candidate = new URL(specifier.replace(/\.js$/, '') + '.ts', context.parentURL)
        if (existsSync(candidate)) return { url: candidate.href, shortCircuit: true }
      }
      return next(specifier, context)
    },
    load(target, context, next) {
      if (!target.startsWith(url('resource') + '/') || !target.endsWith('.ts')) return next(target, context)
      const file = relative(root, fileURLToPath(target))
      let source = readFileSync(fileURLToPath(target), 'utf8')
      if (file.includes('/messages/') || file === client + '/locale.ts') {
        const entries = literals(source, file)
        for (const entry of entries.reverse()) {
          const marker = '__i18n_origin_' + origins.size
          origins.set(marker, { file, key: entry.key, value: entry.value })
          source = source.slice(0, entry.start) + JSON.stringify(marker) + source.slice(entry.end)
        }
      }
      return { format: 'module', shortCircuit: true, source: ts.transpile(source, { module: ts.ModuleKind.ESNext, target: ts.ScriptTarget.ESNext }) }
    },
  })
  try {
    const { supportedLocales } = await import(url(client + '/locale.ts'))
    const site = await import(url('resource/packages/runtime/src/i18n/loader.ts'))
    const { translationLoaders } = await import(url(admin + '/translation-resources.ts'))
    const namespaces = [...new Set(Object.keys(translationLoaders).map(key => key.split('/')[1]))]
    const rows = new Map()
    function add(locale, namespace, source, file) {
      for (const [key, raw] of flatten(source)) {
        const id = namespace + ':' + key
        if (!rows.has(id)) rows.set(id, {})
        rows.get(id)[locale] = origins.get(raw) ?? { file, key, value: raw }
      }
    }
    for (const locale of supportedLocales) {
      await site.prepareGooseTranslations(locale, site.gooseNamespaces)
      for (const [ns, obj] of Object.entries(site.cachedGooseResources()[locale])) add(locale, 'site/' + ns, obj)
      for (const ns of namespaces) add(locale, 'admin/' + ns, (await translationLoaders[locale + '/' + ns]()).default)
      const file = server + '/' + locale + '.json'
      const obj = JSON.parse(readFileSync(resolve(root, file), 'utf8'))
      add(locale, 'go', obj, file)
    }
    return { rows, locales: supportedLocales }
  } finally { hook.deregister() }
}
