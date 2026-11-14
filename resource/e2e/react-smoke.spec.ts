import AxeBuilder from '@axe-core/playwright'
import { expect, test, type Page } from '@playwright/test'
import { fileURLToPath } from 'node:url'

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
    theme: { enabled: false, current: 'gf-light', themeColor: '#fbfdff' },
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

const publishPage = {
  ...homePage,
  component: 'publish.index',
  props: {
    topicId: 0,
    isEditing: false,
    categories: [{
      id: 4,
      name: 'Coding',
      color: '#8241d6',
      isRestricted: false,
      canCreate: true,
    }],
    topic: { title: '', content: '', categoryIds: [], topicStatus: 0 },
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
  meta: { title: 'Publish' },
  url: '/publish',
}

const publishedTopic = {
  id: 99,
  title: 'Freshly published topic',
  description: 'Published body',
  url: '/p/post/99',
  author: { id: 7, username: 'alice', avatarUrl: '' },
  participants: [{ id: 7, username: 'alice', avatarUrl: '' }],
  categories: [{ id: 4, name: 'Coding', url: '/c/Coding/4', color: '#8241d6' }],
  replyCount: 0,
  viewCount: 0,
  pinWeight: 0,
  processStatus: 0,
  activityText: '',
  lastUpdateTime: '2026-09-17T00:00:00Z',
}

const responsiveListBadge = {
  code: 'responsive-list-badge',
  type: 'system',
  grantMode: 'manual',
  name: 'Responsive list badge',
  description: 'Available to themes but hidden by the default topic list',
  iconType: 'image',
  iconKey: 'responsive-list-badge',
  iconUrl: '/responsive-list-badge.svg',
  color: 'blue',
  level: 'special',
  isEnabled: true,
  isWearable: true,
  sortOrder: 1,
  source: 'manual',
  reason: '',
  grantedAt: '2026-09-17T00:00:00Z',
}

const responsiveTopic = {
  ...publishedTopic,
  id: 101,
  title: 'How should a long topic title adapt cleanly across a compact mobile forum list?',
  description: 'The desktop summary remains unchanged.',
  url: '/p/test/60',
  author: { id: 12, username: 'responsive-author', avatarUrl: '', wornBadge: responsiveListBadge },
  participants: [{ id: 12, username: 'responsive-author', avatarUrl: '', wornBadge: responsiveListBadge }],
  categories: [
    { id: 4, name: 'Coding', url: '/c/Coding/4', color: '#8241d6' },
    { id: 5, name: 'Frontend', url: '/c/Frontend/5', color: '#0ea5e9' },
  ],
  replyCount: 128,
  viewCount: 6800,
  pinWeight: 10,
  unseen: true,
}

const compactResponsiveTopic = {
  ...responsiveTopic,
  id: 102,
  title: 'A short mobile topic',
  url: '/p/responsive/102',
  categories: [
    { id: 6, name: 'General', url: '/c/General/6', color: '#22c55e' },
  ],
  replyCount: 0,
  viewCount: 20,
  pinWeight: 0,
  unseen: false,
}

const responsiveTopics = [
  responsiveTopic,
  compactResponsiveTopic,
  ...Array.from({ length: 12 }, (_, index) => ({
    ...compactResponsiveTopic,
    id: 200 + index,
    title: `A later page topic ${index + 1}`,
    url: '/p/test/60?transition=slow',
  })),
]

const publishedTopicPage = {
  ...topicPage,
  props: {
    ...topicPage.props,
    topic: {
      ...topicPage.props.topic,
      ...publishedTopic,
      topicStatus: 1,
      maxPostNo: 1,
      likeCount: 0,
      isLiked: false,
      isBookmarked: false,
      isWatched: false,
      createdAt: '2026-09-17T00:00:00Z',
      updatedAt: '2026-09-17T00:00:00Z',
    },
    postStream: {
      ...topicPage.props.postStream,
      posts: [{
        ...topicPage.props.postStream.posts[0],
        id: 100,
        topicId: 99,
        content: 'Published body',
        renderedContent: '<p>Published body</p>',
      }],
    },
  },
  meta: { title: 'Freshly published topic' },
  url: '/p/post/99',
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
  let published = false
  // The browser suite starts Vite without Go; serve the real image fixtures locally.
  await page.route('**/static/**', route => route.fulfill({
    path: fileURLToPath(new URL(`../${new URL(route.request().url()).pathname.slice(1)}`, import.meta.url)),
  }))
  await page.route('**/__goose_page/**', async route => {
    const url = route.request().url()
    if (url.includes('/p/test/60') && url.includes('transition=slow')) {
      await new Promise(resolve => setTimeout(resolve, 250))
    }
    await route.fulfill({
      json: url.includes('/admin')
        ? adminPage
        : url.includes('/messages')
          ? messagesPage
        : url.includes('/publish')
          ? publishPage
        : url.includes('/p/post/99')
          ? publishedTopicPage
        : url.includes('/p/test/60')
          ? topicPage
        : url.includes('responsive=topics')
          ? {
              ...homePage,
              props: {
                ...homePage.props,
                topics: responsiveTopics,
              },
            }
          : published
            ? {
                ...homePage,
                props: { ...homePage.props, topics: [publishedTopic] },
              }
            : homePage,
    })
  })
  await page.route(url => url.pathname.startsWith('/api/'), async route => {
    if (route.request().url().includes('/api/forum/topics/watch')) {
      await route.fulfill({ json: { code: 0, result: true } })
      return
    }
    if (route.request().url().includes('/api/user-card')) {
      await route.fulfill({
        json: {
          code: 0,
          result: {
            userId: 12,
            username: 'responsive-author',
            nickname: 'Responsive Author',
            avatarUrl: '',
            profileCoverUrl: '',
            bio: '',
            signature: '',
            websiteName: '',
            website: '',
            prestige: 0,
            externalInformation: {},
            isAdmin: false,
            topicCount: 2,
            replyCount: 4,
            likeReceivedCount: 3,
            likeGivenCount: 1,
            followerCount: 5,
            followingCount: 6,
            collectionCount: 0,
            isOnline: true,
            isFollowing: false,
            isSelf: false,
            badges: [],
            lastActiveTime: '2026-09-17T00:00:00Z',
            createdAt: '2026-01-01T00:00:00Z',
          },
        },
      })
      return
    }
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
    if (route.request().url().includes('/api/admin/oauth-settings')) {
      await route.fulfill({
        json: {
          code: 0,
          result: {
            siteUrl: 'https://forum.example/base/',
            providers: [{
              key: 'company-sso',
              displayName: 'Company SSO',
              kind: 'oidc',
              enabled: true,
              clientId: 'company-client',
              clientSecretConfigured: true,
              callbackUrl: 'https://forum.example/base/api/auth/company-sso/callback',
              discoveryUrl: 'https://identity.example/.well-known/openid-configuration',
              scopes: ['openid', 'profile', 'email'],
            }],
          },
        },
      })
      return
    }
    if (route.request().url().includes('/api/forum/chat/messages')) {
      const request = route.request().postDataJSON() as { afterId?: number }
      const result = request.afterId
        ? { ...chatMessages, list: [{ ...chatMessages.list[0], id: 5, content: 'A new message arrived.', createdAt: '2026-09-16T08:05:00Z' }], latestId: 5 }
        : chatMessages
      await route.fulfill({ json: { code: 0, result } })
      return
    }
    if (route.request().url().includes('/api/forum/topics/write')) {
      const input = route.request().postDataJSON() as { title?: string }
      if ((input.title || '').length < 3) {
        await route.fulfill({
          json: {
            code: 1,
            messageCode: 'topic.title.tooShort',
            params: { minLength: 3 },
          },
        })
        return
      }
      published = true
      await route.fulfill({
        json: { code: 0, result: { id: 99, moderationStatus: 'approved' } },
      })
      return
    }
    await route.fulfill({ json: [] })
  })
})

test('presents the OAuth callback as explained copyable segments', async ({ page }, testInfo) => {
  await page.goto('/admin/settings/oauth?lang=zh')
  await expect(page.getByRole('navigation', { name: 'OAuth / OIDC', exact: true }).getByRole('link', { name: 'OAuth 第三方登录', exact: true })).toHaveAttribute('aria-current', 'page')

  const callback = page.getByLabel('回调地址', { exact: true })
  await expect(callback).toHaveText('https://forum.example/base/api/auth/company-sso/callback')
  await page.getByText('/api/auth/', { exact: true }).hover()
  await expect(page.getByRole('tooltip')).toHaveText('GooseForum 固定的 OAuth 回调路径。')
  await expect(page.getByRole('button', { name: '复制完整回调地址' })).toBeEnabled()

  const callbackBox = await callback.boundingBox()
  const viewport = page.viewportSize()
  if (!callbackBox || !viewport) throw new Error('OAuth callback layout was not measurable')
  expect(callbackBox.x).toBeGreaterThanOrEqual(0)
  expect(callbackBox.x + callbackBox.width).toBeLessThanOrEqual(viewport.width)

  await page.getByRole('button', { name: '删除提供方' }).click()
  await expect(page.getByRole('dialog')).toContainText('删除这个 OAuth 提供方？')
  await page.getByRole('button', { name: '取消' }).click()
  await expect(page.getByRole('dialog')).toHaveCount(0)

  await testInfo.attach('oauth-callback-segments', {
    body: await page.screenshot({ fullPage: false }),
    contentType: 'image/png',
  })
})

test('shows a newly published topic when returning to the cached latest list', async ({ page }, testInfo) => {
  await page.goto('/?lang=en')
  await expect(page.getByText('No topics yet')).toBeVisible()
  await page.getByRole('link', { name: 'New topic' }).click()
  await page.getByPlaceholder('Enter topic title').fill('Freshly published topic')
  await page.getByRole('button', { name: 'Coding' }).click()
  await page.getByRole('textbox', { name: 'Write and format the body directly' }).fill('Published body')
  await page.getByRole('button', { name: 'Publish topic' }).click()
  await expect(page.getByRole('heading', { level: 1, name: 'Freshly published topic' })).toBeVisible()

  await page.getByRole('link', { name: 'GooseForum' }).click()

  await expect(page).toHaveURL(/\/$/)
  await expect(page.getByRole('link', { name: 'Freshly published topic' })).toBeVisible()
  await testInfo.attach('published-topic-on-latest-list', {
    body: await page.screenshot({ fullPage: false }),
    contentType: 'image/png',
  })
})

test('uses the compact topic information hierarchy only on mobile', async ({ page }, testInfo) => {
  const errors: string[] = []
  page.on('pageerror', error => errors.push(error.message))
  page.on('console', message => {
    if (message.type() === 'error') errors.push(message.text())
  })

  await page.goto('/?responsive=topics&lang=en')
  await expect(page).toHaveTitle('GooseForum quality smoke test')
  await expect(page.locator('vite-error-overlay')).toHaveCount(0)

  const rows = page.locator('[data-slot="topic-row"]')
  const row = rows.first()
  const compactRow = rows.nth(1)
  const mobileAvatar = row.locator('[data-slot="topic-row-mobile-avatar"]')
  const title = row.locator('[data-slot="topic-row-title"]')
  const pin = title.locator('[data-slot="topic-row-pin"]')
  const replies = row.locator('[data-slot="topic-row-mobile-replies"]')
  const metadata = row.locator('[data-slot="topic-row-mobile-meta"]')
  const activity = row.locator('time')

  await expect(title).toBeVisible()
  await expect(pin).toBeVisible()
  await expect(page.getByAltText('Responsive list badge')).toHaveCount(0)
  if ((page.viewportSize()?.width || 0) >= 1024) {
    await expect(mobileAvatar).toBeHidden()
    await expect(page.getByRole('columnheader', { name: 'Replies' })).toBeVisible()
    expect(errors).toEqual([])
    return
  }

  await expect(mobileAvatar).toBeVisible()
  await expect(replies).toHaveText('128')
  await expect(metadata).toContainText('Coding')
  await expect(metadata).not.toContainText('responsive-author')
  await expect(activity).toBeVisible()

  const [rowBox, avatarBox, titleBox, repliesBox, metadataBox, activityBox] = await Promise.all([
    row.boundingBox(),
    mobileAvatar.boundingBox(),
    title.boundingBox(),
    replies.boundingBox(),
    metadata.boundingBox(),
    activity.boundingBox(),
  ])
  if (!rowBox || !avatarBox || !titleBox || !repliesBox || !metadataBox || !activityBox) {
    throw new Error('The responsive topic row did not produce measurable layout boxes')
  }

  expect(avatarBox.x).toBeLessThan(titleBox.x)
  expect(titleBox.height).toBeGreaterThan(30)
  const titleLineRects = await title.evaluate((element) => {
    const titleNode = Array.from(element.childNodes).find(
      node => node.nodeType === Node.TEXT_NODE && node.textContent?.trim(),
    )
    if (!titleNode) return []
    const range = document.createRange()
    range.selectNodeContents(titleNode)
    return Array.from(range.getClientRects()).map(rect => ({ x: rect.x, y: rect.y }))
  })
  expect(titleLineRects.length).toBeGreaterThanOrEqual(2)
  expect(titleLineRects[1].x).toBeLessThan(titleLineRects[0].x)
  expect(Math.abs(titleLineRects[1].x - titleBox.x)).toBeLessThan(2)
  expect(repliesBox.x).toBeGreaterThan(titleBox.x)
  expect(repliesBox.x + repliesBox.width).toBeLessThanOrEqual(rowBox.x + rowBox.width)
  expect(metadataBox.y).toBeGreaterThan(titleBox.y)
  expect(Math.abs(metadataBox.y - activityBox.y)).toBeLessThan(8)

  const compactRowBox = await compactRow.boundingBox()
  if (!compactRowBox) {
    throw new Error('The compact topic row did not produce a measurable layout box')
  }
  expect(compactRowBox.height).toBeLessThan(70)

  await testInfo.attach('mobile-topic-information-hierarchy', {
    body: await page.screenshot({ fullPage: false }),
    contentType: 'image/png',
  })

  await mobileAvatar.getByRole('link').click()
  await expect(page.getByLabel('Responsive Author')).toBeVisible()
  await page.keyboard.press('Escape')

  const deepTitle = rows.last().locator('[data-slot="topic-row-title"]')
  await deepTitle.scrollIntoViewIfNeeded()
  const previousScrollY = await page.evaluate(() => window.scrollY)
  expect(previousScrollY).toBeGreaterThan(0)

  await deepTitle.click()
  await page.waitForTimeout(100)
  expect(await page.evaluate(() => window.scrollY)).toBeGreaterThan(0)
  await expect(page.getByRole('heading', { level: 1, name: 'Topic detail' })).toBeVisible()
  expect(await page.evaluate(() => window.scrollY)).toBe(0)
  await expect(row).toBeHidden()
  await testInfo.attach('topic-detail-after-transition', {
    body: await page.screenshot({ fullPage: false }),
    contentType: 'image/png',
  })
  expect(errors).toEqual([])
})

test('keeps localized desktop topic headers on one line', async ({ page }) => {
  test.skip((page.viewportSize()?.width || 0) < 1024, 'Desktop table header only')
  await page.goto('/?responsive=topics&lang=ja')

  const activity = page.getByRole('columnheader', { name: 'アクティビティ' })
  await expect(activity).toBeVisible()
  const box = await activity.boundingBox()
  expect(box).not.toBeNull()
  expect(box!.height).toBeLessThanOrEqual(20)
  expect(box!.width).toBe(96)
  expect(await activity.evaluate(element => getComputedStyle(element).whiteSpace)).toBe('nowrap')

  const rowActivity = page.locator('[data-slot="topic-row"]').first().getByRole('cell').last()
  const rowBox = await rowActivity.boundingBox()
  expect(rowBox).not.toBeNull()
  expect(rowBox!.x).toBe(box!.x)
  expect(rowBox!.width).toBe(box!.width)
})

test('translates backend validation codes on the publish page', async ({ page }, testInfo) => {
  await page.goto('/publish?lang=en')
  await page.getByPlaceholder('Enter topic title').fill('x')
  await page.getByRole('button', { name: 'Coding' }).click()
  await page.getByRole('textbox', { name: 'Write and format the body directly' }).fill('Published body')
  await page.getByRole('button', { name: 'Publish topic' }).click()

  await expect(page.getByText('The title must be at least 3 characters.')).toBeVisible()
  await expect(page.getByText('topic.title.tooShort')).toHaveCount(0)
  await testInfo.attach('translated-publish-validation', {
    body: await page.screenshot({ fullPage: false }),
    contentType: 'image/png',
  })
})

test('serves workspace client source during development', async ({ page }) => {
  const clientRequests: string[] = []
  page.on('request', request => {
    if (request.url().includes('/packages/client/')) clientRequests.push(request.url())
  })
  await page.goto('/?lang=en')
  await expect(page.getByText('No topics yet')).toBeVisible()
  expect(clientRequests.some(url => url.includes('/packages/client/src/'))).toBe(true)
  expect(clientRequests.filter(url => url.includes('/packages/client/dist/'))).toEqual([])
})

test('uses matching hover menus for theme and language', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'chromium', 'Desktop hover interaction')
  await page.goto('/?lang=en')
  const theme = page.getByRole('button', { name: 'Choose theme' })
  await expect(theme.locator('svg')).toHaveAttribute('data-icon', 'system-theme')
  await theme.hover()
  const system = page.getByRole('menuitemradio', { name: 'System', exact: true })
  await expect(system).toBeVisible()
  await expect(system).toHaveAttribute('aria-checked', 'true')
  await system.hover()
  await expect(system).toBeVisible()
  await page.getByRole('menuitemradio', { name: 'Light', exact: true }).click()
  await expect(theme.locator('svg')).toHaveClass(/lucide-sun/)
  await theme.hover()
  await page.getByRole('menuitemradio', { name: 'Dark', exact: true }).click()
  await expect(theme.locator('svg')).toHaveClass(/lucide-moon/)
  await theme.hover()
  await page.getByRole('button', { name: 'Switch language' }).hover()
  await expect(system).not.toBeVisible()
  await expect(page.getByRole('menuitemradio', { name: 'English' })).toBeVisible()
  await theme.hover()
  await expect(page.getByRole('menuitemradio', { name: 'English' })).not.toBeVisible()
  await expect(system).toBeVisible()
  await page.getByText('No topics yet').hover()
  await expect(system).not.toBeVisible()
})

test('follows the system theme and remembers explicit preferences across reloads', async ({ page }) => {
  await page.emulateMedia({ colorScheme: 'dark' })
  await page.goto('/?lang=en')
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'gf-dark')
  await page.emulateMedia({ colorScheme: 'light' })
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'gf-light')
  await page.getByRole('button', { name: 'Choose theme' }).click()
  await page.getByRole('menuitemradio', { name: 'Dark', exact: true }).click()
  await page.reload()
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'gf-dark')
  await page.getByRole('button', { name: 'Choose theme' }).click()
  await page.getByRole('menuitemradio', { name: 'System', exact: true }).click()
  await page.reload()
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'gf-light')
  expect(await page.evaluate(() => localStorage.getItem('goose-site-theme'))).toBe('system')
  await page.emulateMedia({ colorScheme: 'dark' })
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'gf-dark')
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
  await expect(page.locator('meta[name="theme-color"]')).toHaveAttribute('content', '#fbfdff')
  expect(await page.evaluate(() =>
    getComputedStyle(document.documentElement)
      .getPropertyValue('--gf-color-base-100')
      .trim(),
  )).toBe('#fbfdff')

  const themeButton = page.getByRole('button', { name: 'Choose theme' })
  await themeButton.click()
  await page.getByRole('menuitemradio', { name: 'Dark', exact: true }).click()
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'gf-dark')
  await expect(page.getByRole('button', { name: 'Choose theme' })).toBeVisible()
  await expect(page.getByRole('menu')).toHaveCount(0)

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

test('prepares fresh admin lists before switching and restores their query on return', async ({ page, isMobile }, testInfo) => {
  const errors: string[] = []
  page.on('pageerror', error => errors.push(error.message))
  let requests = 0
  let failUsers = false
  let releaseUsers!: () => void
  const usersReady = new Promise<void>(resolve => { releaseUsers = resolve })
  await page.route('**/api/admin/user-list', async route => {
    requests++
    if (requests === 1) await usersReady
    if (failUsers) {
      await route.fulfill({ json: { code: 1, message: 'User list unavailable' } })
      return
    }
    const input = route.request().postDataJSON()
    await route.fulfill({ json: { code: 0, result: {
      list: Array.from({ length: 20 }, (_, index) => ({ userId: index + 7, username: index ? `user-${index}` : `alice-${requests}`, roleList: [], restrictionStatus: 'normal', createTime: '2026-10-01', lastActiveTime: '2026-10-11' })),
      total: 35, page: input.page, pageSize: input.pageSize,
    } } })
  })
  await page.route('**/api/admin/topics/list', route => route.fulfill({ json: { code: 0, result: { list: [], total: 0, page: 1, hasNext: false } } }))
  await page.route('**/api/admin/category-list', route => route.fulfill({ json: { code: 0, result: [] } }))
  async function selectPage(path: string) {
    if (isMobile) await page.getByRole('button', { name: 'Toggle Sidebar', exact: true }).click()
    await page.locator(`[data-slot="sidebar"] a[href="${path}"]`).click()
  }
  await page.goto('/admin?lang=en')
  await expect(page.getByRole('heading', { level: 2, name: 'Dashboard' })).toBeVisible()
  await selectPage('/admin/users')
  await expect(page.getByRole('progressbar')).toBeVisible()
  await expect(page.getByRole('heading', { level: 2, name: 'Dashboard' })).toBeVisible()
  expect(new URL(page.url()).pathname).toBe('/admin')
  await page.screenshot({ path: testInfo.outputPath('admin-navigation-pending.png') })
  releaseUsers()
  await expect(page.getByRole('heading', { level: 2, name: 'Users' })).toBeVisible()
  await expect(page.getByRole('progressbar')).toHaveCount(0)
  await expect(page.getByRole('link', { name: 'alice-1', exact: true })).toBeVisible()
  expect(requests).toBe(1)
  await page.getByPlaceholder('Search username…').fill('alice')
  await page.getByRole('button', { name: 'Search', exact: true }).click()
  await expect.poll(() => requests).toBe(2)
  await expect(page.getByRole('button', { name: 'Next', exact: true })).toBeEnabled()
  await page.getByRole('button', { name: 'Next', exact: true }).click()
  await expect.poll(() => new URL(page.url()).searchParams.get('page')).toBe('2')
  await expect.poll(() => requests).toBe(3)
  await page.evaluate(() => window.scrollTo(0, 300))
  const returnScroll = await page.evaluate(() => window.scrollY)
  await selectPage('/admin/posts')
  await expect.poll(() => new URL(page.url()).pathname).toBe('/admin/posts')
  failUsers = true
  await page.goBack()
  await expect(page.getByRole('alert')).toContainText('User list unavailable')
  await expect.poll(() => new URL(page.url()).pathname).toBe('/admin/posts')
  expect(requests).toBe(4)
  failUsers = false
  await page.goBack()
  await expect(page.getByRole('heading', { level: 2, name: 'Users' })).toHaveCount(1)
  await expect(page.getByPlaceholder('Search username…')).toHaveValue('alice')
  expect(new URL(page.url()).searchParams.get('page')).toBe('2')
  await expect(page.getByRole('link', { name: 'alice-5', exact: true })).toHaveCount(1)
  await expect.poll(() => page.evaluate(() => window.scrollY)).toBe(returnScroll)
  expect(requests).toBe(5)
  await page.screenshot({ path: testInfo.outputPath('admin-navigation-restored.png') })
  expect(errors).toEqual([])
})

test('loads the admin dashboard shell and deferred chart without accessibility regressions', async ({ page }, testInfo) => {
  const errors: string[] = []
  page.on('pageerror', error => errors.push(error.message))
  page.on('console', message => {
    if (message.type() === 'error') errors.push(message.text())
  })
  await page.route('https://api.github.com/repos/leancodebox/GooseForum/releases', route =>
    route.fulfill({ json: [] }),
  )

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
  const reply = boundary.getByRole('button', { name: 'Join the discussion' })
  const watch = boundary.getByRole('button', { name: 'Watch comments' })
  await expect(watch).toHaveAttribute('data-variant', 'outline')
  const [boundaryBox, replyBox, watchBox] = await Promise.all([
    boundary.boundingBox(),
    reply.boundingBox(),
    watch.boundingBox(),
  ])
  expect(boundaryBox).not.toBeNull()
  expect(replyBox).not.toBeNull()
  expect(watchBox).not.toBeNull()
  expect(replyBox!.x).toBeLessThan(watchBox!.x)
  const rightGap = boundaryBox!.x + boundaryBox!.width - watchBox!.x - watchBox!.width
  expect(rightGap).toBeGreaterThanOrEqual(23)

  await watch.click()
  const watched = boundary.getByRole('button', { name: 'Watching comments' })
  await expect(watched).toBeVisible()
  await expect(watched).toHaveAttribute('data-variant', 'default')
  await expect(page.getByRole('button', { name: 'Watching comments' })).toHaveCount(2)

  await reply.click()
  const toolbar = page.locator('[data-slot="markdown-composer-toolbar"]')
  await expect(toolbar.getByRole('button', { name: 'Post reply' })).toBeVisible()
  await expect(toolbar.getByRole('radio', { name: 'Markdown' })).toBeVisible()
  await expect(toolbar.getByRole('button', { name: 'Preview' })).toBeVisible()
  if (testInfo.project.name === 'mobile-chromium') {
    const editor = page.locator('.markdown-composer-content').filter({ visible: true }).first()
    expect(await editor.evaluate(element => parseFloat(getComputedStyle(element).fontSize))).toBeGreaterThanOrEqual(16)
  }
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

test('refreshes the active chat with the after cursor', async ({ page }) => {
  await page.clock.install()
  const requests: number[] = []
  page.on('request', request => {
    if (request.url().includes('/api/forum/chat/messages')) {
      const body = request.postDataJSON() as { afterId?: number }
      requests.push(body.afterId || 0)
    }
  })
  await page.goto('/messages?userId=9&lang=en')
  await expect(page.locator('[data-slot="message-scroller-item"]')).toHaveCount(4)
  await page.clock.fastForward(15_000)
  await expect.poll(() => requests).toEqual([0, 4])
  await expect(page.getByRole('region', { name: 'Private conversation' }).getByText('A new message arrived.')).toBeVisible()
  await page.clock.fastForward(15_000)
  await expect.poll(() => requests).toEqual([0, 4, 5])
  await expect(page.getByRole('region', { name: 'Private conversation' }).getByText('A new message arrived.')).toHaveCount(1)
})

test('edits, previews, and saves an announcement in Markdown', async ({ page }, testInfo) => {
  const errors: string[] = []
  page.on('pageerror', error => errors.push(error.message))
  page.on('console', message => {
    if (message.type() === 'error') errors.push(message.text())
  })

  await page.goto('/admin/settings/announcement?lang=en')
  await expect(page).toHaveTitle('Announcement - GooseForum')
  await page.getByRole('radio', { name: 'Markdown', exact: true }).click()
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
    settings: { enabled: true, content: '', items: [{ id: 'legacy', title: '', enabled: true, content: '## Updated announcement\n\n**Everything is ready.**' }] },
  })
  expect(errors).toEqual([])
  await testInfo.attach('verified-announcement-markdown-editor', {
    body: await page.screenshot({ fullPage: false }),
    contentType: 'image/png',
  })
})

test('keeps the avatar edit overlay below the worn badge', async ({ page }, testInfo) => {
  const errors: string[] = []
  page.on('pageerror', error => errors.push(error.message))
  page.on('console', message => {
    if (message.type() === 'error') errors.push(message.text())
  })
  const badge = {
    code: 'early-member', name: 'Early member', description: 'Early forum member',
    iconUrl: '/static/badges/early-member.svg', color: 'green', level: 'normal',
    grantedAt: '2026-01-01T00:00:00Z',
  }
  await page.route('**/__goose_page/settings**', route => route.fulfill({ json: {
    ...homePage,
    component: 'settings.index',
    props: {
      user: {
        id: 7, username: 'avatar-editor', nickname: 'Avatar Editor', email: 'avatar@example.test',
        locale: 'en', avatarUrl: '/static/pic/default-avatar.webp', profileCoverUrl: '',
        bio: '', signature: '', websiteName: '', website: '', prestige: 0,
        createdAt: '2026-01-01T00:00:00Z', externalInformation: {},
        wornBadgeCode: badge.code, badges: [badge], wearableBadges: [badge], wornBadge: badge,
      },
      stats: {
        topicCount: 0, replyCount: 0, followerCount: 0, followingCount: 0,
        likeReceivedCount: 0, likeGivenCount: 0, collectionCount: 0,
        createdAt: '2026-01-01T00:00:00Z',
      },
      tabs: [],
    },
    layout: {
      ...homePage.layout,
      viewer: { ...homePage.layout.viewer, id: 7, username: 'avatar-editor', isAuthenticated: true },
    },
    meta: { title: 'Avatar settings' },
    url: '/settings?lang=en',
  } }))
  await page.goto('/settings?lang=en')
  await expect(page).toHaveTitle(/Avatar settings/)
  const button = page.getByRole('button', { name: 'Change avatar', exact: true })
  await expect(button).toBeVisible()
  await expect(button.locator('[data-slot="avatar-image"]')).toBeVisible()
  const badgeImage = button.getByRole('img', { name: badge.name, exact: true })
  await expect(badgeImage).toBeVisible()
  await expect.poll(() => badgeImage.evaluate((image: HTMLImageElement) => image.naturalWidth)).toBeGreaterThan(0)
  await button.hover()
  const overlay = button.locator(':scope > span').last()
  const supportsHover = await page.evaluate(() => matchMedia('(hover: hover)').matches)
  await expect(overlay.locator('svg')).toHaveCSS('opacity', supportsHover ? '1' : '0')
  const order = await button.evaluate(element => {
    const overlay = element.lastElementChild as HTMLElement
    const image = element.querySelector('img[alt="Early member"]') as HTMLImageElement
    const badge = image.parentElement as HTMLElement
    const avatar = element.querySelector('[data-slot="avatar"]') as HTMLElement
    const rect = badge.getBoundingClientRect()
    const x = rect.left + rect.width * 0.35
    const y = rect.top + rect.height * 0.35
    const originals = [overlay.style.pointerEvents, badge.style.pointerEvents]
    // Include decorative layers in hit testing to inspect their paint order.
    overlay.style.pointerEvents = 'auto'
    badge.style.pointerEvents = 'auto'
    try {
      const layers = document.elementsFromPoint(x, y)
      return {
        badge: layers.indexOf(image), overlay: layers.indexOf(overlay), avatar: layers.indexOf(avatar),
      }
    } finally {
      overlay.style.pointerEvents = originals[0]
      badge.style.pointerEvents = originals[1]
    }
  })
  expect(order.badge).toBeGreaterThanOrEqual(0)
  expect(order.overlay).toBeGreaterThan(order.badge)
  expect(order.avatar).toBeGreaterThan(order.overlay)
  const fileChooser = page.waitForEvent('filechooser')
  await button.click()
  await fileChooser
  expect(errors).toEqual([])
  await testInfo.attach('avatar-overlay-below-badge', {
    body: await page.screenshot({
      path: testInfo.outputPath('avatar-overlay-below-badge.png'), fullPage: false,
    }), contentType: 'image/png',
  })
})

async function openThemeMenu(page: Page) {
  const trigger = page.getByRole('button', { name: 'Choose theme' })
  if (await page.evaluate(() => matchMedia('(hover: hover)').matches)) await trigger.hover()
  else await trigger.click()
}

test('theme reveal preserves mounted content and scroll without loading data again', async ({ page }) => {
  await page.route('**/__goose_page/**', route => route.fulfill({ json: {
    ...homePage,
    props: { ...homePage.props, announcement: {
      enabled: true, publishedAt: new Date().toISOString(),
      html: Array.from({ length: 24 }, (_, i) => `<p>Notice paragraph ${i}</p>`).join(''),
    } },
  } }))
  const requests: string[] = []
  page.on('request', request => {
    if (/\/(__goose_page|api)\//.test(request.url())) requests.push(request.url())
  })
  await page.goto('/?lang=en')
  const announcement = page.locator('.gf-prose-announcement')
  await expect(announcement).toBeVisible()
  const original = await announcement.elementHandle()
  await page.evaluate(() => window.scrollTo(0, 180))
  const scroll = await page.evaluate(() => window.scrollY)
  expect(scroll).toBeGreaterThan(0)
  const count = requests.length
  await openThemeMenu(page)
  await page.getByRole('menuitemradio', { name: 'Dark', exact: true }).click()
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'gf-dark')
  await expect(page.locator('html')).not.toHaveClass(/goose-theme-transition/)
  expect(await announcement.evaluate((node, old) => node === old, original)).toBe(true)
  expect(await page.evaluate(() => window.scrollY)).toBe(scroll)
  expect(requests).toHaveLength(count)
})

for (const fallback of ['reduced motion', 'unsupported API']) {
  test(`theme choice works with ${fallback}`, async ({ page }) => {
    if (fallback === 'reduced motion') await page.emulateMedia({ reducedMotion: 'reduce' })
    else await page.addInitScript(() => Object.defineProperty(document, 'startViewTransition', { value: undefined, configurable: true }))
    await page.goto('/?lang=en')
    await openThemeMenu(page)
    await page.getByRole('menuitemradio', { name: 'Dark', exact: true }).click()
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'gf-dark')
    await expect(page.locator('html')).not.toHaveClass(/goose-theme-transition/)
    await expect(page.getByText('No topics yet')).toBeVisible()
  })
}

test('announcement folding keeps read status independent and loads no data', async ({ page }) => {
  await page.route('**/__goose_page/**', route => route.fulfill({ json: {
    ...homePage,
    props: { ...homePage.props, announcement: {
      enabled: true, html: '<p>Foldable notice</p>', publishedAt: new Date().toISOString(),
    } },
  } }))
  const requests: string[] = []
  page.on('request', request => {
    if (/\/(__goose_page|api)\//.test(request.url())) requests.push(request.url())
  })
  await page.goto('/?lang=en')
  await expect(page.getByText('Foldable notice')).toBeVisible()
  const count = requests.length
  await page.getByRole('button', { name: 'Collapse announcement' }).click()
  await expect(page.getByText('Foldable notice')).not.toBeVisible()
  await expect(page.getByRole('button', { name: 'Mark announcement as read' })).toBeVisible()
  expect(await page.evaluate(() => localStorage.getItem('goose:announcement:last-read-published-at'))).toBeNull()
  await page.getByRole('button', { name: 'Expand announcement' }).click()
  await expect(page.getByText('Foldable notice')).toBeVisible()
  await page.getByRole('button', { name: 'Mark announcement as read' }).click()
  await expect(page.getByRole('button', { name: 'Mark announcement as read' })).toHaveCount(0)
  expect(requests).toHaveLength(count)
})

test('reply transitions preserve the minimized draft and existing fresh-reply behavior', async ({ page }) => {
  page.on('dialog', dialog => dialog.accept())
  await page.addInitScript(() => {
    const animate = Element.prototype.animate
    let paused = false
    Element.prototype.animate = function (...args) {
      const animation = animate.apply(this, args)
      if (!paused && this instanceof HTMLElement && this.classList.contains('goose-composer-surface') && this.dataset.state === 'closed') {
        paused = true
        animation.pause()
      }
      return animation
    }
  })
  await page.goto('/p/test/60?lang=en')
  const reply = page.locator('[data-slot="topic-reply-float-boundary"]').getByRole('button', { name: 'Join the discussion' })
  await reply.click()
  const panel = page.locator('.goose-composer-surface[data-state="open"] section')
  await panel.getByRole('radio', { name: 'Markdown', exact: true }).click()
  await panel.locator('textarea').fill('A draft that must survive transitions')
  await panel.locator('header').getByRole('button', { name: 'Join the discussion' }).click()
  const closed = page.locator('.goose-composer-surface[data-state="closed"]')
  // Exit content becomes inert immediately, even before the visual transition ends.
  await expect(closed).toHaveAttribute('inert', '')
  const bubble = page.locator('.goose-composer-surface[data-state="open"]')
  await bubble.getByRole('button', { name: 'Join the discussion', exact: true }).click()
  await panel.getByRole('radio', { name: 'Editor', exact: true }).click()
  await expect(panel.locator('[contenteditable="true"]')).toHaveText('A draft that must survive transitions')
  await page.evaluate(() => document.getAnimations().forEach(animation => animation.play()))
  await panel.locator('header').getByRole('button', { name: 'Close', exact: true }).click()
  await expect(page.locator('.goose-composer-surface')).toHaveCount(0)
  await reply.click()
  await expect(panel.locator('[contenteditable="true"]')).toHaveText('')
})

test('loads post edit history only after the history action is opened', async ({ page }, testInfo) => {
  let historyRequests = 0
  await page.route('**/api/forum/posts/revisions**', async route => {
    historyRequests++
    await route.fulfill({ json: {
      code: 0,
      result: {
        items: Array.from({ length: 10 }, (_, index) => ({ postId: 61, revisionNo: 10 - index, content: `Original body\n${'LongUnbrokenText'.repeat(30)}\n${'More content\n'.repeat(12)}`, createdAt: '2026-09-14T08:00:00Z' })),
        nextCursor: 0,
        hasMore: false,
      },
    } })
  })
  await page.goto('/p/test/60?lang=en')
  await expect(page.getByRole('heading', { name: 'Topic detail' })).toBeVisible()
  expect(historyRequests).toBe(0)

  await page.getByRole('button', { name: 'View edit history' }).click()
  await expect(page.getByRole('dialog', { name: 'Edit history' })).toContainText('Original body')
  expect(historyRequests).toBe(1)
  const dialog = page.getByRole('dialog', { name: 'Edit history' })
  const dimensions = await dialog.locator('.overflow-y-auto').evaluate(element => ({ width: element.clientWidth, scroll: element.scrollWidth }))
  expect(dimensions.scroll).toBeLessThanOrEqual(dimensions.width)
  await dialog.evaluate(element => Promise.all(element.getAnimations().map(animation => animation.finished)))
  const titleBefore = await dialog.getByRole('heading').boundingBox()
  await dialog.locator('.overflow-y-auto').evaluate(element => { element.scrollTop = element.scrollHeight })
  await expect(dialog.locator('article').last()).toBeInViewport()
  expect(Math.abs((await dialog.getByRole('heading').boundingBox())!.y - titleBefore!.y)).toBeLessThan(1)
  await expect(dialog.getByRole('button', { name: 'Close', exact: true })).toBeVisible()
  await dialog.locator('.overflow-y-auto').evaluate(element => { element.scrollTop = 0 })
  await page.screenshot({ path: testInfo.outputPath('edit-history.png') })
  await dialog.getByRole('button', { name: 'Close', exact: true }).click()
  await expect(dialog).toHaveCount(0)
})

test('settings underline follows variable-width tabs, keyboard selection and resize', async ({ page }) => {
  await page.route('**/__goose_page/settings**', route => route.fulfill({ json: {
    ...homePage,
    component: 'settings.index',
    props: {
      user: {
        id: 7, username: 'alice', nickname: 'Alice', email: 'alice@example.test',
        locale: 'en', avatarUrl: '', profileCoverUrl: '', bio: '', signature: '',
        websiteName: '', website: '', prestige: 0, createdAt: '2026-01-01T00:00:00Z',
        externalInformation: {}, wornBadgeCode: '', badges: [], wearableBadges: [],
      },
      stats: {
        topicCount: 0, replyCount: 0, followerCount: 0, followingCount: 0,
        likeReceivedCount: 0, likeGivenCount: 0, collectionCount: 0, createdAt: '2026-01-01T00:00:00Z',
      },
      privacy: { showTopics: true, showActivity: true, showFollowing: true },
      tabs: [{ key: 'profile', label: 'Profile' }, { key: 'privacy', label: 'Privacy' }],
    },
    layout: { ...homePage.layout, viewer: { ...homePage.layout.viewer, id: 7, username: 'alice', isAuthenticated: true } },
    url: '/settings?lang=en',
  } }))
  await page.goto('/settings?lang=en')
  const list = page.getByRole('tablist', { name: 'Settings sections' })
  const indicator = list.locator(':scope > span[aria-hidden="true"]')
  await expect(list).toHaveAttribute('data-sliding-indicator', 'true')
  const aligned = async () => list.evaluate(element => {
    const tab = element.querySelector('[data-state="active"]')!
    const line = element.querySelector(':scope > span')!
    const button = tab.getBoundingClientRect()
    const underline = line.getBoundingClientRect()
    return Math.abs(button.x - underline.x) < 1 && Math.abs(button.width - underline.width) < 1 && Math.abs(button.bottom - underline.bottom) < 1
  })
  await expect(indicator).toBeVisible()
  await expect.poll(aligned).toBe(true)
  await page.getByRole('tab', { name: 'Privacy', exact: true }).click()
  await expect(page.getByRole('checkbox', { name: 'Show my topics' })).toBeVisible()
  await expect.poll(aligned).toBe(true)
  await page.getByRole('tab', { name: 'Privacy', exact: true }).press('ArrowLeft')
  await expect(page.getByRole('tab', { name: 'Profile', exact: true })).toHaveAttribute('aria-selected', 'true')
  await expect.poll(aligned).toBe(true)
  await page.setViewportSize({ width: 600, height: 900 })
  await expect.poll(aligned).toBe(true)
})

test('keeps the active reply branch while its submission is pending', async ({ page }) => {
  let complete!: () => void
  const created = { ...topicPage.props.postStream.posts[0], id: 62, postNo: 2, content: 'Submitted reply', renderedContent: '<p>Submitted reply</p>', sourceVersion: 1 }
  await page.route('**/api/forum/posts/create', async route => {
    await new Promise<void>(resolve => { complete = resolve })
    await route.fulfill({ json: { code: 0, result: created } })
  })
  await page.route('**/api/forum/posts/window**', route => route.fulfill({ json: { code: 0, result: {
    ...topicPage.props.postStream, posts: [...topicPage.props.postStream.posts, created], afterPostNo: 2, maxPostNo: 2, total: 2,
  } } }))
  await page.goto('/p/test/60?lang=en')
  const reply = page.getByRole('button', { name: 'Reply', exact: true }).first()
  await reply.click()
  const panel = page.locator('.goose-composer-surface section')
  await panel.getByRole('radio', { name: 'Markdown', exact: true }).click()
  await panel.locator('textarea').fill('Submitted reply')
  await panel.getByRole('button', { name: 'Post reply', exact: true }).click()
  await expect(reply).toBeDisabled()
  await expect(page.getByRole('button', { name: 'Edit', exact: true }).first()).toBeDisabled()
  await panel.locator('textarea').fill('Unsubmitted next reply')
  complete()
  await expect(panel.getByRole('button', { name: 'Post reply', exact: true })).toBeEnabled()
  await expect(panel.locator('textarea')).toHaveValue('Unsubmitted next reply')
  await expect(panel).toBeVisible()
})

test('shows a flat suspension notice with the account security entry', async ({ page }, testInfo) => {
  await page.route('**/__goose_page/**', route => route.fulfill({ json: {
    ...homePage, layout: { ...homePage.layout, viewer: { ...publishPage.layout.viewer, restrictionStatus: 'suspended', restrictionUntil: '2099-01-01T00:00:00Z', restrictionReason: 'Community moderation decision' } },
  } }))
  await page.goto('/?lang=en')
  await expect(page.getByText('Your account is suspended. You can read and manage account security.')).toBeVisible()
  await expect(page.getByRole('link', { name: 'Account security', exact: true })).toHaveAttribute('href', '/settings?tab=account')
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  await page.screenshot({ path: testInfo.outputPath('account-suspended.png'), fullPage: true })
})

test('filters personal login history with controls that fit the settings panel', async ({ page }, testInfo) => {
  await page.clock.setFixedTime(new Date('2026-10-04T12:00:00Z'))
  await page.route('**/__goose_page/settings**', route => route.fulfill({ json: {
    ...homePage, component: 'settings.index', url: '/settings?tab=sessions&lang=en',
    props: {
      user: {
        id: 7, username: 'alice', nickname: 'Alice', email: 'alice@example.test',
        locale: 'en', avatarUrl: '', profileCoverUrl: '', bio: '', signature: '',
        websiteName: '', website: '', prestige: 0, createdAt: '2026-01-01T00:00:00Z',
        externalInformation: {}, wornBadgeCode: '', badges: [], wearableBadges: [],
      },
      stats: { topicCount: 0, replyCount: 0, followerCount: 0, followingCount: 0, likeReceivedCount: 0, likeGivenCount: 0, collectionCount: 0, createdAt: '2026-01-01T00:00:00Z' },
      privacy: { showTopics: true, showActivity: true, showFollowing: true },
      tabs: [{ key: 'profile', label: 'Profile' }, { key: 'sessions', label: 'Login devices' }],
    },
    layout: { ...homePage.layout, viewer: { ...homePage.layout.viewer, id: 7, username: 'alice', isAuthenticated: true } },
  } }))
  await page.route('**/api/auth-sessions', route => route.fulfill({ json: { code: 0, result: [] } }))
  let method = ''
  let cursor = 0
  let since = ''
  let until = ''
  await page.route('**/api/auth-logs**', route => {
    const query = new URL(route.request().url()).searchParams
    method = query.get('method') || ''
    cursor = Number(query.get('cursor') || 0)
    since = query.get('since') || ''
    until = query.get('until') || ''
    expect(query.has('page')).toBe(false)
    return route.fulfill({ json: { code: 0, result: {
      list: [{id:cursor ? 5 : 45,userId:7,action:'login_success',authMethod:method || 'password',result:'success',oauthProvider:'',clientIp:'127.0.0.1',userAgent:cursor ? 'Older security browser' : 'Recent security browser',createdAt:'2026-10-03T09:00:00Z'}],
      nextCursor:cursor ? 0 : 45,hasMore:cursor === 0,pageSize:20,
    } } })
  })
  await page.goto('/settings?tab=sessions&lang=en')
  await expect(page.getByRole('heading', { name: 'Login and security history' })).toBeVisible()
  const select = page.getByLabel('Method', { exact: true })
  await expect(select).toHaveAttribute('data-slot', 'select-trigger')
  await page.getByRole('button', {name:'Next', exact:true}).click()
  await expect.poll(() => cursor).toBe(45)
  await expect(page.getByText('Older security browser', {exact:true})).toBeVisible()
  await expect(page.getByRole('button', {name:'Next', exact:true})).toBeDisabled()
  await page.getByRole('button', {name:'Previous', exact:true}).click()
  await expect.poll(() => cursor).toBe(0)
  await page.getByRole('button', {name:'Next', exact:true}).click()
  await expect.poll(() => cursor).toBe(45)
  await select.click()
  const menu = page.getByRole('listbox')
  const triggerBounds = await select.boundingBox()
  await expect.poll(async () => Math.abs((await menu.boundingBox())!.width - triggerBounds!.width)).toBeLessThanOrEqual(1)
  await expect.poll(async () => Math.abs((await menu.boundingBox())!.x - triggerBounds!.x)).toBeLessThanOrEqual(1)
  await page.screenshot({path: testInfo.outputPath('personal-log-select.png')})
  await page.getByRole('option', { name: 'Two-factor authentication', exact: true }).click()
  await expect.poll(() => method).toBe('mfa')
  await expect.poll(() => cursor).toBe(0)
  const range = page.getByLabel('Date range', {exact:true})
  await range.click()
  const calendar = page.locator('[data-slot="date-range-picker-content"][data-state="open"]')
  await expect(calendar).toHaveCSS('opacity', '1')
  expect(await calendar.evaluate(element => element.getBoundingClientRect().right <= innerWidth + 1 && element.getBoundingClientRect().left >= 0)).toBe(true)
  await page.screenshot({path:testInfo.outputPath('personal-log-date-range.png')})
  await calendar.locator('[data-day="10/1/2026"]').first().click()
  expect(since).toBe('')
  await calendar.locator('[data-day="10/4/2026"]').first().click()
  await expect.poll(() => since).not.toBe('')
  expect(new Date(since).getDate()).toBe(1)
  expect(new Date(until).getDate()).toBe(4)
  await range.click()
  await page.getByRole('button', {name:'Clear date range',exact:true}).click()
  await expect.poll(() => since).toBe('')
  expect(until).toBe('')
  await expect(select.locator('[data-slot="select-value"]')).toHaveText('Two-factor authentication')
  expect(await select.locator('[data-slot="select-value"]').evaluate(element => element.scrollWidth <= element.clientWidth + 1)).toBe(true)
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true)
  await page.getByRole('heading', { name: 'Login and security history' }).scrollIntoViewIfNeeded()
  await page.screenshot({ path: testInfo.outputPath('personal-login-logs.png'), fullPage: true })
  await select.click()
  await page.getByRole('option', { name: 'All', exact: true }).click()
  await expect.poll(() => method).toBe('')
})

test('filters the new administrator login history in the existing flat layout', async ({ page }, testInfo) => {
  await page.clock.setFixedTime(new Date('2026-10-04T12:00:00Z'))
  await page.route('**/__goose_page/admin/auth-logs**', route => route.fulfill({ json: {
    ...adminPage, url: '/admin/auth-logs?lang=en',
  } }))
  let filter = ''
  let cursor = 0
  await page.route('**/api/admin/auth-logs**', route => {
    const query = new URL(route.request().url()).searchParams
    filter = query.get('result') || ''
    cursor = Number(query.get('cursor') || 0)
    expect(query.has('page')).toBe(false)
    return route.fulfill({ json: { code: 0, result: {
      list: [{ id: cursor ? 12 : 32, userId: 7, action: 'login_success', authMethod: 'password', result: 'success', oauthProvider: '', clientIp: '127.0.0.1', userAgent: cursor ? 'Older browser' : 'Quality browser', createdAt: '2026-10-03T09:00:00Z' }], nextCursor:cursor ? 0 : 32,hasMore:cursor === 0,pageSize:20,
    } } })
  })
  await page.goto('/admin/auth-logs?lang=en')
  await expect(page.getByRole('heading', { name: 'Login and security history' })).toBeVisible()
  await expect(page.getByText('Quality browser')).toBeVisible()
  await page.getByRole('button', {name:'Next', exact:true}).click()
  await expect.poll(() => cursor).toBe(32)
  await expect(page.getByText('Older browser', {exact:true})).toBeVisible()
  await expect(page.getByRole('button', {name:'Next', exact:true})).toBeDisabled()
  await page.getByRole('button', {name:'Previous', exact:true}).click()
  await expect.poll(() => cursor).toBe(0)
  await page.getByRole('button', {name:'Next', exact:true}).click()
  await expect.poll(() => cursor).toBe(32)
  await page.getByLabel('Result', { exact: true }).click()
  const triggerBounds = await page.getByLabel('Result', {exact:true}).boundingBox()
  const menu = page.getByRole('listbox')
  await expect.poll(async () => Math.abs((await menu.boundingBox())!.width - triggerBounds!.width)).toBeLessThanOrEqual(1)
  await expect.poll(async () => Math.abs((await menu.boundingBox())!.x - triggerBounds!.x)).toBeLessThanOrEqual(1)
  await page.screenshot({path: testInfo.outputPath('admin-log-select.png')})
  await page.getByRole('option', { name: 'Failure', exact: true }).click()
  await expect.poll(() => filter).toBe('failure')
  await expect.poll(() => cursor).toBe(0)
  await page.getByLabel('Date range', {exact:true}).click()
  const calendar = page.locator('[data-slot="date-range-picker-content"][data-state="open"]')
  await expect(calendar).toHaveCSS('opacity', '1')
  expect(await calendar.evaluate(element => element.getBoundingClientRect().right <= innerWidth + 1 && element.getBoundingClientRect().left >= 0)).toBe(true)
  await page.screenshot({path:testInfo.outputPath('admin-log-date-range.png')})
  await calendar.locator('[data-day="10/1/2026"]').first().click()
  await calendar.locator('[data-day="10/4/2026"]').first().click()
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  await page.screenshot({ path: testInfo.outputPath('admin-login-logs.png'), fullPage: true })
  expect((await new AxeBuilder({ page }).withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa']).analyze()).violations).toEqual([])
})

test('restores the browser local draft only after an explicit choice', async ({ page }, testInfo) => {
  page.on('dialog', dialog => dialog.accept())
  await page.goto('/publish?lang=en')
  await page.getByPlaceholder('Enter topic title').fill('Browser draft survives refresh')
  await page.getByRole('button', { name: 'Coding', exact: true }).click()
  await page.getByRole('textbox', { name: 'Write and format the body directly' }).fill('The complete unfinished draft')
  await expect(page.getByText('Local copy saved', { exact: true })).toBeVisible()
  await page.reload()
  await expect(page.getByRole('dialog', { name: 'Local draft found' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Restore copy', exact: true })).toBeVisible()
  await expect(page.getByPlaceholder('Enter topic title')).toHaveValue('')
  await page.getByRole('button', { name: 'Restore copy', exact: true }).click()
  await expect(page.getByPlaceholder('Enter topic title')).toHaveValue('Browser draft survives refresh')
  await expect(page.getByRole('textbox', { name: 'Write and format the body directly' })).toHaveText('The complete unfinished draft')
  const header = page.getByRole('heading', { name: 'Publish topic', exact: true }).locator('xpath=ancestor::header')
  expect(await header.evaluate(element => getComputedStyle(element).borderBottomWidth)).toBe('0px')
  await expect(page.getByText('Local copy saved', { exact: true })).toBeVisible()
  const checklist = page.getByRole('heading', { name: 'Publish checklist', exact: true })
  expect(await checklist.evaluate(element => Boolean(element.closest('aside')?.querySelector('[data-slot="local-draft-status"]')))).toBe(true)
  const checklistBox = await checklist.boundingBox()
  const statusBox = await page.locator('[data-slot="local-draft-status"]').boundingBox()
  expect(statusBox!.y).toBeGreaterThan(checklistBox!.y + checklistBox!.height)
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  await page.screenshot({ path: testInfo.outputPath('local-draft-restored.png'), fullPage: true })
})

test('does not save empty metadata and cleans cleared or discarded local drafts', async ({ page }, testInfo) => {
  const storedDrafts = () => page.evaluate(async () => {
    const request = indexedDB.open('gooseforum-local-drafts', 1)
    const database = await new Promise<IDBDatabase>((resolve, reject) => {
      request.onsuccess = () => resolve(request.result)
      request.onerror = () => reject(request.error)
    })
    const all = database.transaction('drafts', 'readonly').objectStore('drafts').getAll()
    const count = await new Promise<number>((resolve, reject) => {
      all.onsuccess = () => resolve(all.result.length)
      all.onerror = () => reject(all.error)
    })
    database.close()
    return count
  })
  await page.goto('/publish?lang=en')
  page.on('dialog', dialog => dialog.accept())
  await page.getByRole('button', { name: 'Coding', exact: true }).click()
  await page.waitForTimeout(1200)
  expect(await storedDrafts()).toBe(0)
  const title = page.getByPlaceholder('Enter topic title')
  const editor = page.getByRole('textbox', { name: 'Write and format the body directly' })
  await title.fill('Unfinished topic')
  await editor.fill('A single local draft')
  await expect(page.getByText('Local copy saved', { exact: true })).toBeVisible()
  expect(await storedDrafts()).toBe(1)
  await title.clear()
  await editor.press('ControlOrMeta+A')
  await editor.press('Backspace')
  await expect(editor).toHaveText('')
  await expect.poll(storedDrafts).toBe(0)
  await editor.fill('Keep this until choosing discard')
  await expect(page.getByText('Local copy saved', { exact: true })).toBeVisible()
  await page.reload()
  const prompt = page.getByRole('dialog', { name: 'Local draft found' })
  await expect(prompt).toBeVisible()
  await page.screenshot({ path: testInfo.outputPath('single-draft-recovery.png'), fullPage: true })
  await prompt.getByRole('button', { name: 'Discard local draft', exact: true }).click()
  await expect(prompt).not.toBeVisible()
  await expect.poll(storedDrafts).toBe(0)
  await page.reload()
  await expect(page.getByPlaceholder('Enter topic title')).toBeVisible()
  await expect(prompt).not.toBeVisible()
  expect(await storedDrafts()).toBe(0)
})

test('inserts a stable mention and preserves it across editor modes', async ({ page }, testInfo) => {
  await page.route('**/api/mention-users**', route => route.fulfill({ json: { code: 0, result: [{ id: '12', username: 'alice' }] } }))
  await page.goto('/publish?lang=en')
  await page.getByRole('radio', { name: 'Markdown', exact: true }).click()
  const body = page.getByPlaceholder('Enter body text, Markdown supported; paste or drag images here')
  await body.fill('Hello @ali')
  await body.press('End')
  await expect(page.getByRole('option', { name: '@alice', exact: true })).toBeVisible()
  await body.press('Enter')
  await expect(body).toHaveValue('Hello [mention user="12"]@alice[/mention]')
  await page.getByRole('radio', { name: 'Editor', exact: true }).click()
  await expect(page.locator('[contenteditable="true"] [data-mention-user="12"]')).toHaveText('@alice')
  await page.getByRole('radio', { name: 'Markdown', exact: true }).click()
  await expect(body).toHaveValue('Hello [mention user="12"]@alice[/mention]')
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  await page.screenshot({ path: testInfo.outputPath('stable-mention-editor.png'), fullPage: true })
})

for (const selection of ['click', 'keyboard'] as const) {
  test(`keeps the caret on the same line after ${selection} mention selection`, async ({ page }, testInfo) => {
    await page.route('**/api/mention-users**', route => route.fulfill({ json: { code: 0, result: [{ id: '12', username: 'alice' }] } }))
    await page.goto('/publish?lang=en')
    const editor = page.getByRole('textbox', { name: 'Write and format the body directly' })
    await editor.fill('Hello @ali')
    await editor.press('End')
    const option = page.getByRole('option', { name: '@alice', exact: true })
    await expect(option).toBeVisible()
    if (selection === 'click') await option.click()
    else await editor.press('Enter')
    const mention = editor.locator('[data-mention-user="12"]')
    await expect(mention).toHaveText('@alice')
    await expect(editor).toBeFocused()
    const separator = editor.locator('img.ProseMirror-separator')
    await expect(separator).toHaveCSS('display', 'inline')
    await expect(separator).toHaveCSS('margin-top', '0px')
    const mentionBox = await mention.boundingBox()
    const caretLine = await editor.evaluate(element => {
      const range = window.getSelection()?.getRangeAt(0)
      if (!range || !range.collapsed || !element.contains(range.startContainer)) return null
      const paragraph = element.querySelector('p')!
      const rect = paragraph.getBoundingClientRect()
      return { y: rect.y, height: rect.height, lineHeight: Number.parseFloat(getComputedStyle(paragraph).lineHeight) }
    })
    expect(caretLine).not.toBeNull()
    expect(mentionBox).not.toBeNull()
    expect(caretLine!.height).toBeLessThanOrEqual(caretLine!.lineHeight + 1)
    expect(mentionBox!.y).toBeLessThan(caretLine!.y + caretLine!.lineHeight)
    await page.screenshot({ path: testInfo.outputPath('mention-caret-before-typing.png'), fullPage: true })
    await editor.pressSequentially(' continues')
    await expect(editor).toHaveText('Hello @alice continues')
    await editor.press('Enter')
    await editor.pressSequentially('Next line')
    await expect(editor.locator('p')).toHaveCount(2)
    await expect(editor.locator('p').last()).toHaveText('Next line')
    await page.getByRole('radio', { name: 'Markdown', exact: true }).click()
    await expect(page.getByPlaceholder('Enter body text, Markdown supported; paste or drag images here'))
      .toHaveValue('Hello [mention user="12"]@alice[/mention] continues\n\nNext line')
  })
}

test('completes the second-factor challenge with recoverable feedback', async ({ page }, testInfo) => {
  await page.route('**/__goose_page/login**', route => route.fulfill({ json: {
    ...homePage,
    component: 'auth.login',
    props: { initialMode: 'login', redirectUrl: '/', oauthProviders: [] },
    url: '/login?mfa=1&lang=en',
  } }))
  let attempts = 0
  await page.route('**/api/mfa/login', async route => {
    attempts++
    const input = route.request().postDataJSON() as { code: string }
    expect(input.code).toBe(attempts === 1 ? '000000' : '123456')
    await route.fulfill({ json: attempts === 1
      ? { code: 1, messageCode: 'mfa.invalid' }
      : { code: 0, result: { redirect: '/' } },
    })
  })
  await page.goto('/login?mfa=1&lang=en')
  await expect(page.getByRole('heading', { name: 'Two-factor authentication' })).toBeVisible()
  const code = page.getByLabel('Authenticator or recovery code')
  await code.fill('000000')
  await page.getByRole('button', { name: 'Verify', exact: true }).click()
  await expect(page.getByRole('alert')).toBeVisible()
  await expect(code).toBeEnabled()
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  await testInfo.attach('second-factor-login', { body: await page.screenshot({ path: testInfo.outputPath('mfa-login.png') }), contentType: 'image/png' })
  expect((await new AxeBuilder({ page }).withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa']).analyze()).violations).toEqual([])
  await code.fill('123456')
  await page.getByRole('button', { name: 'Verify', exact: true }).click()
  await expect(page).toHaveURL(/\/$/)
  expect(attempts).toBe(2)
})

test('shows MFA availability in its dedicated settings tab', async ({ page }, testInfo) => {
  await page.route('**/__goose_page/settings**', route => route.fulfill({ json: {
    ...homePage,
    component: 'settings.index',
    props: {
      user: {
        id: 7, username: 'alice', nickname: 'Alice', email: 'alice@example.test',
        locale: 'en', avatarUrl: '', profileCoverUrl: '', bio: '', signature: '',
        websiteName: '', website: '', prestige: 0, createdAt: '2026-01-01T00:00:00Z',
        externalInformation: {}, wornBadgeCode: '', badges: [], wearableBadges: [],
      },
      stats: {
        topicCount: 0, replyCount: 0, followerCount: 0, followingCount: 0,
        likeReceivedCount: 0, likeGivenCount: 0, collectionCount: 0, createdAt: '2026-01-01T00:00:00Z',
      },
      privacy: { showTopics: true, showActivity: true, showFollowing: true },
      tabs: [{ key: 'account', label: 'Account', url: '/settings?tab=account' }, { key: 'mfa', label: 'Two-factor authentication', url: '/settings?tab=mfa' }],
    },
    layout: { ...homePage.layout, viewer: { ...homePage.layout.viewer, id: 7, username: 'alice', isAuthenticated: true } },
    url: '/settings?tab=mfa&lang=en',
  } }))
  await page.route('**/api/mfa', route => route.fulfill({ json: {
    code: 0, result: { enabled: false, available: false, remainingCodes: 0 },
  } }))
  await page.goto('/settings?tab=mfa&lang=en')
  await expect(page.getByRole('heading', { name: 'Two-factor authentication' })).toBeVisible()
  await expect(page.getByText('Two-factor authentication is unavailable. Contact the administrator.')).toBeVisible()
  await expect(page.getByRole('button', { name: 'Set up authenticator' })).toHaveCount(0)
  await expect(page.getByRole('tab', { name: 'Two-factor authentication' })).toHaveAttribute('aria-selected', 'true')
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  await testInfo.attach('flat-mfa-settings', { body: await page.screenshot({ path: testInfo.outputPath('mfa-settings.png'), fullPage: true }), contentType: 'image/png' })
  expect((await new AxeBuilder({ page }).withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa']).analyze()).violations).toEqual([])
  await page.route('**/api/mfa', route => route.fulfill({ json: {
    code: 0, result: { enabled: false, available: true, remainingCodes: 0 },
  } }))
  await page.route('**/api/mfa/begin', route => route.fulfill({ json: {
    code: 0, result: { secret: 'JBSWY3DPEHPK3PXP', uri: 'otpauth://totp/GooseForum:alice?secret=JBSWY3DPEHPK3PXP&issuer=GooseForum' },
  } }))
  await page.reload()
  await page.getByLabel('Account password', { exact: true }).fill('password123')
  await page.getByRole('button', { name: 'Set up authenticator' }).click()
  const qr = page.getByRole('img', { name: 'Authenticator setup QR code' })
  await expect(qr).toBeVisible()
  await expect(page.getByLabel('Account password', { exact: true })).toHaveCount(0)
  expect(await qr.locator('path').count()).toBeGreaterThan(0)
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  await testInfo.attach('mfa-qr-setup', { body: await page.screenshot({ path: testInfo.outputPath('mfa-qr-setup.png'), fullPage: true }), contentType: 'image/png' })
  await page.evaluate(() => document.documentElement.setAttribute('data-theme', 'gf-dark'))
  await expect(qr).toHaveCSS('background-color', 'rgb(255, 255, 255)')
  await testInfo.attach('mfa-qr-dark', { body: await page.screenshot({ path: testInfo.outputPath('mfa-qr-dark.png'), fullPage: true }), contentType: 'image/png' })
})

test('announcement restores browser preferences without playing an entry animation', async ({ page }) => {
  let publishedAt = '2026-10-03T09:00:00Z'
  await page.route('**/__goose_page/**', route => route.fulfill({ json: {
    ...homePage,
    props: { ...homePage.props, announcement: {
      enabled: true, html: '<p>Persistent notice</p>', publishedAt,
    } },
  } }))
  await page.goto('/?lang=en')
  const content = page.locator('.goose-announcement-content')
  await expect(content).toHaveCSS('animation-name', 'none')
  await page.getByRole('button', { name: 'Collapse announcement' }).click()
  await expect(page.getByText('Persistent notice')).not.toBeVisible()
  await page.reload()
  await expect(page.getByRole('button', { name: 'Expand announcement' })).toBeVisible()
  await expect(page.getByText('Persistent notice')).not.toBeVisible()
  await page.getByRole('button', { name: 'Expand announcement' }).click()
  await expect(page.getByText('Persistent notice')).toBeVisible()
  await expect(content).toHaveCSS('animation-name', 'none')
  await page.reload()
  await expect(page.getByText('Persistent notice')).toBeVisible()
  await expect(content).toHaveCSS('animation-name', 'none')
  await page.getByRole('button', { name: 'Collapse announcement' }).click()
  await expect(page.getByText('Persistent notice')).not.toBeVisible()
  publishedAt = '2026-10-03T10:00:00Z'
  await page.reload()
  await expect(page.getByText('Persistent notice')).toBeVisible()
  await expect(content).toHaveCSS('animation-name', 'none')
})

test('restoring the reply window never reveals an exiting bubble when its animation is released', async ({ page }) => {
  await page.addInitScript(() => {
    const testWindow = window as Window & { exitOpacities?: number[] }
    testWindow.exitOpacities = []
    const animate = Element.prototype.animate
    Element.prototype.animate = function (...args) {
      const animation = animate.apply(this, args)
      if (this instanceof HTMLElement && this.classList.contains('goose-composer-surface')) {
        const node = this
        const cancel = animation.cancel.bind(animation)
        animation.cancel = () => {
          const finishedExit = node.isConnected && node.dataset.state === 'closed' && animation.playState === 'finished'
          cancel()
          if (finishedExit) testWindow.exitOpacities!.push(Number(getComputedStyle(node).opacity))
        }
      }
      return animation
    }
  })
  await page.goto('/p/test/60?lang=en')
  await page.locator('[data-slot="topic-reply-float-boundary"]').getByRole('button', { name: 'Join the discussion' }).click()
  const panel = page.locator('.goose-composer-surface[data-state="open"] section')
  for (let i = 0; i < 3; i++) {
    await panel.locator('header').getByRole('button', { name: 'Join the discussion' }).click()
    const bubble = page.locator('.goose-composer-surface[data-state="open"]')
    await bubble.getByRole('button', { name: 'Join the discussion', exact: true }).click()
    await expect(panel).toBeVisible()
    await expect(page.locator('.goose-composer-surface[data-state="closed"]')).toHaveCount(0)
  }
  const samples = await page.evaluate(() => (window as Window & { exitOpacities: number[] }).exitOpacities)
  expect(samples.length).toBeGreaterThanOrEqual(3)
  expect(samples.every(opacity => opacity === 0)).toBe(true)
})

test('manages ordered announcements with visual editing and preserves drafts across modes', async ({ page }, testInfo) => {
  await page.route('**/api/admin/announcement', route => route.fulfill({ json: { code: 0, result: {
    enabled: true, content: '', items: [
      { id: 'a', title: 'First notice', content: 'First body', enabled: true },
      { id: 'b', title: 'Second notice', content: 'Second body', enabled: true },
    ],
  } } }))
  await page.goto('/admin/settings/announcement?lang=en')
  const visual = page.getByRole('textbox', { name: 'Edit the announcement directly' })
  await expect(visual).toHaveText('First body')
  await visual.fill('Updated first body')
  await page.getByRole('radio', { name: 'Markdown', exact: true }).click()
  await expect(page.getByRole('textbox', { name: 'Announcement content' })).toHaveValue('Updated first body')
  await page.getByRole('radio', { name: 'Visual', exact: true }).click()
  await expect(visual).toHaveText('Updated first body')
  const list = page.getByRole('region', { name: 'Announcements', exact: true })
  await list.getByRole('button', { name: 'Move down' }).first().click()
  await expect(list.getByRole('button', { name: /1\. Second notice/ })).toBeVisible()
  await expect(visual).toHaveText('Updated first body')
  await list.getByRole('button', { name: 'Add', exact: true }).click()
  await page.getByRole('textbox', { name: 'Announcement title' }).fill('Third notice')
  await visual.fill('Third body')
  await page.getByRole('radio', { name: 'Preview', exact: true }).click()
  await expect(page.locator('[data-slot="announcement-markdown-preview"]')).toHaveText('Third body')
  const save = page.waitForRequest(request => request.url().includes('/api/admin/save-announcement'))
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  const items = (await save).postDataJSON().settings.items
  expect(items.map((item: { title: string }) => item.title)).toEqual(['Second notice', 'First notice', 'Third notice'])
  expect(items[1].content).toBe('Updated first body')
  expect(items[2].content).toBe('Third body')
  expect(new Set(items.map((item: { id: string }) => item.id)).size).toBe(3)
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  await testInfo.attach('admin-multiple-announcements', {
    body: await page.screenshot({ path: testInfo.outputPath('admin-multiple-announcements.png'), fullPage: true }), contentType: 'image/png',
  })
})

test('shows one announcement at a time in order and reopens the panel after a total update', async ({ page }, testInfo) => {
  let publishedAt = Date.parse('2026-10-03T10:00:00Z')
  const requests: string[] = []
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await page.route('**/__goose_page/**', route => route.fulfill({ json: {
    ...homePage, props: { ...homePage.props, announcement: {
      enabled: true, html: '', publishedAt, items: [
        { id: 'b', title: 'Second notice', html: '<p>Second body</p>' },
        { id: 'a', title: 'First notice', html: '<p>First body</p>' + '<p>Long announcement paragraph.</p>'.repeat(8) },
      ],
    } },
  } }))
  page.on('request', request => { if (/\/(__goose_page|api)\//.test(request.url())) requests.push(request.url()) })
  await page.goto('/?lang=en')
  const announcement = page.getByRole('complementary', { name: 'Announcement', exact: true })
  await expect(announcement.getByText('Second body')).toBeVisible()
  await expect(announcement.getByText('First body')).toHaveCount(0)
  const count = requests.length
  const before = await announcement.boundingBox()
  const arrow = announcement.getByRole('button', { name: /^(Collapse|Expand) announcement$/ }).locator('svg')
  const arrowBefore = await arrow.boundingBox()
  const selectors = await announcement.getByLabel('Choose announcement').boundingBox()
  expect(selectors!.x + selectors!.width).toBeLessThanOrEqual(arrowBefore!.x)
  await announcement.getByRole('button', { name: 'Announcement 2: First notice' }).click()
  await expect(announcement.getByText('First body')).toBeVisible()
  await expect(announcement.getByText('Second body')).toHaveCount(0)
  expect((await announcement.boundingBox())!.height).toBeGreaterThan(before!.height)
  const body = announcement.locator('[data-content-variant="announcement"]')
  expect(await body.evaluate(element => {
    const container = element.parentElement!
    return container.scrollHeight <= container.clientHeight
  })).toBe(true)
  expect(requests).toHaveLength(count)
  await arrow.click()
  expect((await arrow.boundingBox())!.x).toBe(arrowBefore!.x)
  await page.reload()
  await expect(announcement.getByRole('button', { name: 'Expand announcement' })).toBeVisible()
  publishedAt += 1
  await page.reload()
  await expect(announcement.getByText('Second body')).toBeVisible()
  await expect(page.locator('.goose-announcement-content')).toHaveCSS('animation-name', 'none')
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  await testInfo.attach('home-multiple-announcements', {
    body: await page.screenshot({ path: testInfo.outputPath('home-multiple-announcements.png'), fullPage: false }), contentType: 'image/png',
  })
})

test('desktop sidebar remembers its width preference and leaves mobile navigation available', async ({ page }) => {
  const requests: string[] = []
  await page.route('**/__goose_page/**', route => route.fulfill({ json: homePage }))
  page.on('request', request => { if (/\/(__goose_page|api)\//.test(request.url())) requests.push(request.url()) })
  await page.goto('/?lang=en')
  const content = page.locator('[data-slot="goose-page-content"]')
  const sidebar = page.locator('#goose-desktop-sidebar')
  if ((page.viewportSize()?.width || 0) < 1024) {
    await page.evaluate(() => localStorage.setItem('goose:shell-sidebar-collapsed', 'true'))
    await page.reload()
    await expect(page.getByRole('button', { name: 'Expand sidebar' })).toBeHidden()
    await page.getByRole('button', { name: 'Open menu', exact: true }).click()
    await expect(page.getByRole('dialog').getByRole('link', { name: 'Topics', exact: true })).toBeVisible()
    await page.getByRole('button', { name: 'Close menu', exact: true }).click()
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
    return
  }
  const width = (await content.boundingBox())!.width
  const count = requests.length
  const more = sidebar.locator('.site-more-trigger')
  const transition = await more.evaluate(element => getComputedStyle(element).transitionProperty)
  expect(transition).toContain('background-color')
  expect(transition).not.toMatch(/\b(all|visibility)\b/)
  const background = await more.evaluate(element => getComputedStyle(element).backgroundColor)
  await more.click()
  await expect(more).toHaveAttribute('data-state', 'open')
  await expect.poll(() => more.evaluate(element => getComputedStyle(element).backgroundColor)).not.toBe(background)
  await page.keyboard.press('Escape')
  await page.getByRole('button', { name: 'Collapse sidebar' }).click()
  await expect(sidebar.locator('.site-more-trigger')).toHaveCSS('visibility', 'hidden')
  await expect(sidebar).toHaveAttribute('inert', '')
  await expect(sidebar).toHaveAttribute('aria-hidden', 'true')
  await expect.poll(async () => (await content.boundingBox())!.width).toBeGreaterThan(width + 200)
  expect(requests).toHaveLength(count)
  await page.reload()
  await expect(page.getByRole('button', { name: 'Expand sidebar' })).toHaveAttribute('aria-expanded', 'false')
  await expect(page.locator('.goose-shell-grid')).toHaveCSS('transition-duration', '0s')
  await page.getByRole('button', { name: 'Expand sidebar' }).click()
  await expect(sidebar).not.toHaveAttribute('inert', '')
  await expect.poll(async () => (await content.boundingBox())!.width).toBe(width)
  await page.getByRole('button', { name: 'Collapse sidebar' }).click()
  await page.getByRole('button', { name: 'Expand sidebar' }).click()
  await expect.poll(async () => (await content.boundingBox())!.width).toBe(width)
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
})

test('unified site status prompts guests after deployment and offers resource recovery', async ({ page }, testInfo) => {
  let deployment = 'old'
  let legacyRequests = 0
  await page.route('**/api/forum/unread-status', route => { legacyRequests++; return route.fulfill({ json: { code: 0, result: {} } }) })
  await page.route('**/api/forum/status', route => route.fulfill({ json: { code: 0, result: { deploymentId: deployment } } }))
  await page.goto('/?lang=en')
  await expect(page.locator('main').first()).toBeVisible()
  // The first request establishes the Vite-only baseline (production uses HTML meta).
  await page.waitForFunction(() => document.querySelector('#goose-site-status'))
  deployment = 'new'
  await page.evaluate(() => document.dispatchEvent(new Event('visibilitychange')))
  const notice = page.locator('#goose-site-status').getByRole('status')
  await expect(notice).toContainText('site has been updated')
  const bounds = await notice.boundingBox()
  expect(bounds!.width).toBeLessThanOrEqual(page.viewportSize()!.width)
  await page.screenshot({ path: testInfo.outputPath('site-update.png') })
  await notice.getByRole('button', { name: 'Later' }).click()
  await expect(notice).toHaveCount(0)
  await page.evaluate(() => window.dispatchEvent(new Event('vite:preloadError')))
  await expect(notice).toContainText('resources failed to load')
  expect(legacyRequests).toBe(0)
})

test('site update refresh respects an unfinished reply', async ({ page }) => {
  await page.route('**/api/forum/status', route => route.fulfill({ json: { code: 0, result: { deploymentId: 'old' } } }))
  await page.goto('/p/test/60?lang=en')
  await page.locator('[data-slot="topic-reply-float-boundary"]').getByRole('button', { name: 'Join the discussion' }).click()
  const panel = page.locator('.goose-composer-surface[data-state="open"] section')
  await panel.getByRole('radio', { name: 'Markdown', exact: true }).click()
  await panel.locator('textarea').fill('Keep my unfinished reply')
  await page.evaluate(() => window.dispatchEvent(new Event('vite:preloadError')))
  const dialog = page.waitForEvent('dialog').then(async warning => {
    expect(warning.type()).toBe('beforeunload')
    await warning.dismiss()
  })
  await page.locator('#goose-site-status').getByRole('button', { name: 'Refresh page' }).click()
  await dialog
  await expect(panel.locator('textarea')).toHaveValue('Keep my unfinished reply')
})

test('admin uses the shared update notice with a save reminder', async ({ page }) => {
  await page.route('**/api/forum/status', route => route.fulfill({ json: { code: 0, result: { deploymentId: 'old' } } }))
  await page.goto('/admin?lang=en')
  await expect(page.locator('#goose-admin-app')).not.toBeEmpty()
  await page.evaluate(() => window.dispatchEvent(new Event('vite:preloadError')))
  const dialog = page.waitForEvent('dialog').then(async warning => {
    expect(warning.type()).toBe('confirm')
    expect(warning.message()).toContain('unsaved changes')
    await warning.dismiss()
  })
  await page.locator('#goose-site-status').getByRole('button', { name: 'Refresh page' }).click()
  await dialog
  await expect(page.locator('#goose-site-status').getByRole('status')).toBeVisible()
})


test('public new topics hint only appears on the main list and refreshes on demand', async ({ page }, testInfo) => {
  let maximum = 10
  let homeRequests = 0
  let statusRequests = 0
  await page.route('**/__goose_page/**', route => {
    if (route.request().url().includes('/p/test/60')) return route.fulfill({ json: topicPage })
    homeRequests++
    return route.fulfill({ json: { ...homePage, props: { ...homePage.props, publicMaxTopicId: maximum } } })
  })
  await page.route('**/api/forum/status', route => {
    statusRequests++
    return route.fulfill({ json: { code: 0, result: { deploymentId: 'old', publicMaxTopicId: maximum } } })
  })
  await page.goto('/?lang=en')
  const hint = page.getByRole('button', { name: 'New topics available — click to view' })
  await expect(page.locator('main').first()).toBeVisible()
  await expect.poll(() => statusRequests).toBe(1)
  await expect(hint).toHaveCount(0)
  maximum = 12
  // Retry the visibility event if the initial request is still settling.
  await expect(async () => {
    await page.evaluate(() => document.dispatchEvent(new Event('visibilitychange')))
    await expect(hint).toBeVisible({ timeout: 1000 })
  }).toPass()
  const beforeRefresh = homeRequests
  const bounds = await hint.boundingBox()
  expect(bounds!.x).toBeGreaterThanOrEqual(0)
  expect(bounds!.x + bounds!.width).toBeLessThanOrEqual(page.viewportSize()!.width)
  await page.screenshot({ path: testInfo.outputPath('new-topics.png') })
  await hint.click()
  await expect(hint).toHaveCount(0)
  expect(homeRequests).toBeGreaterThan(beforeRefresh)
  maximum = 11
  await page.evaluate(() => document.dispatchEvent(new Event('visibilitychange')))
  await expect(hint).toHaveCount(0)
  await page.goto('/p/test/60?lang=en')
  maximum = 20
  await page.evaluate(() => document.dispatchEvent(new Event('visibilitychange')))
  await expect(hint).toHaveCount(0)
})

for (const mode of ['grid', 'carousel']) {
  test(`image gallery ${mode} uploads and survives visual editing`, async ({ page }, testInfo) => {
    await page.route('**/file/img-upload/init', route => route.fulfill({ json: { code: 0, result: { mode: 'proxy' } } }))
    let uploads = 0
    await page.route('**/file/img-upload', route => {
      uploads++
      return route.fulfill({ json: { code: 0, result: { url: `/static/pic/icons/icon-192.png?image=${uploads}` } } })
    })
    await page.goto('/publish?lang=en')
    await page.getByRole('combobox', { name: 'Image layout (multiple uploads)' }).click()
    await page.getByRole('option', { name: mode === 'grid' ? 'Image grid' : 'Image carousel', exact: true }).click()
    await page.locator('input[type="file"][accept="image/*"]').first().setInputFiles([
      fileURLToPath(new URL('../static/pic/icons/icon-192.png', import.meta.url)),
      fileURLToPath(new URL('../static/pic/icons/favicon-32.png', import.meta.url)),
    ])
    const gallery = page.locator(`[contenteditable="true"] [data-gallery="${mode}"]`)
    await expect(gallery.locator('img[src]')).toHaveCount(2)
    await page.getByRole('radio', { name: 'Markdown', exact: true }).click()
    const body = page.getByPlaceholder('Enter body text, Markdown supported; paste or drag images here')
    await expect(body).toHaveValue(new RegExp(`\\[gallery mode="${mode}"\\]`))
    await page.getByRole('radio', { name: 'Editor', exact: true }).click()
    await expect(gallery.locator('img[src]')).toHaveCount(2)
    await page.getByRole('button', { name: 'Preview', exact: true }).click()
    await expect(page.locator(`[data-slot="markdown-composer-preview"] [data-gallery="${mode}"] img`)).toHaveCount(2)
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
    await page.screenshot({ path: testInfo.outputPath(`gallery-${mode}-editor.png`), fullPage: true })
  })
}

test('image gallery carousel navigates and opens the image viewer', async ({ page }, testInfo) => {
  const images = ['First', 'Second', 'Third'].map(alt => `<p><img src="/static/pic/icons/icon-192.png?${alt}" alt="${alt}"></p>`).join('')
  const post = { ...topicPage.props.postStream.posts[0], renderedContent: `<div class="image-gallery image-gallery--carousel" data-gallery="carousel" tabindex="0">${images}</div>` }
  await page.route('**/__goose_page/**', route => route.fulfill({ json: { ...topicPage, props: { ...topicPage.props, postStream: { ...topicPage.props.postStream, posts: [post] } } } }))
  await page.goto('/p/test/60?lang=en')
  const gallery = page.locator('[data-gallery="carousel"]').first()
  await expect(gallery.locator('img')).toHaveCount(3)
  const expectFullSlide = async (index: number) => {
    const geometry = await gallery.evaluate((element, active) => {
      const container = element.getBoundingClientRect()
      const slide = element.children[active].getBoundingClientRect()
      return { width: Math.abs(slide.width - element.clientWidth), left: Math.abs(slide.left - container.left) }
    }, index)
    expect(geometry.width).toBeLessThan(1)
    expect(geometry.left).toBeLessThan(1)
  }
  await expectFullSlide(0)
  await page.getByRole('button', { name: 'Next image', exact: true }).click()
  await expect(page.getByText('2 / 3', { exact: true })).toBeVisible()
  await expectFullSlide(1)
  await page.getByRole('button', { name: 'Next image', exact: true }).click()
  await expect(page.getByText('3 / 3', { exact: true })).toBeVisible()
  await expectFullSlide(2)
  await page.getByRole('button', { name: 'Previous image', exact: true }).click()
  await page.getByRole('button', { name: 'Previous image', exact: true }).click()
  await expect(page.getByText('1 / 3', { exact: true })).toBeVisible()
  await gallery.getByRole('img', { name: 'First', exact: true }).click()
  await expect(page.getByRole('dialog')).toBeVisible()
  await page.keyboard.press('Escape')
  await expect(page.getByRole('dialog')).toHaveCount(0)
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  await page.screenshot({ path: testInfo.outputPath('gallery-carousel-post.png'), fullPage: true })
})

test('image gallery editing groups existing images and changes layout without source edits', async ({ page }, testInfo) => {
  await page.goto('/publish?lang=en')
  await page.getByRole('radio', { name: 'Markdown', exact: true }).click()
  const source = page.getByPlaceholder('Enter body text, Markdown supported; paste or drag images here')
  await source.fill('Keep this text\n\n![First](/static/pic/icons/icon-192.png)\n\n![Second](/static/pic/icons/icon-192.png)\n\nKeep this too')
  await page.getByRole('radio', { name: 'Editor', exact: true }).click()
  const editor = page.getByRole('textbox', { name: 'Write and format the body directly' })
  await editor.getByRole('img', { name: 'First', exact: true }).click()
  await page.getByRole('button', { name: 'Group nearby images', exact: true }).click()
  const group = editor.locator('.gallery-editor-block')
  await expect(group.locator('img[src]')).toHaveCount(2)
  await expect(group).not.toContainText('Keep this')
  await group.hover()
  await group.getByRole('button', { name: 'Image carousel', exact: true }).click()
  await expect(group.locator('[data-gallery="carousel"]')).toHaveCount(1)
  await expect(group.getByRole('button', { name: 'Image carousel', exact: true })).toHaveAttribute('aria-pressed', 'true')
  await page.screenshot({ path: testInfo.outputPath('gallery-edit-controls.png'), fullPage: true })
  await group.getByRole('button', { name: 'Ungroup images', exact: true }).click()
  await expect(group).toHaveCount(0)
  await editor.press('ControlOrMeta+z')
  await expect(group).toHaveCount(1)
  await editor.press('ControlOrMeta+Shift+z')
  await expect(group).toHaveCount(0)
  await expect(editor.locator('img[src]')).toHaveCount(2)
  await expect(editor).toContainText('Keep this text')
  await expect(editor).toContainText('Keep this too')
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
})

test('image gallery uses final layout at rest and supports reordering', async ({ page, isMobile }, testInfo) => {
  await page.goto('/publish?lang=en')
  await page.getByRole('radio', { name: 'Markdown', exact: true }).click()
  const source = page.getByPlaceholder('Enter body text, Markdown supported; paste or drag images here')
  await source.fill('[gallery mode="grid"]\n![First](/static/pic/icons/icon-192.png?1)\n![Second](/static/pic/icons/icon-192.png?2)\n![Third](/static/pic/icons/icon-192.png?3)\n[/gallery]')
  await page.getByRole('radio', { name: 'Editor', exact: true }).click()
  const editor = page.getByRole('textbox', { name: 'Write and format the body directly' })
  const group = editor.locator('.gallery-editor-block')
  await page.getByPlaceholder('Enter topic title').click()
  await page.mouse.move(0, 0)
  await expect(group.locator('.gallery-editor-toolbar')).toBeHidden()
  await expect(group.locator('[data-gallery]')).toHaveCSS('display', 'grid')
  await expect(group).toHaveCSS('padding', '0px')
  await page.screenshot({ path: testInfo.outputPath('gallery-at-rest.png'), fullPage: true })
  if (!isMobile) {
    await group.getByRole('img', { name: 'First', exact: true }).hover()
    const destination = await group.getByRole('img', { name: 'Second', exact: true }).boundingBox()
    await page.mouse.down()
    await page.mouse.move(destination!.x + destination!.width / 2, destination!.y + destination!.height / 2, { steps: 10 })
    await page.mouse.move(destination!.x + destination!.width / 2 + 2, destination!.y + destination!.height / 2)
    await page.mouse.up()
    await expect(group.locator('img[src]').nth(1)).toHaveAttribute('alt', 'First')
    await editor.press('ControlOrMeta+z')
    await expect(group.locator('img[src]').first()).toHaveAttribute('alt', 'First')
  }
  await group.getByRole('img', { name: 'First', exact: true }).click()
  await group.getByRole('button', { name: 'Move image later', exact: true }).click()
  await expect(group.locator('img[src]').nth(1)).toHaveAttribute('alt', 'First')
  await group.getByRole('button', { name: 'Image carousel', exact: true }).click()
  await expect(group.locator('[data-gallery]')).toHaveCSS('display', 'flex')
  await group.getByRole('button', { name: 'Next image', exact: true }).click()
  await expect(group.getByText('2 / 3', { exact: true })).toBeVisible()
  await page.getByRole('radio', { name: 'Markdown', exact: true }).click()
  await expect(source).toHaveValue(/Second[\s\S]*First[\s\S]*Third/)
  await expect(source).toHaveValue(/mode="carousel"/)
})

test('image gallery moves a selected image out and shares the post preview controls', async ({ page, isMobile }, testInfo) => {
  await page.goto('/publish?lang=en')
  await page.getByRole('radio', { name: 'Markdown', exact: true }).click()
  const source = page.getByPlaceholder('Enter body text, Markdown supported; paste or drag images here')
  await source.fill('[gallery mode="grid"]\n![First](/static/pic/icons/icon-192.png?1)\n![Second](/static/pic/icons/icon-192.png?2)\n![Third](/static/pic/icons/icon-192.png?3)\n[/gallery]')
  await page.getByRole('radio', { name: 'Editor', exact: true }).click()
  const editor = page.getByRole('textbox', { name: 'Write and format the body directly' })
  const group = editor.locator('.gallery-editor-block')
  await group.getByRole('img', { name: 'Second', exact: true }).click()
  await expect(group.locator('.gallery-image-selected img[src]')).toHaveAttribute('alt', 'Second')
  await expect(group).toHaveCSS('outline-style', 'dashed')
  await page.screenshot({ path: testInfo.outputPath('gallery-selected-boundary.png'), fullPage: true })
  await group.getByRole('button', { name: 'Move out of group', exact: true }).click()
  await expect(group.locator('img[src]')).toHaveCount(2)
  await expect(editor.getByRole('img', { name: 'Second', exact: true })).toHaveCount(1)
  expect(await editor.getByRole('img', { name: 'Second', exact: true }).evaluate(image => !!image.closest('[data-gallery]'))).toBe(false)
  await editor.press('ControlOrMeta+z')
  await expect(group.locator('img[src]')).toHaveCount(3)
  if (!isMobile) {
    await editor.locator(':scope > p').last().scrollIntoViewIfNeeded()
    await group.getByRole('img', { name: 'Third', exact: true }).hover()
    const destination = await editor.locator(':scope > p').last().boundingBox()
    await page.mouse.down()
    await page.mouse.move(destination!.x + 10, destination!.y + destination!.height / 2, { steps: 10 })
    await page.mouse.up()
    await expect(group.locator('img[src]')).toHaveCount(2)
    expect(await editor.getByRole('img', { name: 'Third', exact: true }).evaluate(image => !!image.closest('[data-gallery]'))).toBe(false)
    await editor.press('ControlOrMeta+z')
    await expect(group.locator('img[src]')).toHaveCount(3)
  }
  await group.getByRole('img', { name: 'First', exact: true }).click()
  await group.getByRole('button', { name: 'Image carousel', exact: true }).click()
  await page.getByRole('button', { name: 'Preview', exact: true }).click()
  const preview = page.locator('[data-slot="markdown-composer-preview"]')
  await expect(preview.locator('[data-content-variant="post"]')).toHaveCount(1)
  await preview.getByRole('button', { name: 'Next image', exact: true }).click()
  await expect(preview.getByText('2 / 3', { exact: true })).toBeVisible()
  await preview.getByRole('img', { name: 'Second', exact: true }).click()
  await expect(page.getByRole('dialog', { name: 'Image preview' })).toBeVisible()
  await page.keyboard.press('Escape')
  await expect(page.getByRole('dialog')).toHaveCount(0)
  await expect.poll(() => preview.locator('[data-gallery="carousel"]').evaluate(element => {
    const container = element.getBoundingClientRect()
    const active = element.children[1].getBoundingClientRect()
    return Math.abs(active.left - container.left)
  })).toBeLessThan(1)
  await page.screenshot({ path: testInfo.outputPath('gallery-shared-preview.png'), fullPage: true })
})

test('dragging an existing editor image with a browser file payload does not upload it', async ({ page }) => {
  let uploads = 0
  await page.route('**/file/img-upload/init', route => { uploads++; return route.fulfill({ json: { code: 0, result: { mode: 'proxy' } } }) })
  await page.goto('/publish?lang=en')
  await page.getByRole('radio', { name: 'Markdown', exact: true }).click()
  await page.getByPlaceholder('Enter body text, Markdown supported; paste or drag images here').fill('![Existing](/static/pic/icons/icon-192.png)')
  await page.getByRole('radio', { name: 'Editor', exact: true }).click()
  const editor = page.getByRole('textbox', { name: 'Write and format the body directly' })
  const payload = await page.evaluateHandle(() => {
    const data = new DataTransfer()
    data.items.add(new File(['image'], 'download', { type: 'image/png' }))
    return data
  })
  await editor.getByRole('img', { name: 'Existing', exact: true }).dispatchEvent('dragstart', { dataTransfer: payload })
  await editor.dispatchEvent('dragover', { dataTransfer: payload })
  await expect(page.getByText('Drop images to upload', { exact: true })).toHaveCount(0)
  await editor.dispatchEvent('drop', { dataTransfer: payload })
  await editor.dispatchEvent('dragend', { dataTransfer: payload })
  await page.getByRole('radio', { name: 'Markdown', exact: true }).click()
  await expect(page.getByText(/download:/)).toHaveCount(0)
  expect(uploads).toBe(0)
  await page.route('**/file/img-upload', route => route.fulfill({ json: { code: 0, result: { url: '/static/pic/icons/icon-192.png?external' } } }))
  const external = await page.evaluateHandle(async () => {
    const response = await fetch('/static/pic/icons/icon-192.png')
    const data = new DataTransfer()
    data.items.add(new File([await response.arrayBuffer()], 'external.png', { type: 'image/png' }))
    return data
  })
  await page.getByPlaceholder('Enter body text, Markdown supported; paste or drag images here').dispatchEvent('drop', { dataTransfer: external })
  await expect(page.getByPlaceholder('Enter body text, Markdown supported; paste or drag images here')).toHaveValue(/external/)
  expect(uploads).toBe(1)
})

test('image gallery splits inline drop images and hides the carousel scrollbar', async ({ page }) => {
  await page.goto('/publish?lang=en')
  await page.getByRole('radio', { name: 'Markdown', exact: true }).click()
  const source = page.getByPlaceholder('Enter body text, Markdown supported; paste or drag images here')
  await source.fill('[gallery mode="carousel"]\n![First](/static/pic/icons/icon-192.png?1)\n![Second](/static/pic/icons/icon-192.png?2)\n[/gallery]')
  await page.getByRole('radio', { name: 'Editor', exact: true }).click()
  const editor = page.getByRole('textbox', { name: 'Write and format the body directly' })
  const gallery = editor.locator('[data-gallery="carousel"]')
  await gallery.getByRole('img', { name: 'First', exact: true }).scrollIntoViewIfNeeded()
  const target = await gallery.getByRole('img', { name: 'First', exact: true }).boundingBox()
  const data = await page.evaluateHandle(() => {
    const transfer = new DataTransfer()
    transfer.setData('text/html', '<img src="/static/pic/icons/icon-192.png?added" alt="Added">')
    return transfer
  })
  await gallery.dispatchEvent('drop', { dataTransfer: data, clientX: target!.x + target!.width - 2, clientY: target!.y + target!.height / 2 })
  await expect(gallery.locator('img[src]')).toHaveCount(3)
  expect(await gallery.locator(':scope > p').evaluateAll(items => items.every(item => item.querySelectorAll('img[src]').length === 1))).toBe(true)
  await editor.press('ControlOrMeta+z')
  await expect(gallery.locator('img[src]')).toHaveCount(2)
  await editor.press('ControlOrMeta+Shift+z')
  await expect(gallery.locator('img[src]')).toHaveCount(3)
  await page.getByRole('radio', { name: 'Markdown', exact: true }).click()
  const lines = (await source.inputValue()).split('\n').filter(line => line.includes('!['))
  expect(lines).toHaveLength(3)
  expect(lines.every(line => line.match(/!\[/g)?.length === 1)).toBe(true)
  await page.getByRole('button', { name: 'Preview', exact: true }).click()
  const preview = page.locator('[data-slot="markdown-composer-preview"]')
  await expect(preview.locator('[data-gallery="carousel"]')).toHaveCSS('scrollbar-width', 'none')
  await preview.getByRole('button', { name: 'Next image', exact: true }).click()
  await expect(preview.getByText('2 / 3', { exact: true })).toBeVisible()
})

test('image gallery drag cancels safely and scrolls the carousel while held at its edge', async ({ page, isMobile }, testInfo) => {
  test.skip(isMobile, 'Mouse drag behavior; touch reordering uses the existing move controls')
  await page.goto('/publish?lang=en')
  await page.getByRole('radio', { name: 'Markdown', exact: true }).click()
  await page.getByPlaceholder('Enter body text, Markdown supported; paste or drag images here').fill('[gallery mode="grid"]\n![First](/static/pic/icons/icon-192.png?1)\n![Second](/static/pic/icons/icon-192.png?2)\n![Third](/static/pic/icons/icon-192.png?3)\n[/gallery]')
  await page.getByRole('radio', { name: 'Editor', exact: true }).click()
  const editor = page.getByRole('textbox', { name: 'Write and format the body directly' })
  const group = editor.locator('.gallery-editor-block')
  const gallery = group.locator('[data-gallery]')
  const order = () => gallery.locator('img[src]').evaluateAll(images => images.map(image => image.getAttribute('alt')))
  await group.getByRole('img', { name: 'First', exact: true }).click({ clickCount: 3 })
  await expect(group).not.toHaveClass(/ProseMirror-selectednode/)
  await group.getByRole('img', { name: 'First', exact: true }).hover()
  const second = await group.getByRole('img', { name: 'Second', exact: true }).boundingBox()
  await page.mouse.down()
  await page.mouse.move(second!.x + second!.width / 2, second!.y + second!.height / 2, { steps: 10 })
  await expect(page.locator('.gallery-drag-preview')).toBeVisible()
  await expect(gallery.locator(':scope > p').first()).toHaveCSS('opacity', '0.25')
  const previewBounds = await page.locator('.gallery-drag-preview').boundingBox()
  expect(Math.abs(previewBounds!.x - second!.x)).toBeLessThan(5)
  await page.screenshot({ path: testInfo.outputPath('gallery-drag-preview.png') })
  await page.keyboard.press('Escape')
  await expect(page.locator('.gallery-drag-preview')).toHaveCount(0)
  await page.mouse.up()
  expect(await order()).toEqual(['First', 'Second', 'Third'])
  await expect(group).not.toHaveClass(/is-dragging/)
  await group.getByRole('img', { name: 'First', exact: true }).hover()
  await page.mouse.down()
  await page.mouse.move(second!.x + second!.width / 2, second!.y + second!.height / 2, { steps: 10 })
  await expect(page.locator('.gallery-drag-preview')).toBeVisible()
  await page.evaluate(() => window.dispatchEvent(new Event('blur')))
  await expect(page.locator('.gallery-drag-preview')).toHaveCount(0)
  await page.mouse.up()
  expect(await order()).toEqual(['First', 'Second', 'Third'])
  await group.getByRole('img', { name: 'First', exact: true }).click()
  await group.getByRole('button', { name: 'Image carousel', exact: true }).click()
  await group.getByRole('img', { name: 'First', exact: true }).hover()
  const next = await group.getByRole('button', { name: 'Next image', exact: true }).boundingBox()
  await page.mouse.down()
  await page.mouse.move(next!.x + next!.width / 2, next!.y + next!.height / 2, { steps: 10 })
  await page.mouse.up()
  expect(await order()).toEqual(['First', 'Second', 'Third'])
  await group.getByRole('img', { name: 'First', exact: true }).hover()
  const bounds = await gallery.boundingBox()
  await page.mouse.down()
  await page.mouse.move(bounds!.x + bounds!.width - 10, bounds!.y + bounds!.height / 2, { steps: 10 })
  await expect.poll(() => gallery.evaluate(element => element.scrollWidth - element.clientWidth - element.scrollLeft)).toBeLessThan(1)
  await page.mouse.move(bounds!.x + bounds!.width / 2, bounds!.y + bounds!.height / 2)
  await page.mouse.up()
  expect(await order()).toEqual(['Second', 'Third', 'First'])
  await expect(page.locator('.gallery-drag-preview')).toHaveCount(0)
  await expect(group).not.toHaveClass(/is-dragging/)
  await expect(gallery).toHaveCSS('scroll-snap-type', 'x mandatory')
  await editor.press('ControlOrMeta+z')
  expect(await order()).toEqual(['First', 'Second', 'Third'])
})

test('image gallery dnd isolates groups and cleans up when a two-image group dissolves', async ({ page, isMobile }) => {
  test.skip(isMobile, 'Mouse drag lifecycle; mobile uses move controls')
  const errors: string[] = []
  page.on('pageerror', error => errors.push(error.message))
  page.on('console', message => { if (message.type() === 'error' && !message.text().includes('401 (Unauthorized)')) errors.push(message.text()) })
  await page.goto('/publish?lang=en')
  await page.getByRole('radio', { name: 'Markdown', exact: true }).click()
  const source = page.getByPlaceholder('Enter body text, Markdown supported; paste or drag images here')
  const markdown = '[gallery mode="grid"]\n![First](/static/pic/icons/icon-192.png?1)\n![Second](/static/pic/icons/icon-192.png?2)\n[/gallery]\n\nBetween groups\n\n[gallery mode="grid"]\n![Third](/static/pic/icons/icon-192.png?3)\n![Fourth](/static/pic/icons/icon-192.png?4)\n[/gallery]'
  await source.fill(markdown)
  await page.getByRole('radio', { name: 'Editor', exact: true }).click()
  const editor = page.getByRole('textbox', { name: 'Write and format the body directly' })
  const groups = editor.locator('[data-gallery]')
  await editor.getByRole('img', { name: 'Third', exact: true }).scrollIntoViewIfNeeded()
  await editor.getByRole('img', { name: 'First', exact: true }).hover()
  let destination = await editor.getByRole('img', { name: 'Third', exact: true }).boundingBox()
  await page.mouse.down()
  await page.mouse.move(destination!.x + destination!.width / 2, destination!.y + destination!.height / 2, { steps: 10 })
  await page.mouse.up()
  await expect(groups).toHaveCount(2)
  await expect(groups.first().getByRole('img', { name: 'First', exact: true })).toHaveCount(1)
  await editor.getByRole('img', { name: 'First', exact: true }).hover()
  destination = await editor.getByText('Between groups', { exact: true }).boundingBox()
  await page.mouse.down()
  await page.mouse.move(destination!.x + 10, destination!.y + destination!.height / 2, { steps: 10 })
  await page.mouse.up()
  await expect(groups).toHaveCount(1)
  await expect(page.locator('.gallery-drag-preview')).toHaveCount(0)
  await editor.press('ControlOrMeta+z')
  await expect(groups).toHaveCount(2)
  await editor.getByRole('img', { name: 'First', exact: true }).hover()
  destination = await editor.getByRole('img', { name: 'Second', exact: true }).boundingBox()
  await page.mouse.down()
  await page.mouse.move(destination!.x + destination!.width / 2, destination!.y + destination!.height / 2, { steps: 10 })
  await expect(page.locator('.gallery-drag-preview')).toBeVisible()
  // Deleting the document unmounts the active NodeView mid-gesture.
  await editor.press('ControlOrMeta+a')
  await editor.press('Backspace')
  await expect(page.locator('.gallery-drag-preview')).toHaveCount(0)
  await page.mouse.up()
  await editor.press('ControlOrMeta+z')
  await expect(groups).toHaveCount(2)
  await page.getByRole('radio', { name: 'Markdown', exact: true }).click()
  await expect(source).toHaveValue(/First[\s\S]*Second[\s\S]*Third[\s\S]*Fourth/)
  expect(errors).toEqual([])
})
