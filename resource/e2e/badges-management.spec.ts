import { expect, test } from './fixtures'
import type { Page } from '@playwright/test'

async function prepareBadges(page: Page, permissions = [0]) {
  const base = { type: 'custom', grantMode: 'manual', description: 'Recognizes helpful community members', iconType: 'asset', iconKey: '', iconUrl: '/static/badges/contributor.svg', color: 'blue', level: 'bronze', isEnabled: true, isWearable: true, sortOrder: 1000, canDelete: true, grantedCount: 0 }
  let badges = [
    { ...base, code: 'helper', name: 'Helpful member', canDelete: false, grantedCount: 1 },
    { ...base, code: 'writer_10', name: 'Writer', type: 'system', grantMode: 'auto', grantRule: { metric: 'topics', threshold: 10 }, color: 'sky', level: 'silver', canDelete: false },
    { ...base, code: 'retired', name: 'Retired event', isEnabled: false },
  ]
  let recipients = [{ userId: 9, username: 'alice', avatarUrl: '', source: 'manual', reason: 'Welcomed new members', grantedBy: 1, grantedByName: 'admin', grantedAt: '2026-11-17T10:00:00+08:00' }]
  const updates: Record<string, unknown>[] = []
  const saves: Record<string, unknown>[] = []
  await page.route('**/__goose_page/**', route => route.fulfill({ json: {
    component: 'admin.shell', props: {}, url: '/admin/badges?lang=en', version: '1.0', meta: { title: 'Badges' },
    layout: { site: { name: 'GooseForum', description: '', logo: '', favicon: '', brandType: 'default', brandText: '', brandImage: '' }, viewer: { id: 1, username: 'admin', email: '', avatarUrl: '', isAuthenticated: true, canAccessAdmin: true, isModerator: false, requiresEmailVerification: false, adminPermissions: permissions }, header: [], sidebar: { activeKey: 'topics', categories: [] }, footer: { links: [], primary: [] }, unread: { notifications: false, messages: false }, theme: { enabled: false, current: 'gf-light', themeColor: '#fbfdff' } },
  } }))
  await page.route('**/api/admin/badges', route => route.fulfill({ json: { code: 0, result: badges } }))
  await page.route('**/api/admin/badge-save', route => {
    const data = route.request().postDataJSON()
    saves.push(data)
    const next = { ...data, code: data.code || 'custom_created', grantedCount: data.grantedCount || 0 }
    badges = [...badges.filter(item => item.code !== next.code), next]
    return route.fulfill({ json: { code: 0, result: true } })
  })
  await page.route('**/api/admin/badge-recipients', route => {
    const { page: currentPage = 1, pageSize = 20, search = '' } = route.request().postDataJSON()
    const matches = recipients.filter(item => item.username.includes(search))
    return route.fulfill({ json: { code: 0, result: { list: matches.slice((currentPage - 1) * pageSize, currentPage * pageSize), ...(search ? {} : { total: recipients.length }), hasNext: matches.length > currentPage * pageSize, page: currentPage, size: pageSize } } })
  })
  await page.route('**/api/admin/user-list', route => route.fulfill({ json: { code: 0, result: { list: [{ userId: 10, username: 'bob', avatarUrl: '' }], total: 1, page: 1, size: 10 } } }))
  await page.route('**/api/admin/badge-recipient-update', route => {
    const data = route.request().postDataJSON()
    updates.push(data)
    if (data.action === 'revoke') recipients = recipients.filter(item => item.userId !== data.userId)
    else recipients.push({ userId: 10, username: 'bob', avatarUrl: '', source: 'manual', reason: data.reason, grantedBy: 1, grantedByName: 'admin', grantedAt: '2026-11-17T11:00:00+08:00' })
    badges = badges.map(item => item.code === data.code ? { ...item, grantedCount: recipients.length } : item)
    return route.fulfill({ json: { code: 0, result: true } })
  })
  await page.route('**/api/admin/img-upload/init', route => route.fulfill({ json: { code: 0, result: { mode: 'proxy' } } }))
  await page.route('**/api/admin/img-upload', route => route.fulfill({ json: { code: 0, result: { url: '/static/badges/sponsor.svg' } } }))
  await page.goto('/admin/badges?lang=en')
  await expect(page.getByRole('heading', { name: 'Badges', exact: true, level: 2 })).toBeVisible()
  return { saves, updates }
}

test('creates a visual badge with generated code and responsive previews', async ({ page }, info) => {
  const { saves } = await prepareBadges(page)
  await page.getByRole('button', { name: 'New badge', exact: true }).click()
  const editor = page.getByRole('dialog')
  await editor.getByLabel('Name', { exact: true }).fill('Community guide')
  await editor.getByLabel('Description', { exact: true }).fill('Helps new members find their way')
  await editor.getByRole('button', { name: 'Robot', exact: true }).click()
  await editor.getByLabel('Level', { exact: true }).click()
  await page.getByRole('option', { name: 'Gold', exact: true }).click()
  await editor.getByLabel('Color', { exact: true }).click()
  await page.getByRole('option', { name: 'Amber', exact: true }).click()
  await expect(editor.getByRole('img', { name: 'Community guide' }).first()).toHaveAttribute('src', '/static/badges/robot.svg')
  await expect(editor.getByLabel('Code', { exact: true })).not.toBeVisible()
  await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true)
  await page.screenshot({ path: `/tmp/gooseforum-badge-editor-${info.project.name}.png`, fullPage: true })
  await editor.getByRole('button', { name: 'Save', exact: true }).click()
  await expect(editor).not.toBeVisible()
  expect(saves[0]).toMatchObject({ code: '', name: 'Community guide', iconUrl: '/static/badges/robot.svg', color: 'amber', level: 'gold', grantMode: 'manual' })
  await expect(page.getByRole('article', { name: 'Community guide' })).toBeVisible()
  await page.screenshot({ path: `/tmp/gooseforum-badges-list-${info.project.name}.png`, fullPage: true })
})

test('uploads replacement artwork and shows an immutable automatic condition', async ({ page }) => {
  const { saves } = await prepareBadges(page)
  await page.getByRole('button', { name: 'Edit Writer', exact: true }).click()
  const editor = page.getByRole('dialog')
  await expect(editor.getByText('Automatically awarded after publishing 10 topics.', { exact: true })).toBeVisible()
  await expect(editor.getByRole('combobox')).toHaveCount(2)
  await editor.getByRole('button', { name: 'Upload image', exact: true }).click()
  await editor.locator('input[type=file]').setInputFiles({ name: 'badge.png', mimeType: 'image/png', buffer: Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVQIHWP4z8DwHwAFgAI/ScLbtAAAAABJRU5ErkJggg==', 'base64') })
  await expect(editor.getByRole('img', { name: 'Writer' }).first()).toHaveAttribute('src', '/static/badges/sponsor.svg')
  await editor.getByRole('button', { name: 'Save', exact: true }).click()
  await expect(editor).not.toBeVisible()
  expect(saves[0]).toMatchObject({ code: 'writer_10', grantMode: 'auto', iconUrl: '/static/badges/sponsor.svg' })
})

test('grants and revokes from the recipients panel with recorded reasons', async ({ page }, info) => {
  const { updates } = await prepareBadges(page)
  await page.getByRole('article', { name: 'Helpful member' }).getByRole('button', { name: 'Recipients 1' }).click()
  const members = page.getByRole('dialog', { name: 'Helpful member · Recipients', exact: true })
  await expect(members.getByText('Welcomed new members')).toBeVisible()
  await members.getByRole('button', { name: 'Revoke', exact: true }).click()
  let action = page.getByRole('dialog', { name: 'Revoke', exact: true })
  await action.getByRole('button', { name: 'Revoke', exact: true }).click()
  await expect(action.getByRole('alert')).toHaveText('Enter a reason for this action.')
  expect(updates).toHaveLength(0)
  await action.getByLabel('Reason for revoking').fill('Awarded to the wrong member')
  await action.getByRole('button', { name: 'Revoke', exact: true }).click()
  await expect(action).not.toBeVisible()
  await expect(members.getByText('No recipients yet')).toBeVisible()
  await members.getByRole('button', { name: 'Grant badge', exact: true }).click()
  action = page.getByRole('dialog', { name: 'Grant badge', exact: true })
  await action.getByLabel('Select a member', { exact: true }).fill('bob')
  await action.getByRole('button', { name: 'Search', exact: true }).click()
  await action.getByRole('button', { name: 'bob #10' }).click()
  await action.getByLabel('Reason for granting').fill('Helped organize the community event')
  const reason = action.getByLabel('Reason for granting')
  await expect(reason).toBeFocused()
  await expect.poll(() => reason.evaluate(element => {
    const field = element.getBoundingClientRect()
    const port = element.closest('form')!.getBoundingClientRect()
    return Math.min(field.left - port.left, port.right - field.right, field.top - port.top, port.bottom - field.bottom)
  })).toBeGreaterThanOrEqual(3)
  await page.screenshot({ path: `/tmp/gooseforum-badge-focus-reason-${info.project.name}.png`, fullPage: true })
  await action.getByRole('button', { name: 'Grant badge', exact: true }).click()
  await expect(action).not.toBeVisible()
  await expect(members.getByRole('link', { name: 'bob', exact: true })).toBeVisible()
  expect(updates).toEqual([
    { code: 'helper', userId: 9, action: 'revoke', reason: 'Awarded to the wrong member' },
    { code: 'helper', userId: 10, action: 'grant', reason: 'Helped organize the community event' },
  ])
  await page.screenshot({ path: `/tmp/gooseforum-badge-recipients-${info.project.name}.png`, fullPage: true })
})

test('requires explicit disabling acknowledgment and preserves settings after a save error', async ({ page }) => {
  const { saves } = await prepareBadges(page)
  await page.getByRole('button', { name: 'Edit Helpful member' }).click()
  const editor = page.getByRole('dialog')
  await editor.getByRole('switch', { name: 'Enabled', exact: true }).click()
  await editor.getByRole('button', { name: 'Save', exact: true }).click()
  expect(saves).toHaveLength(0)
  await editor.getByRole('checkbox').check()
  await page.route('**/api/admin/badge-save', route => route.fulfill({ json: { code: 1, message: 'Save rejected', messageCode: 'admin.badge.saveFailed' } }))
  await editor.getByRole('button', { name: 'Save', exact: true }).click()
  await expect(editor.getByRole('alert')).toBeVisible()
  await expect(editor.getByRole('switch', { name: 'Enabled', exact: true })).not.toBeChecked()
  await expect(editor.getByLabel('Name', { exact: true })).toHaveValue('Helpful member')
})

test('filters definitions without giving site-only administrators recipient management', async ({ page }) => {
  await prepareBadges(page, [5])
  await expect(page.getByRole('article', { name: 'Helpful member' }).getByRole('button', { name: 'Recipients 1' })).toBeDisabled()
  await page.getByRole('combobox', { name: 'Status', exact: true }).click()
  await page.getByRole('option', { name: 'Disabled', exact: true }).click()
  await expect(page.getByRole('article')).toHaveCount(1)
  await expect(page.getByRole('article', { name: 'Retired event' })).toBeVisible()
  await page.getByRole('textbox', { name: 'Search badges by name or description' }).fill('missing')
  await expect(page.getByText('No matching badges')).toBeVisible()
})

test('reports a broken image link and recovers after choosing a built-in icon', async ({ page }) => {
  const { saves } = await prepareBadges(page)
  await page.route('https://images.example.test/broken.png', route => route.fulfill({ status: 404, body: 'missing' }))
  await page.getByRole('button', { name: 'Edit Helpful member' }).click()
  const editor = page.getByRole('dialog')
  await editor.getByRole('button', { name: 'Image link (advanced)', exact: true }).click()
  await editor.getByLabel('Image link (advanced)', { exact: true }).fill('https://images.example.test/broken.png')
  await expect(editor.getByRole('alert')).toContainText('This image could not be loaded.')
  await editor.getByRole('button', { name: 'Save', exact: true }).click()
  expect(saves).toHaveLength(0)
  await editor.getByRole('button', { name: 'Built-in icons', exact: true }).click()
  await editor.getByRole('button', { name: 'Crown', exact: true }).click()
  await editor.getByRole('button', { name: 'Save', exact: true }).click()
  await expect(editor).not.toBeVisible()
  expect(saves[0]).toMatchObject({ iconUrl: '/static/badges/king.svg' })
})

test('paginates recipients and resets the page when searching', async ({ page }) => {
  await prepareBadges(page)
  const requests: { page: number; search: string }[] = []
  await page.route('**/api/admin/badge-recipients', route => {
    const { page: currentPage, search = '' } = route.request().postDataJSON()
    requests.push({ page: currentPage, search })
    const all = Array.from({ length: 21 }, (_, index) => ({ userId: index + 1, username: `member-${index + 1}`, avatarUrl: '', source: 'manual', reason: 'Thank you', grantedBy: 0, grantedByName: '', grantedAt: '2026-11-17T10:00:00+08:00' }))
    const matches = all.filter(item => item.username.includes(search))
    return route.fulfill({ json: { code: 0, result: { list: matches.slice((currentPage - 1) * 20, currentPage * 20), page: currentPage, size: 20, hasNext: matches.length > currentPage * 20, ...(search ? {} : { total: all.length }) } } })
  })
  await page.getByRole('article', { name: 'Helpful member' }).getByRole('button', { name: 'Recipients 1' }).click()
  const members = page.getByRole('dialog')
  await expect(members.getByRole('link')).toHaveCount(20)
  await expect(members.getByText('21 current recipients', { exact: true })).toBeVisible()
  await expect(members.getByText('Page 1', { exact: true })).toBeVisible()
  await members.getByRole('button', { name: 'Next', exact: true }).click()
  await expect(members.getByRole('link')).toHaveCount(1)
  await expect(members.getByRole('link', { name: 'member-21' })).toBeVisible()
  await expect(members.getByRole('button', { name: 'Next', exact: true })).toBeDisabled()
  await expect(members.getByText('Page 2', { exact: true })).toBeVisible()
  await members.getByRole('textbox', { name: 'Search recipient usernames' }).fill('member-2')
  await members.getByRole('button', { name: 'Search', exact: true }).click()
  await expect(members.getByRole('link')).toHaveCount(3)
  await expect(members.getByRole('button', { name: 'Next', exact: true })).toBeDisabled()
  await expect(members.getByText('Page 1', { exact: true })).toBeVisible()
  await expect(members.getByText(/^\d+ current recipients$/)).toHaveCount(0)
  expect(requests.at(-1)).toEqual({ page: 1, search: 'member-2' })
})

test('keeps input focus rings inside the scrollport and wraps fixed square icon buttons', async ({ page }, info) => {
  await prepareBadges(page)
  await page.getByRole('button', { name: 'New badge', exact: true }).click()
  const editor = page.getByRole('dialog')
  const form = editor.locator('#badge-form')
  const name = editor.getByLabel('Name', { exact: true })
  const description = editor.getByLabel('Description', { exact: true })
  await name.fill('Community guide')
  await name.focus()
  await expect(name).toBeFocused()
  await expect.poll(() => name.evaluate(element => {
    const field = element.getBoundingClientRect()
    const port = element.closest('form')!.getBoundingClientRect()
    return Math.min(field.left - port.left, port.right - field.right, field.top - port.top, port.bottom - field.bottom)
  })).toBeGreaterThanOrEqual(3)
  await page.screenshot({ path: `/tmp/gooseforum-badge-focus-input-${info.project.name}.png`, fullPage: true })
  await description.fill('Recognition for helping new community members')
  await description.focus()
  await expect(description).toBeFocused()
  await expect.poll(() => description.evaluate(element => {
    const field = element.getBoundingClientRect()
    const port = element.closest('form')!.getBoundingClientRect()
    return Math.min(field.left - port.left, port.right - field.right, field.top - port.top, port.bottom - field.bottom)
  })).toBeGreaterThanOrEqual(3)
  await page.screenshot({ path: `/tmp/gooseforum-badge-focus-textarea-${info.project.name}.png`, fullPage: true })
  const icons = editor.getByRole('group', { name: 'Built-in icons', exact: true })
  await icons.scrollIntoViewIfNeeded()
  const positions = await icons.getByRole('button').evaluateAll(buttons => buttons.map(button => {
    const box = button.getBoundingClientRect()
    return { x: box.x, y: box.y, width: box.width, height: box.height }
  }))
  expect(positions).toHaveLength(15)
  for (const box of positions) { expect(box.width).toBe(48); expect(box.height).toBe(48) }
  const rows = [...new Set(positions.map(box => box.y))]
  expect(rows.length).toBeGreaterThan(1)
  for (const row of rows) expect(positions.find(box => box.y === row)!.x).toBe(positions[0].x)
  await icons.getByRole('button', { name: 'Crown', exact: true }).click()
  const selected = icons.getByRole('button', { name: 'Crown', exact: true })
  const inset = await selected.evaluate(element => {
    const button = element.getBoundingClientRect()
    const check = element.querySelector('svg')!.getBoundingClientRect()
    return { top: check.top - button.top, right: button.right - check.right }
  })
  expect(inset.top).toBeGreaterThanOrEqual(4)
  expect(inset.right).toBeGreaterThanOrEqual(4)
  await page.screenshot({ path: `/tmp/gooseforum-badge-square-icons-${info.project.name}.png`, fullPage: true })
  await editor.getByText('Advanced settings', { exact: true }).click()
  const sort = editor.getByLabel('Sort', { exact: true })
  await form.evaluate(element => { element.scrollTop = element.scrollHeight })
  await sort.focus()
  await expect(sort).toBeFocused()
  await expect.poll(() => sort.evaluate(element => {
    const field = element.getBoundingClientRect()
    const port = element.closest('form')!.getBoundingClientRect()
    return Math.min(field.left - port.left, port.right - field.right, field.top - port.top, port.bottom - field.bottom)
  })).toBeGreaterThanOrEqual(3)
  await page.screenshot({ path: `/tmp/gooseforum-badge-focus-bottom-${info.project.name}.png`, fullPage: true })
  expect(await form.evaluate(element => element.scrollHeight > element.clientHeight)).toBe(true)
  await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true)
})
