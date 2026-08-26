import { describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { useMarkdownTextarea } from '@/site/composables/useMarkdownTextarea'

const placeholders = {
  bold: 'bold text',
  italic: 'italic text',
  strike: 'struck text',
  link: 'link text',
  quote: 'quoted text',
  listItem: 'list item',
}

function setup(value: string, selectionStart = 0, selectionEnd = selectionStart) {
  const content = ref(value)
  const textarea = {
    focus: vi.fn(),
    selectionStart,
    selectionEnd,
    setSelectionRange: vi.fn(),
  } as unknown as HTMLTextAreaElement
  const editor = ref<HTMLTextAreaElement | null>(textarea)
  return { content, markdown: useMarkdownTextarea({ content, editor }) }
}

describe('useMarkdownTextarea', () => {
  it('formats the current selection with an inline action', () => {
    const { content, markdown } = setup('hello world', 6, 11)

    markdown.applyAction('bold', placeholders)

    expect(content.value).toBe('hello **world**')
  })

  it('uses the shared placeholder when the selection is empty', () => {
    const { content, markdown } = setup('', 0)

    markdown.applyAction('link', placeholders)

    expect(content.value).toBe('[link text](https://)')
  })

  it('formats multiline selections as quote blocks', () => {
    const { content, markdown } = setup('first\nsecond', 0, 12)

    markdown.applyAction('quote', placeholders)

    expect(content.value).toBe('> first\n> second')
  })
})
