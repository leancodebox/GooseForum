import { readdirSync, readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, test } from 'vitest'
import { renderMarkdownPreview } from '../src/runtime/markdown'

type MarkdownCompatCase = {
  name: string
  markdown: string
  contains: string[]
  notContains?: string[]
}

const fixturesRoot = fileURLToPath(new URL('../../testdata/markdown-compat/', import.meta.url))
const fixtures = readdirSync(fixturesRoot, { withFileTypes: true })
  .filter((entry) => entry.isFile() && entry.name.endsWith('.md'))
  .map((entry) => {
    const name = entry.name.replace(/\.md$/, '')
    const fixture = JSON.parse(readFileSync(`${fixturesRoot}/${name}.json`, 'utf8')) as Omit<MarkdownCompatCase, 'name' | 'markdown'>
    return {
      name,
      markdown: readFileSync(`${fixturesRoot}/${entry.name}`, 'utf8'),
      ...fixture,
    }
  })
  .sort((a, b) => a.name.localeCompare(b.name))

describe('markdown compatibility fixtures', () => {
  for (const fixture of fixtures) {
    test(fixture.name, () => {
      const html = renderMarkdownPreview(fixture.markdown)

      for (const expected of fixture.contains) {
        expect(html).toContain(expected)
      }
      for (const unexpected of fixture.notContains || []) {
        expect(html).not.toContain(unexpected)
      }
    })
  }
})
