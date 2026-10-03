import AxeBuilder from '@axe-core/playwright'
import { expect, test, type Page } from '@playwright/test'

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
      await route.fulfill({ json: { code: 0, result: chatMessages } })
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
  await expect(page.getByRole('heading', { name: 'OAuth 登录' })).toBeVisible()

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
  await expect(panel.locator('[contenteditable="true"]')).toHaveText('A draft that must survive transitions')
  await panel.locator('header').getByRole('button', { name: 'Close', exact: true }).click()
  await expect(page.locator('.goose-composer-surface')).toHaveCount(0)
  await reply.click()
  await expect(panel.locator('[contenteditable="true"]')).toHaveText('')
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
