#!/usr/bin/env node
import { readFileSync, statSync } from 'node:fs'
import { gzipSync } from 'node:zlib'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const resourceRoot = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const outputRoot = join(resourceRoot, 'static/dist/react')
const manifestPath = join(outputRoot, '.vite/manifest.json')
const manifest = JSON.parse(readFileSync(manifestPath, 'utf8'))
const scenarios = {
  'site shell': {
    roots: ['index.html'],
    budget: 198 * 1024,
  },
  'home route (zh)': {
    roots: [
      'index.html',
      '../../packages/theme-default/src/site/pages/home-page.tsx',
      '../../packages/client/dist/i18n/messages/zh-auth.js',
      '../../packages/client/dist/i18n/messages/zh-site-shell.js',
      '../../packages/client/dist/i18n/messages/zh-site-home.js',
      '../../packages/client/dist/i18n/messages/zh-site-userCard.js',
      '../../packages/client/dist/i18n/messages/zh-content-common.js',
      '../../packages/client/dist/i18n/messages/zh-server-messages.js',
    ],
    budget: 215 * 1024,
  },
  'admin shell': {
    roots: ['admin/index.html'],
    budget: 195 * 1024,
  },
  'dashboard route (zh, chart deferred)': {
    roots: [
      'admin/index.html',
      'src/admin/pages/dashboard-page.tsx',
      'src/admin/messages/zh-shell.ts',
      'src/admin/messages/zh-dashboard.ts',
    ],
    budget: 200 * 1024,
  },
}
const maxChunkBudget = 150 * 1024

let failed = false
for (const [name, { roots, budget }] of Object.entries(scenarios)) {
  const files = collectInitialFiles(roots)
  const rawBytes = files.reduce((total, file) => total + statSync(join(outputRoot, file)).size, 0)
  const gzipBytes = files.reduce(
    (total, file) => total + gzipSync(readFileSync(join(outputRoot, file))).byteLength,
    0,
  )
  const status = gzipBytes <= budget ? 'PASS' : 'FAIL'
  console.log(`${status} ${name}: ${format(gzipBytes)} gzip / ${format(budget)} budget (${format(rawBytes)} raw)`)
  if (gzipBytes > budget) failed = true
}

const chunks = [...new Set(Object.values(manifest).map(chunk => chunk.file))]
  .filter(file => file.endsWith('.js'))
  .map(file => ({ file, gzipBytes: gzipSync(readFileSync(join(outputRoot, file))).byteLength }))
  .sort((left, right) => right.gzipBytes - left.gzipBytes)
const largestChunk = chunks[0]
if (largestChunk) {
  const status = largestChunk.gzipBytes <= maxChunkBudget ? 'PASS' : 'FAIL'
  console.log(`${status} largest lazy chunk: ${largestChunk.file} is ${format(largestChunk.gzipBytes)} gzip / ${format(maxChunkBudget)} budget`)
  if (largestChunk.gzipBytes > maxChunkBudget) failed = true
}

const developmentMarker = 'GOOSE_DEV_ORIGIN'
const leakedDevelopmentHint = chunks.find(({ file }) =>
  readFileSync(join(outputRoot, file), 'utf8').includes(developmentMarker),
)
if (leakedDevelopmentHint) {
  console.log(`FAIL production bundle contains ${developmentMarker}: ${leakedDevelopmentHint.file}`)
  failed = true
} else {
  console.log(`PASS production bundle excludes ${developmentMarker}`)
}

if (failed) process.exitCode = 1

function collectInitialFiles(roots) {
  const files = new Set()
  const visited = new Set()
  for (const root of roots) visit(root)
  return [...files]

  function visit(key) {
    if (visited.has(key)) return
    visited.add(key)
    const chunk = manifest[key]
    if (!chunk) throw new Error(`Missing ${key} in ${manifestPath}`)
    files.add(chunk.file)
    for (const css of chunk.css || []) files.add(css)
    for (const imported of chunk.imports || []) visit(imported)
  }
}

function format(bytes) {
  return `${(bytes / 1024).toFixed(1)} KiB`
}
