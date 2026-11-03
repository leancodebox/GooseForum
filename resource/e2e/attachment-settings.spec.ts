import { expect, test } from '@playwright/test'

test('attachment types use separate inputs and persist after reload', async ({ page }, info) => {
  const errors: string[] = []
  page.on('pageerror', error => errors.push(error.message))
  page.on('console', message => { if (message.type() === 'error') errors.push(message.text()) })
  await page.route('**/__goose_page/**', route => route.fulfill({ json: {
    component: 'admin.shell', props: {}, url: '/admin/settings/posting?lang=en', version: '1.0', meta: { title: 'Posting settings' },
    layout: { site: { name: 'GooseForum', description: '', logo: '', favicon: '', brandType: 'default', brandText: '', brandImage: '' }, viewer: { id: 1, username: 'admin', email: '', avatarUrl: '', isAuthenticated: true, canAccessAdmin: true, isModerator: false, requiresEmailVerification: false, adminPermissions: [0] }, header: [], sidebar: { activeKey: 'topics', categories: [] }, footer: { links: [], primary: [] }, unread: { notifications: false, messages: false }, theme: { enabled: false, current: 'gf-light', themeColor: '#fbfdff' } },
  } }))
  let settings: Record<string, unknown> = {}
  await page.route('**/api/admin/posting-settings', route => route.fulfill({ json: { code: 0, result: settings } }))
  await page.route('**/api/admin/save-posting-settings', route => {
    settings = route.request().postDataJSON().settings
    return route.fulfill({ json: { code: 0, result: true } })
  })
  await page.goto('/admin/settings/posting?lang=en')
  await expect(page).toHaveTitle(/Community rules/)
  const section = page.locator('fieldset').filter({ has: page.getByText('File attachments', { exact: true }) })
  await expect(section.getByRole('spinbutton')).toHaveCount(2)
  await section.getByRole('textbox', { name: 'Allowed file types 1', exact: true }).fill(' .PDF ')
  await section.getByRole('button', { name: 'Remove .zip', exact: true }).click()
  await section.getByRole('button', { name: 'Add', exact: true }).click()
  await section.getByRole('textbox', { name: 'Allowed file types 3', exact: true }).fill('.zip')
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await expect.poll(() => settings.attachments).toMatchObject({ extensions: ['.pdf', '.txt', '.zip'] })
  await page.reload()
  await expect(section.getByRole('textbox', { name: 'Allowed file types 1', exact: true })).toHaveValue('.pdf')
  await expect(section.getByRole('textbox')).toHaveCount(3)
  await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true)
  await page.screenshot({ path: `/tmp/gooseforum-attachment-settings-${info.project.name}.png`, fullPage: true })
  expect(errors).toEqual([])
})
