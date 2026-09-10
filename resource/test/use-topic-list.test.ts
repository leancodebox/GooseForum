import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createRenderer, defineComponent, nextTick, reactive } from 'vue'
import { useTopicList } from '../src/site/composables/useTopicList'
import { fetchPage } from '../src/runtime/router'
import type { HomeProps, PagePayload, TopicPayload } from '@gooseforum/client'

vi.mock('../src/runtime/router', () => ({ fetchPage: vi.fn() }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))

// Run the composable's real watchers/lifecycle without requiring a browser DOM.
const renderer = createRenderer<object, object>({
  patchProp() {}, insert() {}, remove() {}, setText() {}, setElementText() {},
  createElement: () => ({}), createText: () => ({}), createComment: () => ({}),
  parentNode: () => null, nextSibling: () => null,
})
const cleanup: (() => void)[] = []
const topic = (id: number, unseen = false) => ({ id, unseen }) as TopicPayload
const pagination = (page = 1) => ({ page, nextPage: page + 1, hasNext: true, nextUrl: `/?page=${page + 1}` })
const response = (ids: number[], page = 2) => ({ props: { topics: ids.map((id) => topic(id)), pagination: pagination(page) } }) as PagePayload<HomeProps>

function mountList() {
  const page = reactive({ props: { topics: [topic(1)], pagination: pagination() }, pageUrl: '/' })
  let list!: ReturnType<typeof useTopicList>
  const app = renderer.createApp(defineComponent({ setup() { list = useTopicList(page); return () => null } }))
  app.mount({})
  cleanup.push(() => app.unmount())
  return { page, list }
}

beforeEach(() => {
  vi.stubGlobal('window', { location: { origin: 'http://localhost' }, localStorage: { getItem: () => null, setItem() {} } })
  vi.mocked(fetchPage).mockReset()
})
afterEach(() => { cleanup.splice(0).forEach((fn) => fn()); vi.unstubAllGlobals() })

describe('topic list loading', () => {
  it('merges unique topics and preserves loaded rows when unread flags refresh', async () => {
    const { page, list } = mountList()
    vi.mocked(fetchPage).mockResolvedValue(response([1, 2, 2]))
    await list.loadMore()
    expect(list.topics.value.map((item) => item.id)).toEqual([1, 2])
    expect(list.pagination.value.page).toBe(2)
    page.props.topics = [topic(1, true)]
    await nextTick()
    expect(list.topics.value.map((item) => [item.id, item.unseen])).toEqual([[1, true], [2, false]])
  })

  it('ignores an old response after navigating to another list', async () => {
    const { page, list } = mountList()
    let resolve!: (value: PagePayload<HomeProps>) => void
    vi.mocked(fetchPage).mockReturnValue(new Promise((done) => { resolve = done }))
    const pending = list.loadMore()
    page.props = { topics: [topic(8)], pagination: pagination() }
    page.pageUrl = '/c/other/8'
    await nextTick()
    resolve(response([2]))
    await pending
    expect(list.topics.value.map((item) => item.id)).toEqual([8])
    expect(list.pagination.value.page).toBe(1)
    expect(list.loadingMore.value).toBe(false)
  })

  it('does not append a pending waterfall response after switching to pagination', async () => {
    const { list } = mountList()
    let resolve!: (value: PagePayload<HomeProps>) => void
    vi.mocked(fetchPage).mockReturnValue(new Promise((done) => { resolve = done }))
    const pending = list.loadMore()
    list.setListMode('pagination')
    await nextTick()
    resolve(response([2]))
    await pending
    expect(list.topics.value.map((item) => item.id)).toEqual([1])
    expect(list.pagination.value.page).toBe(1)
  })

  it('allows retry after a failed request and suppresses concurrent loads', async () => {
    const { list } = mountList()
    vi.mocked(fetchPage).mockRejectedValueOnce(new Error('offline'))
    await Promise.all([list.loadMore(), list.loadMore()])
    expect(fetchPage).toHaveBeenCalledTimes(1)
    expect(list.loadError.value).toBe('offline')
    vi.mocked(fetchPage).mockResolvedValueOnce(response([2]))
    await list.loadMore()
    expect(list.loadError.value).toBe('')
    expect(list.topics.value).toHaveLength(2)
  })
})
