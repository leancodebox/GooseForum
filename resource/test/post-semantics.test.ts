import { describe, expect, it } from 'vitest'
import { createSSRApp, h } from 'vue'
import { renderToString } from 'vue/server-renderer'
import { createI18n } from 'vue-i18n'
import PostHeader from '../src/site/components/PostHeader.vue'
import PostReplyReference from '../src/site/components/PostReplyReference.vue'
import { postURL } from '../src/runtime/post-url'
import type { PostPayload, ReplyTargetPayload } from '@gooseforum/client'

async function render(node: ReturnType<typeof h>) {
  const app = createSSRApp({ render: () => node })
  app.use(createI18n({ legacy: false, locale: 'en', missingWarn: false, fallbackWarn: false, messages: { en: {} } }))
  return renderToString(app)
}
const author = { id: 1, username: 'author', avatarUrl: '' }

describe('post source links', () => {
  it('renders an author relationship and a permalink usable from another post window', async () => {
    const post = { id: 42, postNo: 21, author, createdAt: '2026-09-09T12:00:00Z' } as PostPayload
    const html = await render(h(PostHeader, { post, first: false, permalink: postURL(7, post.postNo) }))
    expect(html).toContain('id="post-author-42" rel="author" href="/u/1"')
    expect(html).toContain('href="/p/post/7/21"')
    expect(html).toContain('datetime="2026-09-09T12:00:00Z"')
    expect(postURL(7, 1)).toBe('/p/post/7')
  })

  it('identifies the quoted source with both a visible link and blockquote citation', async () => {
    const target: ReplyTargetPayload = { id: 42, postNo: 21, author, renderedContent: '<p>Quoted reply</p>' }
    const html = await render(h(PostReplyReference, { topicId: 7, target }))
    expect(html).toContain('href="/p/post/7/21"')
    expect(html).toContain('<blockquote cite="/p/post/7/21"')
    expect(html).toContain('<p>Quoted reply</p>')
  })

  it('does not expose a source link or quoted body for unavailable replies', async () => {
    const target: ReplyTargetPayload = { id: 42, postNo: 21, author, unavailable: true, renderedContent: '<p>Hidden</p>' }
    const html = await render(h(PostReplyReference, { topicId: 7, target }))
    expect(html).not.toContain('/p/post/')
    expect(html).not.toContain('<blockquote')
    expect(html).not.toContain('<p>Hidden</p>')
  })
})
