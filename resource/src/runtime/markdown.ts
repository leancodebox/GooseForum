import MarkdownIt from 'markdown-it'
import anchor from 'markdown-it-anchor'
import taskLists from 'markdown-it-task-lists'

export const markdownPreview = new MarkdownIt({
  html: false,
  linkify: true,
  typographer: true,
  breaks: false,
})
  .use(anchor, {
    slugify: (value: string) => value.trim().toLowerCase().replace(/\s+/g, '-'),
  })
  .use(taskLists, { enabled: true })

markdownPreview.renderer.rules.s_open = () => '<del>'
markdownPreview.renderer.rules.s_close = () => '</del>'

export function renderMarkdownPreview(source: string) {
  return markdownPreview.render(source || '')
}
