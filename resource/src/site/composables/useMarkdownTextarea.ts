import { nextTick, type Ref } from 'vue'
import { fencedCodeBlock, formatMarkdownLines, prefixMarkdownBlock, replaceMarkdownSelectionWithBlock, type MarkdownBlockType } from '@/runtime/markdown-editing'

export type MarkdownToolbarAction = 'bold' | 'italic' | 'strike' | 'inlineCode' | 'link' | 'quote' | 'code' | 'bulletList' | 'orderedList'

export interface MarkdownToolbarPlaceholders {
  bold: string
  italic: string
  strike: string
  link: string
  quote: string
  listItem: string
}

export function useMarkdownTextarea(options: {
  content: Ref<string>
  editor: Ref<HTMLTextAreaElement | null>
}) {
  const { content, editor } = options

  function restoreSelection(start: number, end = start) {
    void nextTick(() => {
      editor.value?.focus()
      editor.value?.setSelectionRange(start, end)
    })
  }

  function insertInline(before: string, after = '', placeholder = '') {
    const element = editor.value
    if (!element) return
    const start = element.selectionStart
    const end = element.selectionEnd
    const selected = content.value.slice(start, end) || placeholder
    content.value = `${content.value.slice(0, start)}${before}${selected}${after}${content.value.slice(end)}`
    restoreSelection(start + before.length, start + before.length + selected.length)
  }

  function insertBlock(text: string) {
    const element = editor.value
    if (!element) {
      content.value = content.value ? `${content.value}\n${text}` : text
      return
    }

    const result = replaceMarkdownSelectionWithBlock(content.value, element.selectionStart, element.selectionEnd, text)
    content.value = result.value
    restoreSelection(result.selectionEnd)
  }

  function selectedText(placeholder: string) {
    const element = editor.value
    if (!element) return placeholder
    return content.value.slice(element.selectionStart, element.selectionEnd) || placeholder
  }

  function insertPrefixedBlock(prefix: string, placeholder: string) {
    insertBlock(prefixMarkdownBlock(selectedText(placeholder), prefix))
  }

  function applyAction(action: MarkdownToolbarAction, placeholders: MarkdownToolbarPlaceholders) {
    if (action === 'bold') insertInline('**', '**', placeholders.bold)
    else if (action === 'italic') insertInline('*', '*', placeholders.italic)
    else if (action === 'strike') insertInline('~~', '~~', placeholders.strike)
    else if (action === 'inlineCode') insertInline('`', '`', 'code')
    else if (action === 'link') insertInline('[', '](https://)', placeholders.link)
    else if (action === 'quote') insertPrefixedBlock('> ', placeholders.quote)
    else if (action === 'code') insertBlock(fencedCodeBlock(selectedText('code')))
    else if (action === 'bulletList') insertPrefixedBlock('- ', placeholders.listItem)
    else insertPrefixedBlock('1. ', placeholders.listItem)
  }

  function setBlock(block: MarkdownBlockType) {
    const element = editor.value
    if (!element) return
    if (block === 'code_block') {
      insertBlock(fencedCodeBlock(selectedText('code')))
      return
    }

    const result = formatMarkdownLines(content.value, element.selectionStart, element.selectionEnd, block)
    content.value = result.value
    restoreSelection(result.selectionStart, result.selectionEnd)
  }

  return {
    applyAction,
    insertBlock,
    insertInline,
    selectedText,
    setBlock,
  }
}
