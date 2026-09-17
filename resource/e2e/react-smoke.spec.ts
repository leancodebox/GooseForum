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

const responsiveTopic = {
  ...publishedTopic,
  id: 101,
  title: 'How should a long topic title adapt cleanly across a compact mobile forum list?',
  description: 'The desktop summary remains unchanged.',
  url: '/p/test/60',
  author: { id: 12, username: 'responsive-author', avatarUrl: '' },
  participants: [{ id: 12, username: 'responsive-author', avatarUrl: '' }],
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
  await page.evaluate(() => {
    const state = window as unknown as { __gooseOldListTopFlash: number }
    state.__gooseOldListTopFlash = 0
    window.addEventListener('scroll', () => {
      const topicRow = document.querySelector('[data-slot="topic-row"]')
      if (window.scrollY === 0 && topicRow?.getClientRects().length) {
        state.__gooseOldListTopFlash++
      }
    }, { passive: true })
  })

  await deepTitle.click()
  await page.waitForTimeout(100)
  expect(await page.evaluate(() => window.scrollY)).toBeGreaterThan(0)
  await expect(page.getByRole('heading', { level: 1, name: 'Topic detail' })).toBeVisible()
  expect(await page.evaluate(() => window.scrollY)).toBe(0)
  expect(await page.evaluate(() =>
    (window as unknown as { __gooseOldListTopFlash: number })
      .__gooseOldListTopFlash,
  )).toBe(0)
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
