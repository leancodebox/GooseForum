import { readFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'

const backendFile = new URL('../../app/http/controllers/component/message_code.go', import.meta.url)

export function parseBackendMessageCodes(source) {
  return [...source.matchAll(/\b\w+\s+MessageCode\s*=\s*"([^"]+)"/g)].map(match => match[1])
}

export function validateServerMessages(backendCodes, frontendKeys) {
  const expected = new Set(backendCodes)
  const actual = new Set(frontendKeys)
  return [
    ...(expected.size === backendCodes.length ? [] : ['Duplicate Go MessageCode declarations']),
    ...[...expected].filter(code => !actual.has(code)).map(code => 'Missing translation key ' + code),
    ...[...actual].filter(code => !expected.has(code)).map(code => 'Unknown Go MessageCode ' + code),
  ]
}

export async function checkServerCodes(catalog) {
  const codes = parseBackendMessageCodes(await readFile(backendFile, 'utf8'))
  const errors = validateServerMessages(codes, Object.keys(catalog.get('shared.serverMessages') ?? {}))
  if (errors.length) throw Error(errors.join('\n'))
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const { readCatalog } = await import('./i18n-catalog.mjs')
  await checkServerCodes(await readCatalog())
}
