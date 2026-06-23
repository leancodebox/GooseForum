import TurndownService from 'turndown'

const turndown = new TurndownService({
  headingStyle: 'atx',
  bulletListMarker: '-',
  codeBlockStyle: 'fenced',
})

export function htmlToMarkdown(html: string) {
  return turndown.turndown(html).trim()
}

export function markdownFromClipboard(data: DataTransfer | null) {
  const html = data?.getData('text/html') || ''
  if (!html.trim()) return ''
  return htmlToMarkdown(html)
}
