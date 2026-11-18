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

async function openTopic(page: Page, payload = topicPage()) {
  await page.route('**/__goose_page/**', route => route.fulfill({ json: payload }))
  await page.goto('/p/test/60?lang=en')
  await expect(page.getByRole('button', { name: 'Reply position', exact: true })).toBeVisible()
}

test.beforeEach(async ({ page }) => {
  // Exercise the narrow layout with both mouse and touch-enabled browser projects.
  await page.setViewportSize({ width: 390, height: 844 })
})

test('the mobile progress control fits a short screen and opens a dismissible timeline', async ({ page }) => {
  await page.setViewportSize({ width: 375, height: 667 })
  await openTopic(page)
  const trigger = page.getByRole('button', { name: 'Reply position', exact: true })
  await expect(trigger).toContainText('1 / 6')
  await trigger.click()
  const panel = page.getByRole('dialog', { name: 'Reply position' })
  await expect(panel).toBeVisible()
  await expect(panel.getByRole('slider', { name: 'Reply position' })).toHaveAttribute('aria-valuenow', '1')
  await expect(panel.getByRole('button', { name: /^Earliest content:/ })).toBeVisible()
  await expect(panel.getByRole('button', { name: /^Latest reply:/ })).toBeVisible()
  expect(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth)).toBe(false)
  const bounds = (await panel.boundingBox())!
  expect(bounds.y).toBeGreaterThan(20)
  const header = (await panel.getByRole('heading', { name: 'Reply position', exact: true }).boundingBox())!
  const go = (await panel.getByRole('button', { name: 'Go', exact: true }).boundingBox())!
  expect(header.y).toBeGreaterThanOrEqual(bounds.y)
  expect(header.y + header.height).toBeLessThanOrEqual(bounds.y + bounds.height)
  expect(go.y).toBeGreaterThanOrEqual(bounds.y)
  expect(go.y + go.height).toBeLessThanOrEqual(bounds.y + bounds.height)
  expect(await panel.evaluate(element => element.scrollHeight - element.clientHeight)).toBeLessThanOrEqual(1)
  await page.mouse.click(10, Math.min(30, bounds.y / 2))
  await expect(panel).toBeHidden()
  await expect(trigger).toBeVisible()
  await expect(trigger).toContainText('1 / 6')
})

test('entering a floor jumps after Go and the earliest button jumps immediately', async ({ page }) => {
  await openTopic(page)
  const trigger = page.getByRole('button', { name: 'Reply position', exact: true })
  await trigger.click()
  const panel = page.getByRole('dialog', { name: 'Reply position' })
  await panel.getByRole('spinbutton', { name: 'Jump to post' }).fill('6')
  await panel.getByRole('button', { name: 'Go', exact: true }).click()
  await expect(panel).toBeHidden()
  await expect(trigger).toContainText('6 / 6')
  await expect(page.locator('[data-post-no="6"]')).toBeInViewport()
  await expect(page.getByTestId('post-position-rail')).toHaveAttribute('data-state', 'reading')
  await trigger.click()
  await expect(panel.getByRole('slider', { name: 'Reply position' })).toHaveAttribute('aria-valuenow', '6')
  await panel.getByRole('button', { name: /^Earliest content:/ }).click()
  await expect(panel).toBeHidden()
  await expect(trigger).toContainText('1 / 6')
  await expect(page.locator('[data-post-no="1"]')).toBeInViewport()
})

for (const count of [6, 62]) {
  test(`vertical dragging previews an integer floor in a ${count}-post topic and jumps on release`, async ({ page }, testInfo) => {
    const payload = count === 6 ? topicPage() : topicPage(Array.from({ length: count }, (_, index) => post(index + 1, index ? 100 : 1400)), count)
    await openTopic(page, payload)
    const trigger = page.getByRole('button', { name: 'Reply position', exact: true })
    await trigger.click()
    const panel = page.getByRole('dialog', { name: 'Reply position' })
    const slider = panel.getByRole('slider', { name: 'Reply position' })
    await slider.click({ trial: true })
    await expect(slider).toHaveAttribute('aria-orientation', 'vertical')
    const thumb = (await slider.boundingBox())!
    const track = (await panel.locator('[data-slot="mobile-position-slider"]').boundingBox())!
    const originalY = await page.evaluate(() => scrollY)
    const x = thumb.x + thumb.width / 2
    const y = thumb.y + thumb.height / 2
    const destinationY = track.y + track.height * 0.62
    const touch = testInfo.project.name === 'mobile-chromium' ? await page.context().newCDPSession(page) : undefined
    if (touch) {
      await touch.send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: [{ x, y }] })
      for (let step = 1; step <= 8; step++) await touch.send('Input.dispatchTouchEvent', { type: 'touchMove', touchPoints: [{ x, y: y + (destinationY - y) * step / 8 }] })
    } else {
      await page.mouse.move(x, y)
      await page.mouse.down()
      await page.mouse.move(x, destinationY, { steps: 8 })
    }
    const target = Number(await slider.getAttribute('aria-valuenow'))
    expect(Number.isInteger(target)).toBe(true)
    expect(target).toBeGreaterThan(1)
    expect(target).toBeLessThan(count)
    expect(await page.evaluate(() => scrollY)).toBe(originalY)
    await expect(panel).toBeVisible()
    if (touch) { await touch.send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] }); await touch.detach() }
    else await page.mouse.up()
    await expect(panel).toBeHidden()
    await expect(trigger).toContainText(`${target} / ${count}`)
    await expect(page.locator(`[data-post-no="${target}"]`)).toBeInViewport()
    await expect(page.getByTestId('post-position-rail')).toHaveAttribute('data-state', 'reading')
  })
}

test('mobile navigation resolves a missing unloaded floor after dismissing the panel', async ({ page }) => {
  let release!: () => void
  const pending = new Promise<void>(resolve => { release = resolve })
  const anchors: number[] = []
  await page.route('**/api/forum/posts/window**', async route => {
    anchors.push(Number(new URL(route.request().url()).searchParams.get('anchorPostNo')))
    await pending
    await route.fulfill({ json: { code: 0, result: {
      posts: [post(78, 500), post(79, 120)], replyTargets: [], beforePostNo: 78, afterPostNo: 79,
      hasBefore: true, hasAfter: false, total: 80, maxPostNo: 80,
    } } })
  })
  await openTopic(page, topicPage([post(1, 1400), post(2, 400)], 80, true))
  const trigger = page.getByRole('button', { name: 'Reply position', exact: true })
  await trigger.click()
  const panel = page.getByRole('dialog', { name: 'Reply position' })
  await panel.getByRole('button', { name: /^Latest reply:/ }).click()
  await expect(panel).toBeHidden()
  await expect.poll(() => anchors).toEqual([80])
  await expect(page.locator('[data-post-no="1"]')).toBeVisible()
  release()
  await expect(trigger).toContainText('79 / 80')
  await expect(page.locator('[data-post-no="79"]')).toBeInViewport()
  await expect(page.getByTestId('post-position-rail')).toHaveAttribute('data-state', 'reading')
  await trigger.click()
  await expect(panel.getByRole('slider', { name: 'Reply position' })).toHaveAttribute('aria-valuenow', '79')
  expect(anchors).toEqual([80])
})

test('the mobile timeline entrance stays out of the reply composer', async ({ page }) => {
  await openTopic(page)
  const trigger = page.getByRole('button', { name: 'Reply position', exact: true })
  await page.evaluate(() => window.scrollTo(0, 350))
  const floatingReply = page.locator('[data-slot="topic-reply-floating"]')
  await expect(floatingReply).toBeInViewport()
  await floatingReply.click()
  const composer = page.locator('.goose-composer-surface[data-state="open"] section')
  await expect(composer).toBeVisible()
  await expect(trigger).toBeHidden()
  await expect(floatingReply).toBeHidden()
  await composer.locator('header').getByRole('button', { name: 'Close', exact: true }).click()
  await expect(composer).toBeHidden()
  await expect(trigger).toBeVisible()
  await expect(floatingReply).toBeInViewport()
})

test('the vertical mobile slider uses the theme primary and accepts taps far from the thin rail', async ({ page }, testInfo) => {
  await openTopic(page, topicPage(Array.from({ length: 62 }, (_, index) => post(index + 1, index ? 100 : 1400)), 62))
  const trigger = page.getByRole('button', { name: 'Reply position', exact: true })
  await trigger.click()
  const panel = page.getByRole('dialog', { name: 'Reply position' })
  const slider = panel.getByRole('slider', { name: 'Reply position' })
  await slider.click({ trial: true })
  const thumb = (await slider.boundingBox())!
  expect(thumb.width).toBeGreaterThanOrEqual(44)
  expect(thumb.height).toBeGreaterThanOrEqual(44)
  const root = panel.locator('[data-slot="mobile-position-slider"]')
  const track = (await root.boundingBox())!
  expect(track.height).toBeGreaterThanOrEqual(192)
  expect(track.width).toBeGreaterThan(200)

  // Verify theme inheritance with a non-default primary, not just the stock blue.
  await panel.evaluate(element => element.style.setProperty('--primary', 'rgb(170, 42, 220)'))
  const colors = await slider.evaluate(element => {
    const probe = document.createElement('span')
    element.parentElement!.append(probe)
    probe.style.backgroundColor = 'var(--primary)'
    const primary = getComputedStyle(probe).backgroundColor
    probe.style.backgroundColor = 'var(--success)'
    const success = getComputedStyle(probe).backgroundColor
    const thumbColor = getComputedStyle(element.querySelector('span')!).backgroundColor
    probe.remove()
    return { primary, success, thumbColor }
  })
  expect(colors.thumbColor).toBe(colors.primary)
  expect(colors.thumbColor).not.toBe(colors.success)

  const x = track.x + track.width * 0.75
  const y = track.y + track.height * 0.65
  expect(x - (thumb.x + thumb.width / 2)).toBeGreaterThan(60)
  if (testInfo.project.name === 'mobile-chromium') await page.touchscreen.tap(x, y)
  else await page.mouse.click(x, y)
  await expect(panel).toBeHidden()
  await expect(page.getByTestId('post-position-rail')).toHaveAttribute('data-state', 'reading')
  const selected = Number((await trigger.innerText()).split('/')[0].trim())
  expect(selected).toBeGreaterThan(30)
  expect(selected).toBeLessThan(50)
  await expect(page.locator(`[data-post-no="${selected}"]`)).toBeInViewport()
})

test('touch cancellation preserves reading and a resized slider uses its new hit area', async ({ page }) => {
  await openTopic(page, topicPage(Array.from({ length: 62 }, (_, index) => post(index + 1, index ? 100 : 1400)), 62))
  await page.getByRole('button', { name: 'Reply position', exact: true }).click()
  const panel = page.getByRole('dialog', { name: 'Reply position' })
  const slider = panel.getByRole('slider', { name: 'Reply position' })
  await slider.click({ trial: true })
  const thumb = (await slider.boundingBox())!
  const originalY = await page.evaluate(() => scrollY)
  const x = thumb.x + thumb.width / 2
  const y = thumb.y + thumb.height / 2
  const touch = await page.context().newCDPSession(page)
  await touch.send('Emulation.setTouchEmulationEnabled', { enabled: true, maxTouchPoints: 1 })
  await touch.send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: [{ x, y }] })
  await touch.send('Input.dispatchTouchEvent', { type: 'touchMove', touchPoints: [{ x, y: y + 160 }] })
  await expect(slider).not.toHaveAttribute('aria-valuenow', '1')
  await touch.send('Input.dispatchTouchEvent', { type: 'touchCancel', touchPoints: [] })
  await expect(panel).toBeVisible()
  await expect(slider).toHaveAttribute('aria-valuenow', '1')
  expect(await page.evaluate(() => scrollY)).toBe(originalY)
  await expect(page.locator('[data-slot="mobile-topic-position"]')).toContainText('1 / 62')
  const root = panel.locator('[data-slot="mobile-position-slider"]')
  const previousHeight = (await root.boundingBox())!.height
  await page.setViewportSize({ width: 390, height: 600 })
  await expect.poll(async () => (await root.boundingBox())!.height).toBeLessThan(previousHeight)
  const resized = (await root.boundingBox())!
  const point = { x: resized.x + resized.width * 0.75, y: resized.y + resized.height * 0.75 }
  await touch.send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: [point] })
  const expected = Math.round(1 + (62 - 1) * 0.75)
  await expect(slider).toHaveAttribute('aria-valuenow', String(expected))
  await touch.send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] })
  await touch.detach()
  await expect(panel).toBeHidden()
  await expect(page.getByRole('button', { name: 'Reply position', exact: true })).toContainText(`${expected} / 62`)
  await expect(page.locator(`[data-post-no="${expected}"]`)).toBeInViewport()
})
