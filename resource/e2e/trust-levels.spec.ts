import { expect, test } from '@playwright/test'
import type { TrustProgress, TrustHistory } from '@gooseforum/client'

const layout = {
  site: { name: 'GooseForum', description: '', logo: '', favicon: '', brandType: 'default', brandText: '', brandImage: '' },
  viewer: { id: 7, username: 'admin', email: '', avatarUrl: '', isAuthenticated: true, canAccessAdmin: true, isModerator: false, requiresEmailVerification: false, adminPermissions: [0] },
  header: [], sidebar: { activeKey: 'topics', categories: [] }, footer: { links: [], primary: [] }, unread: { notifications: false, messages: false }, theme: { enabled: false, current: 'gf-light', themeColor: '#fbfdff' },
}

test('trust conditions and image limits save and survive reload', async ({ page }, info) => {
  const errors: string[] = []
  page.on('pageerror', e => errors.push(e.message))
  page.on('console', m => { if (m.type() === 'error') errors.push(m.text()) })
  await page.routeWebSocket(/ws:\/\/127\.0\.0\.1:3011\//, socket => socket.send(JSON.stringify({ type: 'connected' })))
  let settings = {
    schemaVersion: 3,
    registrationCooldown: { minPostAccountAgeMinutes: 0, minUploadAccountAgeMinutes: 1440 },
    promotionRules: [{ level: 100, enabled: true, minAccountAgeHours: 24, minVisitDays: 2, minTopicsVisited: 5 }, { level: 200, enabled: true, minAccountAgeHours: 336, minVisitDays: 7, minTopicsVisited: 20 }],
    levelLimits: [0, 100, 200, 1000].map(level => ({ level, allowImageUpload: true, maxDailyUploads: 10, maxDailyTopics: level === 0 ? 3 : null, maxDailyReplies: level === 0 ? 10 : null, maxImagesPerPost: level === 0 ? 1 : null, maxExternalLinksPerPost: level === 0 ? 2 : null, maxMentionsPerPost: level === 0 ? 2 : null })),
  }
  let saves = 0
  await page.route('**/__goose_page/**', route => route.fulfill({ json: { component: 'admin.shell', props: {}, layout, meta: { title: 'Trust levels' }, url: '/admin/settings/trust?lang=en', version: '1.0' } }))
  await page.route('**/api/admin/get-trust-settings', route => route.fulfill({ json: { code: 0, result: settings } }))
  await page.route('**/api/admin/save-trust-settings', route => { saves++; settings = route.request().postDataJSON().settings; return route.fulfill({ json: { code: 0, result: true } }) })
  await page.route('**/api/admin/security-settings', route => route.fulfill({ json: { code: 0, result: { enableSignup: true, enableEmailVerification: false, allowedDomains: [], registrationIPLimit: 5, registrationEmailLimit: 3, registrationGlobalLimit: 100 } } }))
  await page.route('**/api/admin/posting-settings', route => route.fulfill({ json: { code: 0, result: {} } }))
  let savedPage = ''
  await page.route('**/api/admin/save-security-settings', route => { savedPage = 'security'; return route.fulfill({ json: { code: 0, result: true } }) })
  await page.route('**/api/admin/save-posting-settings', route => { savedPage = 'posting'; return route.fulfill({ json: { code: 0, result: true } }) })
  await page.goto('/admin/settings/trust?lang=en')
  await expect(page.getByLabel('Automatic promotion', { exact: true })).toHaveCount(0)
  await expect(page.getByRole('navigation', { name: 'Community rules', exact: true }).getByRole('link', { name: 'Trust levels', exact: true })).toHaveAttribute('aria-current', 'page')
  await expect(page).toHaveTitle(/Community rules/)
  const primaryMenu = page.locator('[data-slot="sidebar-menu-button"][href="/admin/settings/security"]')
  const localMenu = page.getByRole('navigation', { name: 'Community rules', exact: true }).getByRole('link', { name: 'Trust levels', exact: true })
  const selectedColors = await localMenu.evaluate(element => ({ color: getComputedStyle(element).color, background: getComputedStyle(element).backgroundColor }))
  if (info.project.name === 'mobile-chromium') await page.locator('[data-slot="sidebar-trigger"]').click()
  await expect(primaryMenu).toHaveAttribute('aria-current', 'page')
  await expect(primaryMenu).toHaveCSS('color', selectedColors.color)
  await expect(primaryMenu).toHaveCSS('background-color', selectedColors.background)
  await expect(primaryMenu.locator('svg')).toHaveCSS('color', selectedColors.color)
  if (info.project.name === 'mobile-chromium') await page.keyboard.press('Escape')
  await expect(localMenu.locator('svg')).toHaveCSS('color', selectedColors.color)
  await page.getByLabel('Account age (hours)').first().fill('48')
  await page.getByRole('tab', { name: 'Limits', exact: true }).click()
  await page.getByLabel('Images per post', { exact: true }).first().fill('3')
  await page.getByLabel('Daily new topics (including drafts)', { exact: true }).first().fill('5')
  await page.getByLabel('Daily new replies', { exact: true }).first().fill('12')
  await page.getByLabel('Daily uploads', { exact: true }).first().fill('7')
  await page.getByLabel('Daily uploads', { exact: true }).nth(1).fill('6')
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await expect(page.getByRole('alert')).toContainText('Member: daily uploads cannot be stricter than New user.')
  expect(saves).toBe(0)
  await page.getByLabel('Daily uploads', { exact: true }).nth(1).fill('10')
  const unlimited = page.getByRole('checkbox', { name: 'Daily new replies unlimited', exact: true }).nth(1)
  await expect(unlimited).toBeEnabled()
  await expect(unlimited).toHaveCSS('opacity', '1')
  await unlimited.uncheck()
  await expect(page.getByLabel('Daily new replies', { exact: true }).nth(1)).toBeEnabled()
  await unlimited.check()
  await expect(page.getByLabel('Daily new replies', { exact: true }).nth(1)).toBeDisabled()
  const titleGap = await page.getByLabel('Daily new replies', { exact: true }).first().evaluate(input => {
    const label = document.querySelector(`label[for="${input.id}"]`)!
    return input.getBoundingClientRect().top - label.getBoundingClientRect().bottom
  })
  expect(titleGap).toBeLessThanOrEqual(10)
  await page.getByRole('tab', { name: 'Registration cooldown', exact: true }).click()
  await expect(page.getByLabel('Posting cooldown (minutes)', { exact: true })).toHaveCount(1)
  await page.getByLabel('Posting cooldown (minutes)', { exact: true }).fill('30')
  await page.getByLabel('Upload cooldown (minutes)', { exact: true }).fill('60')
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await expect.poll(() => settings.levelLimits[0].maxImagesPerPost).toBe(3)
  expect(settings.promotionRules[0].minAccountAgeHours).toBe(48)
  expect(settings.levelLimits[0].maxDailyTopics).toBe(5)
  expect(settings.levelLimits[0].maxDailyReplies).toBe(12)
  expect(settings.levelLimits[0].maxDailyUploads).toBe(7)
  expect(settings.registrationCooldown.minPostAccountAgeMinutes).toBe(30)
  expect(settings.registrationCooldown.minUploadAccountAgeMinutes).toBe(60)
  expect(settings).not.toHaveProperty('restrictionsEnabled')
  expect(settings).not.toHaveProperty('autoPromotionEnabled')
  await page.reload()
  await page.getByRole('tab', { name: 'Limits', exact: true }).click()
  await expect(page.getByLabel('Images per post', { exact: true }).first()).toHaveValue('3')
  await expect(page.getByLabel('Daily uploads', { exact: true }).first()).toHaveValue('7')
  await page.getByRole('tab', { name: 'Registration cooldown', exact: true }).click()
  await expect(page.getByLabel('Posting cooldown (minutes)', { exact: true })).toHaveValue('30')
  await expect(page.getByLabel('Upload cooldown (minutes)', { exact: true })).toHaveValue('60')
  await page.getByRole('tab', { name: 'Limits', exact: true }).click()
  expect(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth + 1)).toBe(false)
  await page.screenshot({ path: `/tmp/gooseforum-trust-${info.project.name}.png`, fullPage: true })
  const nav = page.getByRole('navigation', { name: 'Community rules', exact: true })
  await expect(page.locator('main h2')).toHaveCount(0)
  await nav.getByRole('link', { name: 'Security', exact: true }).click()
  await expect(page).toHaveURL(/\/admin\/settings\/security/)
  await expect(nav.getByRole('link', { name: 'Security', exact: true })).toHaveAttribute('aria-current', 'page')
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await expect.poll(() => savedPage).toBe('security')
  await nav.getByRole('link', { name: 'Posting', exact: true }).click()
  await expect(page).toHaveURL(/\/admin\/settings\/posting/)
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await expect.poll(() => savedPage).toBe('posting')
  expect(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth + 1)).toBe(false)
  await page.screenshot({ path: `/tmp/gooseforum-posting-columns-${info.project.name}.png`, fullPage: true })
  expect(errors).toEqual([])
})

test('administrator promotes and locks a user with history', async ({ page }, info) => {
  const errors: string[] = []
  page.on('pageerror', e => errors.push(e.message))
  await page.routeWebSocket(/ws:\/\/127\.0\.0\.1:3011\//, socket => socket.send(JSON.stringify({ type: 'connected' })))
  const user = { userId: 8, username: 'alice', email: 'alice@example.test', validate: 1, prestige: 0, roleId: 0, roleList: [], createTime: '2026-10-01', trustLevel: 100 }
  let progress: TrustProgress = { level: 100, name: '成员', locked: false, version: 2, visitDays: 3, topicsVisited: 10, createdAt: '2026-10-01T00:00:00Z', next: null, definitions: [] }
  const history: TrustHistory[] = []
  await page.route('**/__goose_page/**', route => route.fulfill({ json: { component: 'admin.shell', props: {}, layout, meta: { title: 'Users' }, url: '/admin/users?lang=en', version: '1.0' } }))
  await page.route('**/api/admin/**', route => {
    const path = new URL(route.request().url()).pathname
    let result: unknown = []
    if (path.endsWith('/user-list')) result = { list: [user], page: 1, pageSize: 10, total: 1 }
    if (path.endsWith('/user-badge-options')) result = { options: [], active: [] }
    if (path.endsWith('/user-mfa-status')) result = { enabled: false }
    if (path.endsWith('/user-trust-level')) result = progress
    if (path.endsWith('/user-trust-level-history')) result = history
    if (path.endsWith('/set-user-trust-level')) {
      const body = route.request().postDataJSON()
      expect(body.expectedVersion).toBe(2)
      history.push({ id: 1, userId: 8, fromLevel: progress.level, toLevel: body.level, fromLocked: false, toLocked: body.locked, source: 'admin', operatorId: 7, reason: body.reason, createdAt: '2026-10-16T00:00:00Z' })
      progress = { ...progress, level: body.level, locked: body.locked, version: 3 }
      result = progress
    }
    return route.fulfill({ json: { code: 0, result } })
  })
  await page.goto('/admin/users?lang=en')
  await page.getByRole('button', { name: 'Edit user', exact: true }).filter({ visible: true }).first().click()
  const dialog = page.getByRole('dialog', { name: 'Edit user', exact: true })
  await dialog.getByRole('tab', { name: 'Trust level', exact: true }).click()
  await dialog.getByLabel('Level', { exact: true }).click()
  await page.getByRole('option', { name: 'Community leader', exact: true }).click()
  await dialog.getByRole('switch', { name: 'Lock automatic promotion' }).check()
  await dialog.getByLabel('Reason', { exact: true }).fill('Community appointment')
  await dialog.getByRole('button', { name: 'Save level', exact: true }).click()
  await expect(dialog.getByText('Community appointment', { exact: false })).toBeVisible()
  expect(progress.level).toBe(1000)
  expect(progress.locked).toBe(true)
  expect(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth + 1)).toBe(false)
  await page.screenshot({ path: `/tmp/gooseforum-trust-user-${info.project.name}.png` })
  expect(errors).toEqual([])
})

test('promotion notification opens personal permissions and upgrade comparison', async ({ page }, info) => {
  const errors: string[] = []
  page.on('pageerror', error => errors.push(error.message))
  await page.routeWebSocket(/ws:\/\/127\.0\.0\.1:3011\//, socket => socket.send(JSON.stringify({ type: 'connected' })))
  const limits = { level: 100, allowImageUpload: true, maxDailyTopics: 3, maxDailyReplies: 10, maxDailyUploads: 10, maxImagesPerPost: 1, maxExternalLinksPerPost: 2, maxMentionsPerPost: 2 }
  const viewer = { ...layout.viewer, id: 7, username: 'alice', isAuthenticated: true }
  const siteLayout = { ...layout, viewer }
  const settings = {
    component: 'settings.index', layout: siteLayout, meta: { title: 'Account' }, url: '/settings?tab=account&lang=en', version: '1.0',
    props: {
      user: { id: 7, username: 'alice', nickname: 'Alice', email: 'alice@example.test', locale: 'en', avatarUrl: '', profileCoverUrl: '', bio: '', signature: '', websiteName: '', website: '', prestige: 0, createdAt: '2026-01-01T00:00:00Z', externalInformation: {}, wornBadgeCode: '', badges: [], wearableBadges: [] },
      stats: { topicCount: 0, replyCount: 0, followerCount: 0, followingCount: 0, likeReceivedCount: 0, likeGivenCount: 0, collectionCount: 0, createdAt: '2026-01-01T00:00:00Z' },
      privacy: { showTopics: true, showActivity: true, showFollowing: true }, tabs: [{ key: 'account', url: '/settings?tab=account', active: true }],
    },
  }
  const notifications = { component: 'notifications.index', layout: siteLayout, meta: { title: 'Notifications' }, url: '/notifications?lang=en', version: '1.0', props: {
    total: 1, unreadCount: 1, pagination: { page: 1, nextPage: 2, hasNext: false, nextUrl: '' },
    notifications: [{ id: 11, eventType: 'system', isRead: false, createdAt: '2026-10-20T00:00:00Z', title: '', content: '', actor: { id: 0, username: '' }, payload: { actorId: 0, templateKey: 'notifications.templates.trustPromotion', templateParams: { trustLevel: 100 } } }],
  } }
  await page.route('**/__goose_page/**', route => route.fulfill({ json: route.request().url().includes('/settings') ? settings : notifications }))
  let marked = 0
  await page.route('**/api/forum/notification/mark-read', route => { marked = route.request().postDataJSON().notificationId; return route.fulfill({ json: { code: 0, result: true } }) })
  await page.route('**/api/mfa', route => route.fulfill({ json: { code: 0, result: { enabled: false, available: true, remainingCodes: 0 } } }))
  await page.route('**/api/trust-level', route => route.fulfill({ json: { code: 0, result: { level: 100, name: '成员', locked: false, version: 0, visitDays: 3, topicsVisited: 10, createdAt: '2026-01-01T00:00:00Z', next: { level: 200, enabled: true, minVisitDays: 7, minTopicsVisited: 20, minAccountAgeHours: 336 }, definitions: [], limits, nextLimits: { ...limits, level: 200, maxDailyTopics: 10, maxDailyReplies: null, maxImagesPerPost: null }, registrationCooldown: { minPostAccountAgeMinutes: 0, minUploadAccountAgeMinutes: 1440 }, communityUploadsEnabled: true, communityUploadRoleExempt: false, restrictionStatus: 'normal' } } }))
  await page.goto('/notifications?lang=en')
  const promotion = page.getByRole('link', { name: 'Promoted to Member. View your permissions and limits' })
  await expect(promotion).toHaveAttribute('href', '/settings?tab=account')
  await promotion.click()
  await expect(page).toHaveURL(/\/settings\?tab=account/)
  await expect.poll(() => marked).toBe(11)
  await expect(page.getByRole('heading', { name: 'Current level permissions and limits' })).toBeVisible()
  const reply = page.getByRole('row', { name: /Daily replies/ })
  await expect(reply).toContainText('10')
  await expect(reply).toContainText('Unlimited')
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true)
  await page.screenshot({ path: `/tmp/gooseforum-trust-progress-${info.project.name}.png`, fullPage: true })
  expect(errors).toEqual([])
})
