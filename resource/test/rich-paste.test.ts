import { describe, expect, test } from 'vitest'
import { htmlToMarkdown } from '../src/runtime/rich-paste'

describe('htmlToMarkdown', () => {
  test('converts common rich text formatting to markdown', () => {
    const markdown = htmlToMarkdown('<h2>Hello</h2><p><strong>Bold</strong> and <a href="https://example.com">link</a></p><ul><li>One</li><li>Two</li></ul>')

    expect(markdown).toContain('## Hello')
    expect(markdown).toContain('**Bold**')
    expect(markdown).toContain('[link](https://example.com)')
    expect(markdown).toMatch(/-\s+One/)
    expect(markdown).toMatch(/-\s+Two/)
  })
})
