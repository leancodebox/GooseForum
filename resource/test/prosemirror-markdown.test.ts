import { describe, expect, test } from 'vitest'
import { Fragment } from 'prosemirror-model'
import { parseEditableVisualMarkdown, parseVisualMarkdown, serializeVisualMarkdown, visualMarkdownSchema } from '../src/runtime/prosemirror-markdown'

function roundTrip(markdown: string) {
  return serializeVisualMarkdown(parseVisualMarkdown(markdown))
}

describe('ProseMirror Markdown conversion', () => {
  test('preserves the supported document structures', () => {
    const markdown = [
      '# Heading',
      '',
      'A **bold**, *italic*, ~~deleted~~ and [linked](https://example.com) paragraph.',
      '',
      '> Quoted text',
      '',
      '- First',
      '- Second',
      '',
      '```go',
      'fmt.Println("hello")',
      '```',
      '',
      '![image](https://example.com/image.png)',
    ].join('\n')

    expect(roundTrip(markdown)).toBe(markdown)
  })

  test('does not typographically rewrite source punctuation or bare URLs', () => {
    const markdown = `"quotes" -- https://example.com`

    expect(roundTrip(markdown)).toBe(markdown)
  })

  test('supports and normalizes the remaining CommonMark block forms', () => {
    const markdown = [
      'Setext heading',
      '==============',
      '',
      '---',
      '',
      '1. First',
      '2. Second',
      '',
      '    indented code',
      '',
      'A line with a hard break.  ',
      'Next line.',
    ].join('\n')

    const result = roundTrip(markdown)
    expect(result).toContain('# Setext heading')
    expect(result).toContain('---')
    expect(result).toContain('1. First')
    expect(result).toContain('2. Second')
    expect(result).toContain('indented code')
    expect(result).toContain('A line with a hard break.\\\nNext line.')
  })

  test('supports CommonMark references, autolinks, escapes and entities', () => {
    const markdown = [
      '[reference][target] and <https://example.com>.',
      '',
      '\\*literal asterisk\\* &amp;',
      '',
      '[target]: https://example.com "Title"',
    ].join('\n')

    const result = roundTrip(markdown)
    expect(result).toContain('[reference](https://example.com "Title")')
    expect(result).toContain('<https://example.com>')
    expect(result).toContain('\\*literal asterisk\\* &')
  })

  test('round-trips Markdown tables', () => {
    const markdown = [
      '| Name | Description |',
      '| :--- | ---: |',
      '| GooseForum | A **modern** forum |',
      '| Escaped | A \\| B |',
    ].join('\n')

    expect(roundTrip(markdown)).toBe(markdown)
  })

  test('keeps a newly inserted empty table stable across remounts', () => {
    const markdown = [
      '|  |  |  |',
      '| --- | --- | --- |',
      '|  |  |  |',
      '|  |  |  |',
    ].join('\n')

    const firstPass = roundTrip(markdown)
    expect(roundTrip(firstPass)).toBe(firstPass)
  })

  test('does not serialize table cell line breaks as HTML', () => {
    const schema = visualMarkdownSchema
    const header = schema.nodes.table_header.create(null, schema.text('Header'))
    const body = schema.nodes.table_cell.create(null, [schema.text('First'), schema.nodes.hard_break.create(), schema.text('Second')])
    const table = schema.nodes.table.create(null, [
      schema.nodes.table_row.create(null, [header]),
      schema.nodes.table_row.create(null, [body]),
    ])
    const markdown = serializeVisualMarkdown(schema.nodes.doc.create(null, [table]))

    expect(markdown).toContain('| First Second |')
    expect(markdown).not.toContain('<br>')
  })

  test('does not serialize editor-only boundary paragraphs', () => {
    for (const markdown of ['```\ncode\n```', '---', '| A |\n| --- |\n| B |']) {
      const doc = parseVisualMarkdown(markdown)
      const paragraph = visualMarkdownSchema.nodes.paragraph.create()
      const editableDoc = doc.copy(Fragment.from(paragraph).append(doc.content).append(Fragment.from(paragraph)))
      expect(serializeVisualMarkdown(editableDoc)).toBe(markdown)
    }
  })

  test('restores editor-only boundary paragraphs after a Markdown round trip', () => {
    for (const markdown of [
      '```\ncode\n```',
      '---',
      '| A |\n| --- |\n| B |',
      '| A |\n| --- |\n| B |\n\n---\n\n```\ncode\n```',
    ]) {
      const firstEditableDoc = parseEditableVisualMarkdown(markdown)
      const persistedMarkdown = serializeVisualMarkdown(firstEditableDoc)
      const restoredEditableDoc = parseEditableVisualMarkdown(persistedMarkdown)
      const restoredNodeNames = Array.from(
        { length: restoredEditableDoc.childCount },
        (_, index) => restoredEditableDoc.child(index).type.name,
      )
      const boundaryNames = new Set(['code_block', 'horizontal_rule', 'table'])

      expect(restoredEditableDoc.firstChild?.type.name).toBe('paragraph')
      expect(restoredEditableDoc.lastChild?.type.name).toBe('paragraph')
      expect(restoredNodeNames.every((name, index) => (
        index === 0 || !boundaryNames.has(name) || !boundaryNames.has(restoredNodeNames[index - 1])
      ))).toBe(true)
      expect(persistedMarkdown).toBe(markdown)
    }
  })

  test('keeps adjacent tables separate when an editable paragraph is inserted between them', () => {
    const firstTable = parseVisualMarkdown('| A |\n| --- |\n| B |').firstChild!
    const secondTable = parseVisualMarkdown('| C |\n| --- |\n| D |').firstChild!
    const paragraph = visualMarkdownSchema.nodes.paragraph.create()
    const doc = visualMarkdownSchema.nodes.doc.create(null, [firstTable, paragraph, secondTable])

    const markdown = serializeVisualMarkdown(doc)
    expect(markdown).toContain('| B |\n\n| C |')
    expect(parseVisualMarkdown(markdown).childCount).toBe(2)
  })

  test('keeps surrounding paragraphs separate from tables', () => {
    const markdown = [
      'Text before the table.',
      '',
      '| Header |',
      '| --- |',
      '| Value |',
      '',
      'Text after the table.',
    ].join('\n')

    expect(roundTrip(markdown)).toBe(markdown)
  })
})
