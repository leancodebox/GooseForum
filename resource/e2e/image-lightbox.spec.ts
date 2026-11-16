import { expect, test } from './fixtures'

const imagePost = {
  id: 61, topicId: 60, postNo: 1, content: 'Pictures',
  renderedContent: '<p><img src="/lightbox-test.svg" alt="Mountain view"></p><p><img src="/lightbox-test.svg?second" alt="Second view"></p>',
  processStatus: 0, isHidden: false, canModerate: false, isOwnPost: false,
  author: { id: 7, username: 'alice', avatarUrl: '' }, createdAt: '2026-10-11T08:00:00Z',
}
const topic = {
  component: 'topic.detail',
  props: {
    topic: {
      id: 60, title: 'Pictures', description: '', url: '/p/test/60', topicStatus: 1, processStatus: 0,
      author: imagePost.author, participants: [], categories: [], replyCount: 0, maxPostNo: 1,
      viewCount: 10, likeCount: 0, isLiked: false, isBookmarked: false, isWatched: false,
      createdAt: imagePost.createdAt, updatedAt: imagePost.createdAt,
    },
    postStream: { posts: [imagePost], replyTargets: [], beforePostNo: 1, afterPostNo: 1, hasBefore: false, hasAfter: false, total: 1, maxPostNo: 1 },
    hotTopics: [], permissions: { isOwnTopic: false, canPost: false, canModerateTopic: false },
  },
  layout: {
    site: { name: 'GooseForum', description: '', logo: '', favicon: '', brandType: 'default', brandText: '', brandImage: '' },
    viewer: { id: 0, username: '', email: '', avatarUrl: '', isAuthenticated: false, canAccessAdmin: false, isModerator: false, requiresEmailVerification: false, adminPermissions: [] },
    header: [], sidebar: { activeKey: 'topics', categories: [] }, footer: { links: [], primary: [] },
    unread: { notifications: false, messages: false }, theme: { enabled: false, current: 'gf-light', themeColor: '#fbfdff' },
  },
  meta: { title: 'Pictures' }, url: '/p/test/60', version: '1.0',
}

test.beforeEach(async ({ page }) => {
  await page.route('**/__goose_page/**', route => route.fulfill({ json: topic }))
  await page.route('**/lightbox-test.svg*', route => route.fulfill({ contentType: 'image/svg+xml', body: '<svg xmlns="http://www.w3.org/2000/svg" width="1600" height="1000"><rect width="1600" height="1000" fill="#bad9ed"/><path d="M0 1000L650 100L1600 1000" fill="#6f917d"/></svg>' }))
  await page.goto('/p/test/60?lang=en')
  await page.getByRole('img', { name: 'Mountain view', exact: true }).click()
  await expect(page.getByRole('dialog', { name: 'Image preview' })).toBeVisible()
})

for (const theme of ['gf-light', 'gf-dark']) {
  test(`the image viewer keeps a translucent black backdrop in ${theme} and closes on blank-area input`, async ({ page, isMobile }, testInfo) => {
    await page.evaluate(value => document.documentElement.setAttribute('data-theme', value), theme)
    const dialog = page.getByRole('dialog', { name: 'Image preview' })
    const viewport = dialog.locator('div.overflow-auto')
    await expect(dialog).toHaveCSS('opacity', '1')
    await expect(dialog.locator('img')).toHaveCSS('opacity', '1')
    await expect(page.locator('[data-slot="dialog-overlay"]')).toHaveCSS('opacity', '1')
    const colors = await dialog.evaluate(element => {
      const context = document.createElement('canvas').getContext('2d')!
      const rgba = (color: string) => {
        context.clearRect(0, 0, 1, 1)
        context.fillStyle = color
        context.fillRect(0, 0, 1, 1)
        return Array.from(context.getImageData(0, 0, 1, 1).data)
      }
      const overlay = getComputedStyle(document.querySelector('[data-slot="dialog-overlay"]')!)
      return {
        backdrop: rgba(overlay.backgroundColor), content: rgba(getComputedStyle(element).backgroundColor), blur: overlay.backdropFilter,
        controls: Array.from(element.querySelectorAll('button'), button => rgba(getComputedStyle(button).backgroundColor)),
      }
    })
    expect(colors.backdrop.slice(0, 3)).toEqual([0, 0, 0])
    expect(colors.backdrop[3] / 255).toBeGreaterThanOrEqual(0.6)
    expect(colors.backdrop[3] / 255).toBeLessThanOrEqual(0.7)
    expect(colors.content[3]).toBe(0)
    expect(['none', 'blur(0px)']).toContain(colors.blur)
    // Arrow controls can overlap a pure-white image: keep their white icons legible there too.
    for (const [red, green, blue, alpha] of colors.controls) {
      const linear = (channel: number) => {
        const value = (channel * alpha / 255 + 255 - alpha) / 255
        return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4
      }
      const luminance = 0.2126 * linear(red) + 0.7152 * linear(green) + 0.0722 * linear(blue)
      expect(1.05 / (luminance + 0.05)).toBeGreaterThanOrEqual(3)
    }
    for (const name of ['Original size', 'Close', 'Previous image', 'Next image']) {
      await expect(dialog.getByRole('button', { name, exact: true })).toHaveCSS('color', 'rgb(255, 255, 255)')
    }
    await page.screenshot({ path: testInfo.outputPath(`image-lightbox-${theme}.png`) })
    const box = (await viewport.boundingBox())!
    if (isMobile) await page.touchscreen.tap(box.x + 5, box.y + 10)
    else await page.mouse.click(box.x + 5, box.y + 10)
    await expect(dialog).toHaveCount(0)
  })
}

test('image navigation, zoom and panning do not dismiss the viewer', async ({ page }) => {
  const dialog = page.getByRole('dialog', { name: 'Image preview' })
  await dialog.getByRole('button', { name: 'Next image', exact: true }).click()
  await expect(dialog).toContainText('2 / 2')
  await dialog.getByRole('button', { name: 'Previous image', exact: true }).click()
  await expect(dialog).toContainText('1 / 2')
  const image = dialog.locator('img')
  await image.click()
  await expect(dialog.getByRole('button', { name: 'Fit to screen', exact: true })).toHaveAttribute('aria-pressed', 'true')
  const viewport = dialog.locator('div.overflow-auto')
  const before = await viewport.evaluate(element => ({ left: element.scrollLeft, top: element.scrollTop }))
  const box = (await viewport.boundingBox())!
  await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2)
  await page.mouse.down()
  await page.mouse.move(box.x + box.width / 2 - 60, box.y + box.height / 2 - 60, { steps: 5 })
  await page.mouse.up()
  await expect(dialog).toBeVisible()
  await expect(dialog.getByRole('button', { name: 'Fit to screen', exact: true })).toHaveAttribute('aria-pressed', 'true')
  const after = await viewport.evaluate(element => ({ left: element.scrollLeft, top: element.scrollTop }))
  expect(after.left > before.left || after.top > before.top).toBe(true)
  await page.keyboard.press('Escape')
  await expect(dialog).toHaveCount(0)
})

test('a gesture starting on the backdrop and ending on the image does not close it', async ({ page }) => {
  const dialog = page.getByRole('dialog', { name: 'Image preview' })
  const viewport = (await dialog.locator('div.overflow-auto').boundingBox())!
  const image = (await dialog.locator('img').boundingBox())!
  await page.mouse.move(viewport.x + 5, viewport.y + 10)
  await page.mouse.down()
  await page.mouse.move(image.x + image.width / 2, image.y + image.height / 2, { steps: 5 })
  await page.mouse.up()
  await expect(dialog).toBeVisible()
  await expect(dialog.getByRole('button', { name: 'Original size', exact: true })).toHaveAttribute('aria-pressed', 'false')
})
