import assert from 'node:assert/strict'
import { readdirSync, readFileSync } from 'node:fs'
import { dirname, join, relative, resolve } from 'node:path'
import { test } from 'node:test'
import ts from 'typescript'

const root = resolve(import.meta.dirname, '..')
const layers = {
  client: [],
  runtime: ['client'],
  ui: [],
  'theme-default': ['client', 'runtime', 'ui'],
}

for (const [name, allowed] of Object.entries(layers)) {
  test(`${name} respects workspace dependency boundaries`, () => {
    const packageRoot = join(root, 'packages', name)
    for (const file of sources(join(packageRoot, 'src'))) {
      const source = readFileSync(file, 'utf8')
      for (const { fileName: specifier } of ts.preProcessFile(source, true, true).importedFiles) {
        if (specifier.startsWith('.')) {
          const target = relative(packageRoot, resolve(dirname(file), specifier))
          assert(!target.startsWith('..'), `${file}: cross-package relative import ${specifier}`)
        }
        if (specifier.startsWith('@gooseforum/')) {
          const dependency = specifier.split('/')[1]
          assert(dependency === name || allowed.includes(dependency), `${file}: forbidden dependency ${dependency}`)
        }
      }
    }
  })
}

test('admin and browser host do not import the default theme', () => {
  for (const directory of ['admin', 'host']) {
    for (const file of sources(join(root, 'apps/web/src', directory))) {
      assert(!readFileSync(file, 'utf8').includes('@gooseforum/theme-default'), file)
    }
  }
})

function sources(directory) {
  return readdirSync(directory, { withFileTypes: true }).flatMap(entry => {
    const file = join(directory, entry.name)
    return entry.isDirectory() ? sources(file) : /\.tsx?$/.test(entry.name) ? [file] : []
  })
}
