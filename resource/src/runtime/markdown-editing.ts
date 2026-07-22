export type MarkdownBlockType = 'paragraph' | 'code_block' | `heading_${1 | 2 | 3 | 4 | 5 | 6}`

export interface MarkdownEditResult {
  value: string
  selectionStart: number
  selectionEnd: number
}

function blockBoundary(left: string, right: string) {
  if (!left || !right) return ''
  const trailingNewlines = left.match(/\n*$/)?.[0].length || 0
  const leadingNewlines = right.match(/^\n*/)?.[0].length || 0
  return '\n'.repeat(Math.max(0, 2 - trailingNewlines - leadingNewlines))
}

export function replaceMarkdownSelectionWithBlock(source: string, start: number, end: number, block: string): MarkdownEditResult {
  const before = source.slice(0, start)
  const after = source.slice(end)
  const prefix = blockBoundary(before, block)
  const suffix = blockBoundary(block, after)
  const blockStart = before.length + prefix.length
  return {
    value: `${before}${prefix}${block}${suffix}${after}`,
    selectionStart: blockStart,
    selectionEnd: blockStart + block.length,
  }
}

export function prefixMarkdownBlock(text: string, prefix: string) {
  return text.split('\n').map(line => `${prefix}${line}`).join('\n')
}

export function fencedCodeBlock(text: string) {
  return `\`\`\`\n${text}\n\`\`\``
}

export function createMarkdownTable(rowCount: number, columnCount: number) {
  const rows = Math.max(1, rowCount)
  const columns = Math.max(1, columnCount)
  const row = `|${Array.from({ length: columns }, () => '  |').join('')}`
  const separator = `|${Array.from({ length: columns }, () => ' --- |').join('')}`
  return [row, separator, ...Array.from({ length: rows - 1 }, () => row)].join('\n')
}

export function formatMarkdownLines(source: string, start: number, end: number, block: Exclude<MarkdownBlockType, 'code_block'>): MarkdownEditResult {
  const effectiveEnd = end > start && source[end - 1] === '\n' ? end - 1 : end
  const lineStart = source.lastIndexOf('\n', start - 1) + 1
  const nextLineBreak = source.indexOf('\n', effectiveEnd)
  const lineEnd = nextLineBreak === -1 ? source.length : nextLineBreak
  const prefix = block === 'paragraph' ? '' : `${'#'.repeat(Number(block.slice(-1)))} `
  const replacement = source.slice(lineStart, lineEnd)
    .split('\n')
    .map(line => line.trim() ? `${prefix}${line.replace(/^#{1,6}\s+/, '')}` : line)
    .join('\n')
  return {
    value: `${source.slice(0, lineStart)}${replacement}${source.slice(lineEnd)}`,
    selectionStart: lineStart,
    selectionEnd: lineStart + replacement.length,
  }
}
