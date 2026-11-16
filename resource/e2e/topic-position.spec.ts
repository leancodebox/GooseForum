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
      hotTopics: [], permissions: { isOwnTopic: false, canPost: false, canModerateTopic: false },
    },
    layout: {
      site: { name: 'GooseForum', description: '', logo: '', favicon: '', brandType: 'default', brandText: '', brandImage: '' },
      viewer: { id: 0, username: '', email: '', avatarUrl: '', isAuthenticated: false, canAccessAdmin: false, isModerator: false, requiresEmailVerification: false, adminPermissions: [] },
      header: [], sidebar: { activeKey: 'topics', categories: [] }, footer: { links: [], primary: [] },
      unread: { notifications: false, messages: false }, theme: { enabled: false, current: 'gf-light', themeColor: '#fbfdff' },
    },
    meta: { title: 'Reading position' }, url: '/p/test/60', version: '1.0',
  }
}

async function openTopic(page: Page, payload = topicPage()) {
  await page.route('**/__goose_page/**', route => route.fulfill({ json: payload }))
  await page.goto('/p/test/60?lang=en')
  await expect(page.getByRole('slider', { name: 'Reply position' })).toBeVisible()
}

test.beforeEach(async ({ page }, testInfo) => {
  test.skip(testInfo.project.name === 'mobile-chromium', 'The topic position rail is only displayed at desktop widths.')
  await page.setViewportSize({ width: 1440, height: 900 })
})

test('the position marker moves continuously while reading inside one long post', async ({ page }) => {
  await openTopic(page)
  const rail = page.getByTestId('post-position-rail')
  const slider = page.getByRole('slider', { name: 'Reply position' })
  await page.evaluate(() => window.scrollTo({ top: 200, behavior: 'instant' }))
  await expect(slider).toHaveAttribute('aria-valuenow', '1')
  await expect.poll(async () => Number(await rail.getAttribute('data-progress'))).toBeGreaterThan(0)
  const previous = Number(await rail.getAttribute('data-progress'))
  const previousTop = await slider.evaluate(element => element.getBoundingClientRect().top - element.closest('[data-testid="post-position-rail"]')!.getBoundingClientRect().top)
  await page.evaluate(() => window.scrollTo({ top: 650, behavior: 'instant' }))
  await expect.poll(async () => Number(await rail.getAttribute('data-progress'))).toBeGreaterThan(previous)
  await expect(slider).toHaveAttribute('aria-valuenow', '1')
  await expect.poll(() => slider.evaluate(element => element.getBoundingClientRect().top - element.closest('[data-testid="post-position-rail"]')!.getBoundingClientRect().top)).toBeGreaterThan(previousTop)
})

test('jumping to the latest reply does not rewind the position and reaches the last short post', async ({ page }) => {
  await openTopic(page)
  const slider = page.getByRole('slider', { name: 'Reply position' })
  const rail = page.getByTestId('post-position-rail')
  await slider.evaluate(element => {
    const values: string[] = []
    const observer = new MutationObserver(() => values.push(element.getAttribute('aria-valuenow') || ''))
    observer.observe(element, { attributes: true, attributeFilter: ['aria-valuenow'] })
    Object.assign(window, { positionTrace: { values, observer } })
  })
  await page.getByRole('button', { name: /^Latest reply:/ }).click()
  await expect(slider).toHaveAttribute('aria-valuenow', '6')
  await expect(rail).toHaveAttribute('data-state', 'reading')
  await expect(page.locator('[data-post-no="6"]')).toBeInViewport()
  const values = await page.evaluate(() => {
    const trace = (window as unknown as { positionTrace: { values: string[]; observer: MutationObserver } }).positionTrace
    trace.observer.disconnect()
    return trace.values
  })
  expect(values).toContain('6')
  expect(values.slice(values.indexOf('6')).every(value => value === '6')).toBe(true)
  await page.getByRole('button', { name: /^Earliest content:/ }).click()
  await expect(rail).toHaveAttribute('data-state', 'reading')
  await expect(slider).toHaveAttribute('aria-valuenow', '1')
  await expect(page.locator('[data-post-no="1"]')).toBeInViewport()
})

test('dragging previews a floor without scrolling until the pointer is released', async ({ page }) => {
  await openTopic(page)
  const rail = page.getByTestId('post-position-rail')
  const slider = page.getByRole('slider', { name: 'Reply position' })
  const thumb = (await slider.boundingBox())!
  const startY = await page.evaluate(() => scrollY)
  await page.mouse.move(thumb.x + thumb.width / 2, thumb.y + thumb.height / 2)
  await page.mouse.down()
  await page.mouse.move(thumb.x + thumb.width / 2, thumb.y + thumb.height / 2 + 85, { steps: 8 })
  await expect(rail).toHaveAttribute('data-state', 'dragging')
  const target = Number(await slider.getAttribute('aria-valuenow'))
  expect(target).toBeGreaterThan(1)
  expect(target).toBeLessThan(6)
  expect(await page.evaluate(() => scrollY)).toBe(startY)
  await page.mouse.up()
  await expect(rail).toHaveAttribute('data-state', 'reading')
  await expect(slider).toHaveAttribute('aria-valuenow', String(target))
  await expect(page.locator(`[data-post-no="${target}"]`)).toBeInViewport()
})

test('an unloaded jump keeps the current content while waiting and resolves a missing floor', async ({ page }) => {
  let release: (() => void) | undefined
  const pending = new Promise<void>(resolve => { release = resolve })
  const requests: number[] = []
  await page.route('**/api/forum/posts/window**', async route => {
    const anchor = Number(new URL(route.request().url()).searchParams.get('anchorPostNo'))
    requests.push(anchor)
    await pending
    await route.fulfill({ json: { code: 0, result: {
      posts: [post(78, 500), post(79, 120)], replyTargets: [], beforePostNo: 78, afterPostNo: 79,
      hasBefore: true, hasAfter: false, total: 80, maxPostNo: 80,
    } } })
  })
  await openTopic(page, topicPage([post(1, 1400), post(2, 400)], 80, true))
  const rail = page.getByTestId('post-position-rail')
  const slider = page.getByRole('slider', { name: 'Reply position' })
  await page.getByRole('button', { name: /^Latest reply:/ }).click()
  await expect(rail).toHaveAttribute('data-state', 'jumping')
  await expect.poll(() => requests).toEqual([80])
  await expect(page.locator('[data-post-no="1"]')).toBeVisible()
  release!()
  await expect(slider).toHaveAttribute('aria-valuenow', '79')
  await expect(rail).toHaveAttribute('data-state', 'reading')
  await expect(page.locator('[data-post-no="79"]')).toBeInViewport()
  expect(requests).toEqual([80])
})

test('a wheel gesture interrupts a short animated jump and returns control to reading', async ({ page }) => {
  await openTopic(page, topicPage([post(1, 700), post(2, 450), post(3, 500), post(4, 400), post(5, 400), post(6, 200)]))
  const rail = page.getByTestId('post-position-rail')
  const slider = page.getByRole('slider', { name: 'Reply position' })
  const destination = await page.locator('[data-post-no="2"]').evaluate(element => element.getBoundingClientRect().top + scrollY - Math.min(innerHeight * 0.38, 340))
  await slider.focus()
  await page.keyboard.press('ArrowDown')
  await expect(rail).toHaveAttribute('data-state', 'jumping')
  await page.mouse.move(600, 450)
  await page.mouse.wheel(0, -250)
  await expect(rail).toHaveAttribute('data-state', 'reading')
  await expect(slider).toHaveAttribute('aria-valuenow', '1')
  // Observe beyond the maximum jump duration to catch an animation that resumes after cancellation.
  await page.waitForTimeout(450)
  expect(await page.evaluate(() => scrollY)).toBeLessThan(destination - 50)
  await expect(slider).toHaveAttribute('aria-valuenow', '1')
})

test('a wheel gesture cancels an unloaded jump and ignores its late response', async ({ page }) => {
  let release: (() => void) | undefined
  const pending = new Promise<void>(resolve => { release = resolve })
  await page.route('**/api/forum/posts/window**', async route => {
    await pending
    await route.fulfill({ json: { code: 0, result: {
      posts: [post(79, 500), post(80, 120)], replyTargets: [], beforePostNo: 79, afterPostNo: 80,
      hasBefore: true, hasAfter: false, total: 80, maxPostNo: 80,
    } } })
  })
  await openTopic(page, topicPage([post(1, 1400), post(2, 400)], 80, true))
  const rail = page.getByTestId('post-position-rail')
  const request = page.waitForRequest('**/api/forum/posts/window**')
  await page.getByRole('button', { name: /^Latest reply:/ }).click()
  await request
  await expect(rail).toHaveAttribute('data-state', 'jumping')
  await page.mouse.move(600, 450)
  await page.mouse.wheel(0, 120)
  await expect(rail).toHaveAttribute('data-state', 'reading')
  const response = page.waitForResponse('**/api/forum/posts/window**')
  release!()
  await (await response).finished()
  // Allow the fulfilled request and the next React render to run before checking the retained window.
  await page.evaluate(() => new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))))
  await expect(page.locator('[data-post-no="1"]')).toBeVisible()
  await expect(page.locator('[data-post-no="80"]')).toHaveCount(0)
  await expect(page.getByRole('slider', { name: 'Reply position' })).toHaveAttribute('aria-valuenow', '1')
  await expect(rail).toHaveAttribute('data-state', 'reading')
})

test('content growth and viewport resizing refresh the reading progress without another scroll', async ({ page }) => {
  await openTopic(page)
  const rail = page.getByTestId('post-position-rail')
  await page.evaluate(() => window.scrollTo({ top: 650, behavior: 'instant' }))
  await expect.poll(async () => Number(await rail.getAttribute('data-progress'))).toBeGreaterThan(0)
  const beforeGrowth = Number(await rail.getAttribute('data-progress'))
  await page.locator('[data-post-no="1"] .gf-prose div').evaluate(element => { (element as HTMLElement).style.height = '2200px' })
  await expect.poll(async () => Number(await rail.getAttribute('data-progress'))).toBeLessThan(beforeGrowth)
  const beforeResize = Number(await rail.getAttribute('data-progress'))
  await page.setViewportSize({ width: 1440, height: 600 })
  await expect.poll(async () => Number(await rail.getAttribute('data-progress'))).toBeLessThan(beforeResize)
  await expect(page.getByRole('slider', { name: 'Reply position' })).toHaveAttribute('aria-valuenow', '1')
})

test('reduced-motion keyboard navigation selects whole floors and supports Home and End', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await openTopic(page)
  const slider = page.getByRole('slider', { name: 'Reply position' })
  const rail = page.getByTestId('post-position-rail')
  await slider.focus()
  await page.keyboard.press('ArrowDown')
  await expect(slider).toHaveAttribute('aria-valuenow', '2')
  await expect(rail).toHaveAttribute('data-state', 'reading')
  await page.keyboard.press('ArrowDown')
  await expect(slider).toHaveAttribute('aria-valuenow', '3')
  await expect(rail).toHaveAttribute('data-state', 'reading')
  await page.keyboard.press('End')
  await expect(slider).toHaveAttribute('aria-valuenow', '6')
  await expect(rail).toHaveAttribute('data-state', 'reading')
  await page.keyboard.press('Home')
  await expect(slider).toHaveAttribute('aria-valuenow', '1')
  await expect(rail).toHaveAttribute('data-state', 'reading')
  await page.keyboard.press('PageDown')
  await expect(slider).toHaveAttribute('aria-valuenow', '6')
  await expect(rail).toHaveAttribute('data-state', 'reading')
})

test('leaving a pending jump and going Back restores a usable cached topic without the late window', async ({ page }) => {
  const topic = topicPage([post(1, 1400), post(2, 400)], 80, true)
  const home = {
    ...topic, component: 'home.index', url: '/', meta: { title: 'GooseForum' },
    props: {
      sort: 'latest', tabs: [{ key: 'latest', label: 'Latest', url: '/', active: true }], topics: [],
      pagination: { page: 1, nextPage: 2, hasNext: false, nextUrl: '' }, announcement: { enabled: false, html: '' },
    },
  }
  let release: (() => void) | undefined
  const pending = new Promise<void>(resolve => { release = resolve })
  await page.route('**/__goose_page/**', route => route.fulfill({
    json: new URL(route.request().url()).pathname.includes('/p/') ? topic : home,
  }))
  await page.route('**/api/forum/posts/window**', async route => {
    await pending
    await route.fulfill({ json: { code: 0, result: {
      posts: [post(79, 500), post(80, 120)], replyTargets: [], beforePostNo: 79, afterPostNo: 80,
      hasBefore: true, hasAfter: false, total: 80, maxPostNo: 80,
    } } })
  })
  await page.goto('/p/test/60?lang=en')
  const request = page.waitForRequest('**/api/forum/posts/window**')
  await page.getByRole('button', { name: /^Latest reply:/ }).click()
  await request
  await expect(page.getByTestId('post-position-rail')).toHaveAttribute('data-state', 'jumping')
  await page.getByRole('link', { name: 'Topics', exact: true }).first().click()
  await expect(page.getByText('No topics yet')).toBeVisible()
  const response = page.waitForResponse('**/api/forum/posts/window**')
  release!()
  await (await response).finished()
  await page.goBack()
  await expect(page.getByTestId('post-position-rail')).toBeVisible()
  await expect(page.getByTestId('post-position-rail')).toHaveAttribute('data-state', 'reading')
  await expect(page.getByRole('slider', { name: 'Reply position' })).toHaveAttribute('aria-valuenow', '1')
  await expect(page.locator('[data-post-no="1"]')).toBeVisible()
  await expect(page.locator('[data-post-no="80"]')).toHaveCount(0)
  await expect(page.getByRole('button', { name: /^Latest reply:/ })).toBeEnabled()
})

test('a cancelled desktop drag uses the shifted sticky rail bounds on the next tap', async ({ page }) => {
  await openTopic(page)
  const rail = page.getByTestId('post-position-rail')
  const slider = page.getByRole('slider', { name: 'Reply position' })
  const root = rail.locator('[data-orientation="vertical"]').first()
  const before = (await root.boundingBox())!
  const thumb = (await slider.boundingBox())!
  const originalY = await page.evaluate(() => scrollY)
  const touch = await page.context().newCDPSession(page)
  await touch.send('Emulation.setTouchEmulationEnabled', { enabled: true, maxTouchPoints: 1 })
  const point = { x: thumb.x + thumb.width / 2, y: thumb.y + thumb.height / 2 }
  await touch.send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: [point] })
  await touch.send('Input.dispatchTouchEvent', { type: 'touchMove', touchPoints: [{ x: point.x, y: point.y + 40 }] })
  await expect(rail).toHaveAttribute('data-state', 'dragging')
  await expect(slider).not.toHaveAttribute('aria-valuenow', '1')
  await touch.send('Input.dispatchTouchEvent', { type: 'touchCancel', touchPoints: [] })
  await expect(rail).toHaveAttribute('data-state', 'reading')
  await expect(slider).toHaveAttribute('aria-valuenow', '1')
  expect(await page.evaluate(() => scrollY)).toBe(originalY)

  await page.evaluate(() => window.scrollTo({ top: 650, behavior: 'instant' }))
  await expect.poll(async () => (await root.boundingBox())!.y).toBeLessThan(before.y - 50)
  const shifted = (await root.boundingBox())!
  const destination = { x: shifted.x + shifted.width / 2, y: shifted.y + shifted.height * 0.75 }
  await touch.send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: [destination] })
  await expect(slider).toHaveAttribute('aria-valuenow', '5')
  await touch.send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] })
  await touch.detach()
  await expect(rail).toHaveAttribute('data-state', 'reading')
  await expect(slider).toHaveAttribute('aria-valuenow', '5')
  await expect(page.locator('[data-post-no="5"]')).toBeInViewport()
})
