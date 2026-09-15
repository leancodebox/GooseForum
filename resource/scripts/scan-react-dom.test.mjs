import assert from 'node:assert/strict'
import { mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import { spawnSync } from 'node:child_process'
import test from 'node:test'

const scanner = resolve(import.meta.dirname, 'scan-react-dom.mjs')

test('fails on high-confidence TSX DOM and class errors', () => {
  withFixture(
    `export function Invalid() {
      return <div className="flex flex"><span className="block grid" />
        <img src="/avatar.webp" />
        <a href="https://example.com" target="_blank">Example</a>
      </div>
    }`,
    result => {
      assert.equal(result.status, 1)
      const report = JSON.parse(result.stdout)
      assert.deepEqual(
        report.findings.map(item => item.code).sort(),
        [
          'conflicting-class',
          'duplicate-class',
          'image-alt',
          'unsafe-blank-target',
        ],
      )
    },
  )
})

test('accepts a clean TSX fixture', () => {
  withFixture(
    `export function Valid() {
      return <div className="flex items-center">
        <img src="/avatar.webp" alt="" />
        <img {...{ src: '/dynamic.webp', alt: 'Dynamic' }} />
        <a href="https://example.com" target="_blank" rel="noreferrer">Example</a>
      </div>
    }`,
    result => {
      assert.equal(result.status, 0)
      assert.deepEqual(JSON.parse(result.stdout).findings, [])
    },
  )
})

test('scans JSX inside attribute expressions', () => {
  withFixture('export const view = <Card content={<img src="/missing-alt.webp" />} />', result => {
    assert.equal(result.status, 1)
    assert.equal(JSON.parse(result.stdout).findings[0].code, 'image-alt')
  })
})

test('fails on invalid TSX syntax', () => {
  withFixture('export const view = <div>', result => {
    assert.equal(result.status, 1)
    assert.ok(JSON.parse(result.stdout).findings.some(item => item.code === 'parse-error'))
  })
})

function withFixture(source, verify) {
  const directory = mkdtempSync(join(tmpdir(), 'gooseforum-react-dom-scan-'))
  try {
    writeFileSync(join(directory, 'fixture.tsx'), source)
    const result = spawnSync(
      process.execPath,
      [scanner, '--scan-root', directory, '--fail-on-errors', '--json'],
      { encoding: 'utf8' },
    )
    assert.equal(result.stderr, '')
    verify(result)
  } finally {
    rmSync(directory, { recursive: true, force: true })
  }
}
