import { describe, expect, test } from 'vitest'
import { hasUnsupportedVisualMarkdown, htmlToMarkdown } from '../src/runtime/rich-paste'

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

describe('hasUnsupportedVisualMarkdown', () => {
  test('detects structures that Turndown cannot preserve', () => {
    expect(hasUnsupportedVisualMarkdown('- [x] Done')).toBe(true)
    expect(hasUnsupportedVisualMarkdown('1. [ ] Pending')).toBe(true)
    expect(hasUnsupportedVisualMarkdown('| A | B |\n|---|---|\n| 1 | 2 |')).toBe(false)
  })

  test('allows common markdown blocks', () => {
    expect(hasUnsupportedVisualMarkdown('# Title\n\n> Quote\n\n```go\nfmt.Println()\n```')).toBe(false)
  })

  test('detects task lists nested in blockquotes', () => {
    expect(hasUnsupportedVisualMarkdown('> - [ ] quoted task')).toBe(true)
  })
})
