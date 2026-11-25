import { readFile, writeFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { dirname, join } from 'node:path'
import iconImports from 'lucide-react/dynamicIconImports'

const require = createRequire(import.meta.url)
const root = dirname(require.resolve('lucide-react/package.json'))
const license = await readFile(join(root, 'LICENSE'), 'utf8')
const entries = await Promise.all(Object.entries(iconImports).map(async ([name, load]) =>
  [name, (await load()).__iconData],
))
// One icon per line keeps library updates reviewable.
const icons = []
const aliases = {}
for (const [name, data] of entries) {
  if (name === data.name) icons.push(JSON.stringify(name) + ':' + JSON.stringify(data))
  else aliases[name] = data.name
}
const catalog = '{"license":' + JSON.stringify(license) + ',"aliases":' + JSON.stringify(aliases)
  + ',"icons":{\n' + icons.join(',\n') + '\n}}\n'
const output = new URL('../../icons/lucide.json', import.meta.url)
if (await readFile(output, 'utf8').catch(() => '') !== catalog) await writeFile(output, catalog)
