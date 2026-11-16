import type { Page } from '@playwright/test'
import { expect, test } from './fixtures'

function post(postNo: number, height = 260) {
  return {
    id: 60 + postNo, topicId: 60, postNo, content: `Post ${postNo}`,
    renderedContent: `<p>Post ${postNo}</p><div style="height:${height}px">Reading content ${postNo}</div>`,
    processStatus: 0, isHidden: false, canModerate: false,
    author: { id: 7, username: 'alice', avatarUrl: '' },
    createdAt: `2026-10-${String(10 + postNo % 20).padStart(2, '0')}T08:00:00Z`, isOwnPost: false,
  }
}

function topicPage(posts = [post(1, 1400), post(2, 180), post(3, 500), post(4, 160), post(5, 400), post(6, 120)], maxPostNo = 6, hasAfter = false) {
  return {
    component: 'topic.detail',
    props: {
      topic: {
        id: 60, title: 'Reading position', description: '', url: '/p/test/60', topicStatus: 1, processStatus: 0,
        author: { id: 7, username: 'alice', avatarUrl: '' }, participants: [], categories: [],
        replyCount: maxPostNo - 1, maxPostNo, viewCount: 10, likeCount: 0,
        isLiked: false, isBookmarked: false, isWatched: false,
        createdAt: '2026-10-11T08:00:00Z', updatedAt: '2026-10-16T08:00:00Z',
      },
      postStream: {
        posts, replyTargets: [], beforePostNo: posts[0].postNo, afterPostNo: posts.at(-1)!.postNo,
        hasBefore: false, hasAfter, total: maxPostNo, maxPostNo,
      },
      hotTopics: [], permissions: { isOwnTopic: false, canPost: true, canModerateTopic: false },
    },
    layout: {
      site: { name: 'GooseForum', description: '', logo: '', favicon: '', brandType: 'default', brandText: '', brandImage: '' },
      viewer: { id: 7, username: 'alice', email: '', avatarUrl: '', isAuthenticated: true, canAccessAdmin: false, isModerator: false, requiresEmailVerification: false, adminPermissions: [] },
      header: [], sidebar: { activeKey: 'topics', categories: [] }, footer: { links: [], primary: [] },
      unread: { notifications: false, messages: false }, theme: { enabled: false, current: 'gf-light', themeColor: '#fbfdff' },
    },
    meta: { title: 'Reading position' }, url: '/p/test/60', version: '1.0',
  }
}

async function openEarlierWindow(page: Page) {
  const payload = topicPage(Array.from({ length: 10 }, (_, index) => post(index + 31)), 40)
  payload.props.postStream.hasBefore = true
  await page.route('**/__goose_page/**', route => route.fulfill({ json: payload }))
  await page.goto('/p/test/60?lang=en')
  await expect(page.getByRole('button', { name: 'Load earlier replies', exact: true })).toBeVisible()
}

async function startFrameTrace(page: Page, postNo: number) {
  return page.evaluate(postNo => {
    const anchor = document.querySelector<HTMLElement>(`[data-post-no="${postNo}"]`)!
    const state = { running: true, frames: [] as Array<{ top: number; floor: string }> }
    Object.assign(window, { earlierTrace: state })
    const snapshot = () => ({ top: anchor.getBoundingClientRect().top, floor: document.querySelector('[data-slot="mobile-topic-position"]')!.textContent!.trim() })
    const sample = () => { if (!state.running) return; state.frames.push(snapshot()); requestAnimationFrame(sample) }
    requestAnimationFrame(sample)
    return snapshot()
  }, postNo)
}

async function finishFrameTrace(page: Page) {
  await page.evaluate(() => new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))))
  return page.evaluate(() => {
    const state = (window as unknown as { earlierTrace: { running: boolean; frames: Array<{ top: number; floor: string }> } }).earlierTrace
    state.running = false
    return state.frames
  })
}

for (const continueReading of [false, true]) {
  test(`loading earlier replies preserves the visible post and floor${continueReading ? ' after scrolling during the request' : ''}`, async ({ page }) => {
    let release!: () => void
    const pending = new Promise<void>(resolve => { release = resolve })
    let requested = false
    await page.route('**/api/forum/posts/window**', async route => {
      requested = true
      await pending
      await route.fulfill({ json: { code: 0, result: {
        posts: Array.from({ length: 10 }, (_, index) => post(index + 21, 100)), replyTargets: [],
        beforePostNo: 21, afterPostNo: 30, hasBefore: true, hasAfter: true, maxPostNo: 40, total: 40,
      } } })
    })
    await openEarlierWindow(page)
    await page.getByRole('button', { name: 'Load earlier replies', exact: true }).click()
    await expect.poll(() => requested).toBe(true)
    if (continueReading) {
      await page.evaluate(() => window.scrollTo({ top: 600, behavior: 'instant' }))
      await expect.poll(() => page.locator('[data-slot="mobile-topic-position"]').textContent()).not.toContain('31 / 40')
    }
    const before = await startFrameTrace(page, 31)
    release()
    await expect(page.locator('[data-post-no="21"]')).toBeAttached()
    await expect.poll(() => page.locator('[data-post-no="31"]').evaluate(element => element.getBoundingClientRect().top)).toBeCloseTo(before.top, 0)
    const frames = await finishFrameTrace(page)
    expect(frames.length).toBeGreaterThan(0)
    expect(frames.every(frame => Math.abs(frame.top - before.top) <= 1)).toBe(true)
    expect(frames.every(frame => frame.floor === before.floor)).toBe(true)
  })
}

test('repeated earlier loads preserve position when the earlier-load button disappears', async ({ page }) => {
  const cursors: number[] = []
  await page.route('**/api/forum/posts/window**', route => {
    const cursor = Number(new URL(route.request().url()).searchParams.get('beforePostNo'))
    cursors.push(cursor)
    const first = cursor - 10
    return route.fulfill({ json: { code: 0, result: {
      posts: Array.from({ length: 10 }, (_, index) => post(first + index, 100)), replyTargets: [],
      beforePostNo: first, afterPostNo: cursor - 1, hasBefore: first > 1, hasAfter: true, maxPostNo: 40, total: 40,
    } } })
  })
  await openEarlierWindow(page)
  for (const cursor of [31, 21, 11]) {
    await page.evaluate(() => window.scrollTo({ top: 0, behavior: 'instant' }))
    await expect.poll(() => page.locator('[data-slot="mobile-topic-position"]').textContent()).toContain(`${cursor} / 40`)
    const before = await startFrameTrace(page, cursor)
    await page.getByRole('button', { name: 'Load earlier replies', exact: true }).click()
    await expect(page.locator(`[data-post-no="${cursor - 10}"]`)).toBeAttached()
    await expect.poll(() => page.locator(`[data-post-no="${cursor}"]`).evaluate(element => element.getBoundingClientRect().top)).toBeCloseTo(before.top, 0)
    const frames = await finishFrameTrace(page)
    expect(frames.every(frame => Math.abs(frame.top - before.top) <= 1)).toBe(true)
    expect(frames.every(frame => frame.floor === before.floor)).toBe(true)
  }
  expect(cursors).toEqual([31, 21, 11])
  await expect(page.getByRole('button', { name: 'Load earlier replies', exact: true })).toHaveCount(0)
})
