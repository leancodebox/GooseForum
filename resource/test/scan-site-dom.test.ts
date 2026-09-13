import { mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { resolve } from 'node:path'
import { spawnSync } from 'node:child_process'
import { afterEach, describe, expect, it } from 'vitest'

const scanner = resolve(import.meta.dirname, '../scripts/scan-site-dom.mjs')
const temporaryDirectories: string[] = []

afterEach(() => {
  for (const directory of temporaryDirectories.splice(0)) rmSync(directory, { force: true, recursive: true })
})

describe('site DOM scanner', () => {
  it('detects legacy classes and raw dialogs on non-interactive elements', () => {
    const directory = createFixture(`
      <template>
        <div role="dialog" aria-modal="true" class="fixed inset-0 gf-menu-surface">
          <p>Legacy dialog</p>
        </div>
      </template>
    `)

    const result = runScanner(directory, '--json')
    const inventory = JSON.parse(result.stdout)

    expect(result.status).toBe(0)
    expect(inventory.legacyClasses).toHaveLength(1)
    expect(inventory.rawDialogs).toHaveLength(1)
    expect(inventory.overlays).toHaveLength(1)
  })

  it('fails the guard for native buttons and component risks', () => {
    const directory = createFixture(`
      <template>
        <button type="button">Native action</button>
        <Button>Missing variant</Button>
        <TabsList class="h-8">
          <TabsTrigger class="h-8" value="all">All</TabsTrigger>
        </TabsList>
      </template>
    `)

    const report = runScanner(directory, '--json')
    const inventory = JSON.parse(report.stdout)
    const guarded = runScanner(directory, '--fail-on-legacy')
    const guardedJson = runScanner(directory, '--fail-on-legacy', '--json')

    expect(inventory.nativeElements.button.count).toBe(1)
    expect(inventory.buttonVariantRisks).toHaveLength(1)
    expect(inventory.tabHeightRisks).toHaveLength(1)
    expect(guarded.status).toBe(1)
    expect(guardedJson.status).toBe(1)
  })

  it('allows file inputs while guarding admin native controls', () => {
    const directory = createFixture(`
      <template>
        <input type="file" />
        <input />
        <select><option value="all">All</option></select>
        <textarea />
      </template>
    `)

    const report = runScanner(directory, '--json')
    const inventory = JSON.parse(report.stdout)
    const guarded = runScanner(directory, '--fail-on-native-controls', '--json')

    expect(inventory.nativeControlRisks).toHaveLength(3)
    expect(inventory.nativeControlRisks.some((item: string) => item.includes('type=file'))).toBe(false)
    expect(guarded.status).toBe(1)
  })
})

function createFixture(source: string) {
  const directory = mkdtempSync(resolve(tmpdir(), 'gooseforum-dom-scan-'))
  temporaryDirectories.push(directory)
  writeFileSync(resolve(directory, 'Fixture.vue'), source)
  return directory
}

function runScanner(directory: string, ...args: string[]) {
  return spawnSync(process.execPath, [scanner, '--scan-root', directory, ...args], {
    cwd: resolve(import.meta.dirname, '..'),
    encoding: 'utf8',
  })
}
