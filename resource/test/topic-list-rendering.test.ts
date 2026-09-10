import { describe, expect, it } from 'vitest'
import { createSSRApp, h } from 'vue'
import { renderToString } from 'vue/server-renderer'
import { createI18n } from 'vue-i18n'
import TopicList from '../src/site/components/TopicList.vue'
import TopicListFooter from '../src/site/components/TopicListFooter.vue'
import type { TopicPayload } from '@gooseforum/client'

const topic: TopicPayload = {
  id: 529, title: 'Go libraries', description: 'Tools for Go', url: '/p/post/529',
  author: { id: 1, username: 'author', avatarUrl: '' }, participants: [],
  categories: [{ id: 4, name: 'Coding', url: '/c/Coding/4', color: '' }],
  replyCount: 3, viewCount: 25, pinWeight: 0, processStatus: 0,
  activityText: '', lastUpdateTime: '2026-09-09T12:00:00Z',
}

async function render(node: ReturnType<typeof h>) {
  const app = createSSRApp({ render: () => node })
  app.use(createI18n({ legacy: false, locale: 'en', missingWarn: false, fallbackWarn: false, messages: { en: {} } }))
  return renderToString(app)
}

describe('topic discovery markup', () => {
  it('renders complete table rows with heading and real detail/category links', async () => {
    const html = await render(h(TopicList, { topics: [topic] }))
    expect(html).toMatch(/<table[^>]*>[\s\S]*<thead>[\s\S]*<th scope="col"/)
    expect(html).toMatch(/<tbody[^>]*>[\s\S]*<tr[^>]*>[\s\S]*<th scope="row"/)
    expect(html.match(/<td[ >]/g)).toHaveLength(4)
    expect(html).toMatch(/<h2[^>]*>[\s\S]*<a href="\/p\/post\/529" rel="bookmark"/)
    expect(html).toContain('href="/c/Coding/4"')
    expect(html).toContain('<time datetime="2026-09-09T12:00:00Z">')
  })

  it('keeps moderation actions outside anchors and empty content outside the table', async () => {
    const html = await render(h(TopicList, { topics: [topic] }, {
      activity: () => h('button', 'Restore'), empty: () => h('p', 'Empty state'),
    }))
    expect(html.replace(/<!--.*?-->/g, '')).toMatch(/<td[^>]*><button>Restore<\/button><\/td>/)
    expect(html).toMatch(/<\/table>[\s\S]*<p>Empty state<\/p>/)
  })

  it.each(['waterfall', 'pagination'] as const)('exposes a visible next-page anchor in %s mode', async (mode) => {
    const html = await render(h(TopicListFooter, {
      pagination: { page: 1, nextPage: 2, hasNext: true, nextUrl: '/?page=2' },
      mode, loadingMore: false, hasTopics: true, loadError: '',
    }))
    expect(html).toContain('href="/?page=2" rel="next"')
    expect(html).not.toContain('sr-only')
    expect(html).not.toContain('<button')
  })
})
