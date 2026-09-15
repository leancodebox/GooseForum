import AxeBuilder from '@axe-core/playwright'
import { expect, test } from '@playwright/test'

const homePage = {
  component: 'home.index',
  props: {
    sort: 'latest',
    tabs: [{ key: 'latest', label: 'Latest', url: '/', active: true }],
    topics: [],
    pagination: { page: 1, nextPage: 2, hasNext: false, nextUrl: '' },
    announcement: { enabled: false, html: '' },
  },
  layout: {
    site: {
      name: 'GooseForum',
      description: 'Quality smoke test',
      logo: '',
      favicon: '',
      brandType: 'default',
      brandText: '',
      brandImage: '',
    },
    viewer: {
      id: 0,
      username: '',
      email: '',
      avatarUrl: '',
      isAuthenticated: false,
      canAccessAdmin: false,
      isModerator: false,
      requiresEmailVerification: false,
      adminPermissions: [],
    },
    header: [],
    sidebar: { activeKey: 'topics', categories: [] },
    footer: { links: [], primary: [] },
    unread: { notifications: false, messages: false },
    theme: { enabled: false, current: 'gf-light', themeColor: '#ffffff' },
  },
  meta: { title: 'GooseForum quality smoke test' },
  url: '/',
  version: '1.0',
}

const adminPage = {
  ...homePage,
  component: 'admin.shell',
  props: {},
  layout: {
    ...homePage.layout,
    viewer: {
      ...homePage.layout.viewer,
      id: 1,
      username: 'admin',
      isAuthenticated: true,
      canAccessAdmin: true,
      isModerator: true,
      adminPermissions: [0],
    },
  },
  meta: { title: 'GooseForum admin quality smoke test' },
  url: '/admin',
}

const topicPage = {
  ...homePage,
  component: 'topic.detail',
  props: {
    topic: {
      id: 60,
      title: 'Topic detail',
      description: 'Description',
      url: '/p/test/60',
      topicStatus: 1,
      processStatus: 0,
      author: { id: 7, username: 'alice', avatarUrl: '' },
      participants: [{ id: 7, username: 'alice', avatarUrl: '' }],
      categories: [],
      replyCount: 0,
      maxPostNo: 1,
      viewCount: 10,
      likeCount: 0,
      isLiked: false,
      isBookmarked: false,
      isWatched: false,
      createdAt: '2026-09-14T08:00:00Z',
      updatedAt: '2026-09-14T08:00:00Z',
    },
    postStream: {
      posts: [{
        id: 61,
        topicId: 60,
        postNo: 1,
        content: 'Original body',
        renderedContent: '<p>Original body</p>',
        processStatus: 0,
        isHidden: false,
        canModerate: false,
        author: { id: 7, username: 'alice', avatarUrl: '' },
        createdAt: '2026-09-14T08:00:00Z',
        isOwnPost: true,
      }],
      replyTargets: [],
      beforePostNo: 1,
      afterPostNo: 1,
      hasBefore: false,
      hasAfter: false,
      total: 1,
      maxPostNo: 1,
    },
    hotTopics: [],
    permissions: { isOwnTopic: true, canPost: true, canModerateTopic: false },
  },
  layout: {
    ...homePage.layout,
    viewer: {
      ...homePage.layout.viewer,
      id: 7,
      username: 'alice',
      isAuthenticated: true,
    },
  },
  meta: { title: 'Topic detail' },
  url: '/p/test/60',
}

const messagesPage = {
  ...homePage,
  component: 'messages.index',
  props: {
    conversations: [{
      id: 4,
      peerId: 9,
      peerUsername: 'bob',
      peerAvatar: '',
      lastMsg: 'OK',
      lastMsgTime: '2026-09-16T08:03:00Z',
      unreadCount: 0,
      convId: 4,
      peerUrl: '/u/9',
    }],
    suggestedUsers: [],
  },
  layout: {
    ...homePage.layout,
    viewer: {
      ...homePage.layout.viewer,
      id: 7,
      username: 'alice',
      isAuthenticated: true,
    },
  },
  meta: { title: 'Messages' },
  url: '/messages?userId=9',
}

const chatMessages = {
  list: [1, 2, 3, 4].map(id => ({
    id,
    senderId: 9,
    content: id === 1 ? 'An older, longer message remains visible.' : 'OK',
    msgType: 1,
    isRead: 1,
    createdAt: `2026-09-16T08:0${id}:00Z`,
    isSelf: false,
  })),
  hasMoreBefore: false,
  hasMoreAfter: false,
  nextBeforeId: 0,
  latestId: 4,
}

test.beforeEach(async ({ page }) => {
  await page.route('**/__goose_page/**', async route => {
    const url = route.request().url()
    await route.fulfill({
      json: url.includes('/admin')
        ? adminPage
        : url.includes('/messages')
          ? messagesPage
        : url.includes('/p/test/60')
          ? topicPage
          : homePage,
    })
  })
  await page.route(url => url.pathname.startsWith('/api/'), async route => {
    if (route.request().url().includes('/api/admin/announcement')) {
      await route.fulfill({
        json: {
          code: 0,
          result: {
            enabled: true,
            content: '## Maintenance\n\nThe forum will be read-only tonight.',
          },
        },
      })
      return
    }
    if (route.request().url().includes('/api/admin/save-announcement')) {
      await route.fulfill({ json: { code: 0, result: true } })
      return
    }
    if (route.request().url().includes('/api/forum/chat/messages')) {
      await route.fulfill({ json: { code: 0, result: chatMessages } })
      return
    }
    await route.fulfill({ json: [] })
  })
})

test('loads, responds to a primary control, and meets automated WCAG checks', async ({ page }, testInfo) => {
  const errors: string[] = []
  page.on('pageerror', error => errors.push(error.message))
  page.on('console', message => {
    if (message.type() === 'error') errors.push(message.text())
  })

  await page.goto('/?lang=en')
  await expect(page).toHaveTitle('GooseForum quality smoke test')
  await expect(page.getByText('No topics yet')).toBeVisible()

  const themeButton = page.getByRole('button', { name: 'Switch to dark theme' })
  await themeButton.click()
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'gf-dark')
  await expect(page.getByRole('button', { name: 'Switch to light theme' })).toBeVisible()

  const accessibility = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
    .analyze()
  expect(accessibility.violations).toEqual([])
  expect(errors).toEqual([])
  await testInfo.attach('verified-page', {
    body: await page.screenshot({ fullPage: false }),
    contentType: 'image/png',
  })
})

test('loads the admin dashboard shell and deferred chart without accessibility regressions', async ({ page }, testInfo) => {
  const errors: string[] = []
  page.on('pageerror', error => errors.push(error.message))
  page.on('console', message => {
    if (message.type() === 'error') errors.push(message.text())
  })

  await page.goto('/admin?lang=en')
  await expect(page).toHaveTitle('Dashboard - GooseForum')
  await expect(page.getByRole('heading', { level: 2, name: 'Dashboard' })).toBeVisible()
  await expect(page.getByText('Traffic overview')).toBeVisible()

  const accessibility = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
    .analyze()
  expect(accessibility.violations).toEqual([])
  expect(errors).toEqual([])
  await testInfo.attach('verified-admin-page', {
    body: await page.screenshot({ fullPage: false }),
    contentType: 'image/png',
  })
})

test('cold-loads the badges route after preparing its assets translations', async ({ page }) => {
  const errors: string[] = []
  page.on('pageerror', error => errors.push(error.message))
  page.on('console', message => {
    if (message.type() === 'error') errors.push(message.text())
  })

  await page.goto('/admin/badges?lang=en')
  await expect(page).toHaveTitle('Badges - GooseForum')
  await expect(page.getByRole('heading', { level: 2, name: 'Badges' })).toBeVisible()
  await expect(page.getByText('Manage system and custom badges.')).toBeVisible()
  expect(errors).toEqual([])
})

test('keeps the topic reply control inside the content edge and shares one composer action row', async ({ page }, testInfo) => {
  const errors: string[] = []
  page.on('pageerror', error => errors.push(error.message))
  page.on('console', message => {
    if (message.type() === 'error') errors.push(message.text())
  })

  await page.goto('/p/test/60?lang=en')
  await expect(page).toHaveTitle('Topic detail')
  await expect(page.getByRole('heading', { level: 1, name: 'Topic detail' })).toBeVisible()

  const boundary = page.locator('[data-slot="topic-reply-float-boundary"]')
  const reply = page.getByRole('button', { name: 'Join the discussion' })
  const [boundaryBox, replyBox] = await Promise.all([boundary.boundingBox(), reply.boundingBox()])
  expect(boundaryBox).not.toBeNull()
  expect(replyBox).not.toBeNull()
  const rightGap = boundaryBox!.x + boundaryBox!.width - replyBox!.x - replyBox!.width
  expect(rightGap).toBeGreaterThanOrEqual(23)

  await reply.click()
  const toolbar = page.locator('[data-slot="markdown-composer-toolbar"]')
  await expect(toolbar.getByRole('button', { name: 'Post reply' })).toBeVisible()
  await expect(toolbar.getByRole('radio', { name: 'Markdown' })).toBeVisible()
  await expect(toolbar.getByRole('button', { name: 'Preview' })).toBeVisible()
  expect(errors).toEqual([])
  await testInfo.attach('verified-topic-composer', {
    body: await page.screenshot({ fullPage: false }),
    contentType: 'image/png',
  })
})

test('keeps short-message avatars inside every message row', async ({ page }, testInfo) => {
  const errors: string[] = []
  page.on('pageerror', error => errors.push(error.message))
  page.on('console', message => {
    if (message.type() === 'error') errors.push(message.text())
  })

  await page.goto('/messages?userId=9&lang=en')
  await expect(page).toHaveTitle('Messages')
  const items = page.locator('[data-slot="message-scroller-item"]')
  await expect(items).toHaveCount(4)
  const itemCount = await items.count()
  for (let index = 0; index < itemCount; index += 1) {
    const item = items.nth(index)
    const avatar = item.locator('[data-slot="message-avatar"]')
    await expect(avatar).toBeVisible()
    const [itemBox, avatarBox] = await Promise.all([
      item.boundingBox(),
      avatar.boundingBox(),
    ])
    expect(itemBox).not.toBeNull()
    expect(avatarBox).not.toBeNull()
    expect(avatarBox!.y).toBeGreaterThanOrEqual(itemBox!.y)
  }
  expect(errors).toEqual([])
  await testInfo.attach('verified-message-avatars', {
    body: await page.screenshot({ fullPage: false }),
    contentType: 'image/png',
  })
})

test('edits, previews, and saves an announcement in Markdown', async ({ page }, testInfo) => {
  const errors: string[] = []
  page.on('pageerror', error => errors.push(error.message))
  page.on('console', message => {
    if (message.type() === 'error') errors.push(message.text())
  })

  await page.goto('/admin/settings/announcement?lang=en')
  await expect(page).toHaveTitle('Announcement - GooseForum')
  const editor = page.getByRole('textbox', { name: 'Announcement content' })
  await expect(editor).toHaveValue(/Maintenance/)
  await editor.fill('## Updated announcement\n\n**Everything is ready.**')
  await page.getByRole('radio', { name: 'Preview' }).click()
  const preview = page.locator('[data-slot="announcement-markdown-preview"]')
  await expect(preview.getByRole('heading', { name: 'Updated announcement' })).toBeVisible()
  await expect(preview.getByText('Everything is ready.')).toBeVisible()

  const saveRequest = page.waitForRequest(request =>
    request.url().includes('/api/admin/save-announcement'),
  )
  await page.getByRole('button', { name: 'Save' }).click()
  expect((await saveRequest).postDataJSON()).toMatchObject({
    settings: { enabled: true, content: '## Updated announcement\n\n**Everything is ready.**' },
  })
  expect(errors).toEqual([])
  await testInfo.attach('verified-announcement-markdown-editor', {
    body: await page.screenshot({ fullPage: false }),
    contentType: 'image/png',
  })
})
