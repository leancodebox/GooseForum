import { expect, test } from './fixtures'
import type { Locator, Page } from '@playwright/test'

async function prepareDashboard(page: Page) {
  await page.clock.setFixedTime(new Date('2026-11-17T12:00:00Z'))
  await page.route('**/__goose_page/**', route => route.fulfill({ json: {
    component: 'admin.shell', props: {}, url: '/admin?lang=zh', version: '1.0', meta: { title: 'Dashboard' },
    layout: { site: { name: 'GooseForum', description: '', logo: '', favicon: '', brandType: 'default', brandText: '', brandImage: '' }, viewer: { id: 1, username: 'admin', email: '', avatarUrl: '', isAuthenticated: true, canAccessAdmin: true, isModerator: false, requiresEmailVerification: false, adminPermissions: [0] }, header: [], sidebar: { activeKey: 'topics', categories: [] }, footer: { links: [], primary: [] }, unread: { notifications: false, messages: false }, theme: { enabled: false, current: 'gf-light', themeColor: '#fbfdff' } },
  } }))
  await page.route('**/api/forum/get-site-statistics', route => route.fulfill({ json: { code: 0, result: {} } }))
  await page.route('**/api/admin/server-version', route => route.fulfill({ json: { code: 0, result: { version: 'dev', mode: 'development' } } }))
  await page.route('https://api.github.com/repos/leancodebox/GooseForum/releases', route => route.fulfill({ json: [] }))
  const requests: { startDate: string; endDate: string }[] = []
  await page.route('**/api/admin/traffic-overview', route => {
    requests.push(route.request().postDataJSON())
    return route.fulfill({ json: { code: 0, result: [] } })
  })
  await page.goto('/admin?lang=zh')
  await expect(page.getByText('流量概览', { exact: true })).toBeVisible()
  await page.addStyleTag({ content: '[data-slot="calendar"] *, [data-slot="popover-content"] { transition: none !important; animation: none !important; }' })
  await page.getByRole('button', { name: /2026年11月11日.*2026年11月17日/ }).click()
  const popup = page.locator('[data-slot="popover-content"][data-state="open"]')
  await expect(popup).toBeVisible()
  return { popup, requests }
}

function day(popup: Locator, date: string) {
  return popup.locator(`[role="gridcell"][data-day="${date}"]:not([data-outside]) button`)
}

// Compare rendered pixels with and without the shadow. This catches an adjacent
// range cell painting over the ring even when computed box-shadow is correct.
async function expectVisibleFocusRing(page: Page, button: Locator) {
  await button.scrollIntoViewIfNeeded()
  await page.mouse.move(0, 0)
  await button.focus()
  await expect(button).toBeFocused()
  const box = await button.boundingBox()
  if (!box) throw new Error('Focused date has no bounds')
  const clip = { x: Math.floor(box.x) - 4, y: Math.floor(box.y) - 4, width: Math.ceil(box.width) + 10, height: Math.ceil(box.height) + 10 }
  const style = await button.getAttribute('style')
  // The normal ring color matches selected-day blue exactly. A contrasting
  // probe color lets us detect painting over that neighbor without changing
  // the ring's real dimensions, focus state or stacking order.
  await button.evaluate(element => element.style.setProperty('--tw-ring-color', 'rgb(255 0 255 / 0.8)', 'important'))
  const withRing = await page.screenshot({ clip, scale: 'css', animations: 'disabled' })
  await button.evaluate(element => element.style.setProperty('box-shadow', 'none', 'important'))
  const withoutRing = await page.screenshot({ clip, scale: 'css', animations: 'disabled' })
  await button.evaluate((element, original) => {
    if (original === null) element.removeAttribute('style')
    else element.setAttribute('style', original)
  }, style)
  const differences = await page.evaluate(async ({ withRing, withoutRing, box, clip }) => {
    async function pixels(base64: string) {
      const image = new Image()
      image.src = `data:image/png;base64,${base64}`
      await image.decode()
      const canvas = document.createElement('canvas')
      canvas.width = image.width
      canvas.height = image.height
      const context = canvas.getContext('2d')!
      context.drawImage(image, 0, 0)
      return { data: context.getImageData(0, 0, canvas.width, canvas.height).data, width: canvas.width }
    }
    const [before, after] = await Promise.all([pixels(withRing), pixels(withoutRing)])
    const left = box.x - clip.x, top = box.y - clip.y
    const points = {
      left: [left - 2, top + box.height / 2], right: [left + box.width + 1, top + box.height / 2],
      top: [left + box.width / 2, top - 2], bottom: [left + box.width / 2, top + box.height + 1],
    }
    return Object.fromEntries(Object.entries(points).map(([edge, [x, y]]) => {
      const offset = (Math.round(y) * before.width + Math.round(x)) * 4
      return [edge, Math.max(...[0, 1, 2].map(channel => Math.abs(before.data[offset + channel] - after.data[offset + channel])))]
    }))
  }, { withRing: withRing.toString('base64'), withoutRing: withoutRing.toString('base64'), box, clip })
  const stacking = await button.evaluate(element => {
    const cell = element.closest('[role="gridcell"]')!
    const next = cell.nextElementSibling
    const style = getComputedStyle(cell)
    return { date: cell.getAttribute('data-day'), focused: cell.getAttribute('data-focused'), zIndex: style.zIndex, display: style.display, position: style.position, isolation: style.isolation, classes: cell.className, nextZIndex: next && getComputedStyle(next).zIndex }
  })
  for (const [edge, difference] of Object.entries(differences)) expect(difference, `${edge} edge is covered: ${JSON.stringify({ differences, stacking, box, clip })}`).toBeGreaterThan(5)
}

for (const theme of ['gf-light', 'gf-dark']) {
  test(`keeps every date focus ring visible across range cells in ${theme}`, async ({ page }, info) => {
    const { popup } = await prepareDashboard(page)
    await page.evaluate(theme => document.documentElement.setAttribute('data-theme', theme), theme)
    await day(popup, '2026-11-10').click()
    await expect(day(popup, '2026-11-10')).toHaveAttribute('data-range-start', 'true')
    await expect(day(popup, '2026-11-17')).toHaveAttribute('data-range-end', 'true')
    for (const date of ['2026-11-10', '2026-11-17', '2026-11-12', '2026-11-09']) {
      await expectVisibleFocusRing(page, day(popup, date))
    }
    await day(popup, '2026-11-10').focus()
    await page.keyboard.press('ArrowRight')
    await expect(day(popup, '2026-11-11')).toBeFocused()
    await expectVisibleFocusRing(page, day(popup, '2026-11-11'))
    await page.screenshot({ path: info.outputPath(`calendar-focus-${theme}.png`), fullPage: true })
    expect(await popup.evaluate(element => {
      const bounds = element.getBoundingClientRect()
      return bounds.left >= 0 && bounds.right <= innerWidth + 1
    })).toBe(true)
  })
}

test('keeps same-day, cross-week and cross-month ranges usable with apply, cancel and presets', async ({ page }, info) => {
  const { popup, requests } = await prepareDashboard(page)
  await day(popup, '2026-11-10').click()
  await popup.getByRole('button', { name: '取消', exact: true }).click()
  await expect(popup).not.toBeVisible()
  expect(requests).toHaveLength(1)
  await page.getByRole('button', { name: /2026年11月11日.*2026年11月17日/ }).click()
  await day(popup, '2026-11-17').click()
  await expect(day(popup, '2026-11-17')).toHaveAttribute('data-range-start', 'true')
  await expect(day(popup, '2026-11-17')).toHaveAttribute('data-range-end', 'true')
  const single = day(popup, '2026-11-17')
  const corners = await single.evaluate(element => {
    const style = getComputedStyle(element)
    return [style.borderTopLeftRadius, style.borderTopRightRadius, style.borderBottomLeftRadius, style.borderBottomRightRadius].map(parseFloat)
  })
  expect(corners.every(radius => radius > 0)).toBe(true)
  await popup.getByRole('button', { name: '应用', exact: true }).focus()
  await expect(single).not.toBeFocused()
  await page.mouse.move(0, 0)
  for (const singleTheme of ['gf-dark', 'gf-light']) {
    await page.evaluate(theme => document.documentElement.setAttribute('data-theme', theme), singleTheme)
    await page.screenshot({ path: info.outputPath(`calendar-single-${singleTheme}.png`), fullPage: true })
    const singleBox = await single.boundingBox()
    if (!singleBox) throw new Error('Single day has no bounds')
    const singleClip = { x: Math.floor(singleBox.x) - 6, y: Math.floor(singleBox.y) - 6, width: Math.ceil(singleBox.width) + 12, height: Math.ceil(singleBox.height) + 12 }
    const singlePng = await page.screenshot({ path: info.outputPath(`calendar-single-detail-${singleTheme}.png`), clip: singleClip, scale: 'css' })
    const cornerDifferences = await page.evaluate(async ({ png, box, clip }) => {
      const image = new Image()
      image.src = `data:image/png;base64,${png}`
      await image.decode()
      const canvas = document.createElement('canvas')
      canvas.width = image.width
      canvas.height = image.height
      const context = canvas.getContext('2d')!
      context.drawImage(image, 0, 0)
      const { data } = context.getImageData(0, 0, canvas.width, canvas.height)
      const background = Array.from(data.slice(0, 3))
      const left = Math.round(box.x - clip.x), top = Math.round(box.y - clip.y)
      return [[left, top], [left + box.width - 1, top], [left, top + box.height - 1], [left + box.width - 1, top + box.height - 1]].map(([x, y]) => {
        const offset = (y * canvas.width + x) * 4
        return Math.max(...background.map((channel, index) => Math.abs(channel - data[offset + index])))
      })
    }, { png: singlePng.toString('base64'), box: singleBox, clip: singleClip })
    expect(cornerDifferences, 'A single-day range must not leave a square range background behind its round button').toEqual([0, 0, 0, 0])
  }
  await popup.getByRole('button', { name: '应用', exact: true }).click()
  await expect.poll(() => requests.at(-1)).toEqual({ startDate: '2026-11-17', endDate: '2026-11-17' })
  await page.getByRole('button', { name: /2026年11月17日.*2026年11月17日/ }).click()
  await day(popup, '2026-10-30').click()
  await day(popup, '2026-11-05').click()
  await expect(day(popup, '2026-10-30')).toHaveAttribute('data-range-start', 'true')
  await expect(day(popup, '2026-11-05')).toHaveAttribute('data-range-end', 'true')
  await expect(day(popup, '2026-11-02')).toHaveAttribute('data-range-middle', 'true')
  await expectVisibleFocusRing(page, day(popup, '2026-11-05'))
  await page.screenshot({ path: info.outputPath('calendar-cross-month.png'), fullPage: true })
  await popup.getByRole('button', { name: '应用', exact: true }).click()
  await expect.poll(() => requests.at(-1)).toEqual({ startDate: '2026-10-30', endDate: '2026-11-05' })
  await page.getByRole('button', { name: /2026年10月30日.*2026年11月5日/ }).click()
  await popup.getByRole('button', { name: '近 7 天', exact: true }).click()
  await expect.poll(() => requests.at(-1)).toEqual({ startDate: '2026-11-11', endDate: '2026-11-17' })
  await expect(popup).not.toBeVisible()
})
