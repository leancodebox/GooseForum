import { describe, expect, it, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import {
  createGooseClient,
  formatCompactNumber,
  GooseClientError,
  GooseProtocolError,
  authResources,
  normalizeLocale,
  pageComponents,
} from '../src/index.js'

function pagePayload(overrides: Record<string, unknown> = {}) {
  return {
    component: 'home.index',
    props: { topics: [] },
    meta: { title: 'GooseForum' },
    layout: {},
    url: '/',
    version: '1.0',
    ...overrides,
  }
}

function jsonResponse(value: unknown, status = 200) {
  return new Response(JSON.stringify(value), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

describe('i18n contracts', () => {
  it.each([
    ['zh-CN', 'zh'],
    ['EN_us', 'en'],
    ['ja,zh;q=0.9', 'ja'],
    ['it-IT', 'it'],
    ['unknown', undefined],
  ])('normalizes %s to %s', (input, expected) => {
    expect(normalizeLocale(input)).toBe(expected)
  })

  it('keeps every auth locale structurally complete', () => {
    expect(Object.keys(authResources)).toEqual(['zh', 'en', 'ja', 'it'])
    expect(authResources.en.validation.loginRequired).toBeTruthy()
    expect(authResources.ja.server.passwordResetMailQueued).toBeTruthy()
  })
})

describe('format contracts', () => {
  it.each([
    [999, '999'],
    [1_250, '1.3k'],
    [12_500, '13k'],
    [1_250_000, '1.3m'],
  ])('formats %d as %s', (value, expected) => {
    expect(formatCompactNumber(value)).toBe(expected)
  })
})

describe('page client', () => {
  it('requests page payloads with the protocol headers', async () => {
    const fetchMock = vi.fn(async () => jsonResponse(pagePayload()))
    const client = createGooseClient({ baseURL: 'https://forum.example', fetch: fetchMock })

    const payload = await client.pages.fetch('/?sort=hot')

    expect(payload.component).toBe('home.index')
    expect(fetchMock).toHaveBeenCalledOnce()
    const [url, init] = fetchMock.mock.calls[0]!
    expect(url).toBe('https://forum.example/?sort=hot')
    expect(new Headers(init?.headers).get('X-Goose-Page')).toBe('true')
    expect(new Headers(init?.headers).get('Accept')).toBe('application/json')
  })

  it('accepts a structured 404 page', async () => {
    const client = createGooseClient({
      fetch: async () => jsonResponse(pagePayload({ component: 'error.index' }), 404),
    })

    await expect(client.pages.fetch('/missing')).resolves.toMatchObject({ component: 'error.index' })
  })

  it('rejects incompatible payload versions', async () => {
    const client = createGooseClient({ fetch: async () => jsonResponse(pagePayload({ version: '2.0' })) })

    await expect(client.pages.fetch('/')).rejects.toBeInstanceOf(GooseProtocolError)
  })

  it('rejects unknown page components', async () => {
    const client = createGooseClient({
      fetch: async () => jsonResponse(pagePayload({ component: 'plugin.unknown' })),
    })

    await expect(client.pages.fetch('/')).rejects.toBeInstanceOf(GooseProtocolError)
  })

  it('supports extension page components without weakening the default client', () => {
    const client = createGooseClient({ pages: { components: ['plugin.dashboard'] } })

    expect(client.pages.parse(pagePayload({ component: 'plugin.dashboard' }))).toMatchObject({
      component: 'plugin.dashboard',
    })
  })

  it('can accept future components as generic payloads when explicitly enabled', () => {
    const client = createGooseClient({ pages: { allowUnknownComponents: true } })

    expect(client.pages.parse(pagePayload({ component: 'future.page' })).component).toBe('future.page')
  })

  it.each([
    ['props', undefined, 'props'],
    ['meta', undefined, 'meta'],
    ['layout', undefined, 'layout'],
    ['url', undefined, 'url'],
  ])('rejects payloads missing the %s envelope field', (field, value, expected) => {
    const client = createGooseClient()

    expect(() => client.pages.parse(pagePayload({ [field]: value }))).toThrow(expected)
  })

  it('runs application-specific payload validation', () => {
    const validate = vi.fn(() => { throw new Error('invalid site contract') })
    const client = createGooseClient({ pages: { validate } })

    expect(() => client.pages.parse(pagePayload())).toThrow('invalid site contract')
    expect(validate).toHaveBeenCalledOnce()
  })

  it('reads and validates the embedded initial payload', () => {
    const client = createGooseClient({ fetch: async () => jsonResponse({}) })
    const root = {
      querySelector: () => ({ textContent: JSON.stringify(pagePayload()) }),
    } as unknown as ParentNode

    expect(client.readInitialPayload(root).component).toBe('home.index')
  })

  it('keeps the public page component list unique', () => {
		expect(pageComponents).toHaveLength(21)
    expect(new Set(pageComponents).size).toBe(pageComponents.length)
  })

  it('matches the public page components declared by the Go server', () => {
    const source = readFileSync(
      new URL('../../../../app/http/controllers/forum/page_component.go', import.meta.url),
      'utf8',
    )
    const serverComponents = [...source.matchAll(/PageComponent\w+\s+PageComponent\s*=\s*"([^"]+)"/g)]
      .map((match) => match[1])
      .filter((component) => component !== 'admin.shell')

    expect(serverComponents).toEqual([...pageComponents])
  })
})

describe('API client', () => {
  it('serializes JSON and unwraps the result envelope', async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ code: 0, result: { id: 42 } }))
    const client = createGooseClient({ baseURL: 'https://forum.example', fetch: fetchMock })

    const result = await client.request<{ id: number }>('/api/example', {
      method: 'POST',
      json: { title: 'hello' },
    })

    expect(result).toEqual({ id: 42 })
    const [, init] = fetchMock.mock.calls[0]!
    expect(new Headers(init?.headers).get('Content-Type')).toBe('application/json')
    expect(init?.body).toBe(JSON.stringify({ title: 'hello' }))
  })

  it('preserves structured server errors', async () => {
    const client = createGooseClient({
      fetch: async () => jsonResponse({
        code: 1001,
        messageCode: 'topic.notFound',
        params: { id: 7 },
      }),
    })

    const error = await client.request('/api/example').catch((reason) => reason)
    expect(error).toBeInstanceOf(GooseClientError)
    expect(error).toMatchObject({ code: 1001, messageCode: 'topic.notFound', params: { id: 7 } })
  })

  it('preserves successful message metadata when requested', async () => {
    const client = createGooseClient({
      fetch: async () => jsonResponse({
        code: 0,
        messageCode: 'auth.register.success',
        params: { email: 'user@example.com' },
        result: 42,
      }),
    })

    await expect(client.requestWithMeta<number>('/api/example')).resolves.toEqual({
      value: 42,
      message: undefined,
      messageCode: 'auth.register.success',
      params: { email: 'user@example.com' },
    })
  })

  it('exposes typed domain APIs with stable routes and payloads', async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ code: 0, result: true }))
    const client = createGooseClient({ baseURL: 'https://forum.example', fetch: fetchMock })

    await client.api.topics.like(42, 1)
    await client.api.notifications.list('unread', 10, 30)
    await client.api.themes.publish()

    expect(fetchMock).toHaveBeenCalledTimes(3)
    expect(fetchMock.mock.calls[0]?.[0]).toBe('https://forum.example/api/forum/topics/like')
    expect(fetchMock.mock.calls[0]?.[1]?.body).toBe(JSON.stringify({ topicId: 42, action: 1 }))
    expect(fetchMock.mock.calls[1]?.[0]).toBe('https://forum.example/api/forum/notifications?filter=unread&cursor=10&limit=30')
    expect(fetchMock.mock.calls[2]?.[0]).toBe('https://forum.example/api/admin/publish-site-theme')
  })

  it('exposes admin category APIs without coupling them to a UI framework', async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ code: 0, result: [] }))
    const client = createGooseClient({ baseURL: 'https://forum.example', fetch: fetchMock })

    await client.admin.categories.list()
    await client.admin.categories.saveAccess(7, [{ accessGroupId: 3, level: 2 }])
    await client.admin.categories.addModerator(7, { userId: 42 })

    expect(fetchMock.mock.calls.map(([url]) => url)).toEqual([
      'https://forum.example/api/admin/category-list',
      'https://forum.example/api/admin/category-access/save',
      'https://forum.example/api/admin/category-moderator-add',
    ])
    expect(fetchMock.mock.calls[1]?.[1]?.body).toBe(JSON.stringify({
      categoryId: 7,
      grants: [{ accessGroupId: 3, level: 2 }],
    }))
    expect(fetchMock.mock.calls[2]?.[1]?.body).toBe(JSON.stringify({ categoryId: 7, userId: 42 }))
  })

  it('exposes access-group administration with stable request bodies', async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ code: 0, result: 7 }))
    const client = createGooseClient({ baseURL: 'https://forum.example', fetch: fetchMock })

    await client.admin.accessGroups.save({ name: 'Staff', joinMode: 'invite_only', status: 1 })
    await client.admin.accessGroups.saveMember({ groupId: 7, username: 'goose', memberRole: 'manager' })
    await client.admin.accessGroups.reviewApplication(7, 19, true)

    expect(fetchMock.mock.calls.map(([url]) => url)).toEqual([
      'https://forum.example/api/admin/access-group/save',
      'https://forum.example/api/admin/access-group/member-save',
      'https://forum.example/api/admin/access-group/application-review',
    ])
    expect(fetchMock.mock.calls[2]?.[1]?.body).toBe(JSON.stringify({ groupId: 7, memberId: 19, approve: true }))
  })

  it('exposes role administration with stable request bodies', async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ code: 0, result: [] }))
    const client = createGooseClient({ baseURL: 'https://forum.example', fetch: fetchMock })

    await client.admin.roles.list()
    await client.admin.roles.permissions()
    await client.admin.roles.save({ id: 4, roleName: 'Moderator', permissions: [2, 4] })
    await client.admin.roles.delete(4)

    expect(fetchMock.mock.calls.map(([url]) => url)).toEqual([
      'https://forum.example/api/admin/role-list',
      'https://forum.example/api/admin/get-permission-list',
      'https://forum.example/api/admin/role-save',
      'https://forum.example/api/admin/role-delete',
    ])
    expect(fetchMock.mock.calls[2]?.[1]?.body).toBe(JSON.stringify({ id: 4, roleName: 'Moderator', permissions: [2, 4] }))
  })

  it('exposes user editing and badge administration', async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ code: 0, result: [] }))
    const client = createGooseClient({ baseURL: 'https://forum.example', fetch: fetchMock })

    await client.admin.users.edit({ userId: 9, status: 0, validate: 1, roleId: 3 })
    await client.admin.users.badgeOptions(9)
    await client.admin.users.saveBadges(9, ['helper'])

    expect(fetchMock.mock.calls.map(([url]) => url)).toEqual([
      'https://forum.example/api/admin/user-edit',
      'https://forum.example/api/admin/user-badge-options',
      'https://forum.example/api/admin/save-user-badges',
    ])
    expect(fetchMock.mock.calls[2]?.[1]?.body).toBe(JSON.stringify({ userId: 9, badgeCodes: ['helper'] }))
  })

  it('exposes topic moderation with stable routes and request bodies', async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ code: 0, result: true }))
    const client = createGooseClient({ baseURL: 'https://forum.example', fetch: fetchMock })

    await client.admin.topics.setCategories(12, [3, 7])
    await client.admin.topics.setPin(12, 20)
    await client.admin.topics.review('post', { id: 44, version: 2, action: 'reject', reason: 'spam' })

    expect(fetchMock.mock.calls.map(([url]) => url)).toEqual([
      'https://forum.example/api/admin/topics/categories-edit',
      'https://forum.example/api/admin/topics/pin-edit',
      'https://forum.example/api/admin/posts/review',
    ])
    expect(fetchMock.mock.calls[0]?.[1]?.body).toBe(JSON.stringify({ topicId: 12, categoryId: [3, 7] }))
    expect(fetchMock.mock.calls[2]?.[1]?.body).toBe(JSON.stringify({ id: 44, version: 2, action: 'reject', reason: 'spam' }))
  })

  it('exposes page configuration and image upload contracts', async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ code: 0, result: [] }))
    const client = createGooseClient({ baseURL: 'https://forum.example', fetch: fetchMock })

    await client.admin.pages.saveLinks([{ name: 'Friends', links: [] }])
    await client.admin.pages.saveSponsors({ sponsors: { level0: [], level1: [], level2: [], level3: [] }, content: { title: 'Sponsors', description: '' }, contact: { title: '', description: '', buttonText: '', buttonLink: '' }, rules: [] })
    await client.admin.pages.uploadImage(new File(['logo'], 'logo.png', { type: 'image/png' }))

    expect(fetchMock.mock.calls.map(([url]) => url)).toEqual([
      'https://forum.example/api/admin/save-friend-links',
      'https://forum.example/api/admin/save-sponsors',
      'https://forum.example/api/admin/img-upload',
    ])
    expect(fetchMock.mock.calls[0]?.[1]?.body).toBe(JSON.stringify({ linksInfo: [{ name: 'Friends', links: [] }] }))
    expect(fetchMock.mock.calls[2]?.[1]?.body).toBeInstanceOf(FormData)
    expect(new Headers(fetchMock.mock.calls[2]?.[1]?.headers).has('Content-Type')).toBe(false)
  })

  it('exposes badge and file-resource administration', async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ code: 0, result: [] }))
    const client = createGooseClient({ baseURL: 'https://forum.example', fetch: fetchMock })
    await client.admin.assets.deleteBadge('helper')
    await client.admin.assets.files({ page: 2, pageSize: 20 })
    expect(fetchMock.mock.calls.map(([url]) => url)).toEqual([
      'https://forum.example/api/admin/badge-delete',
      'https://forum.example/api/admin/file-resources',
    ])
    expect(fetchMock.mock.calls[1]?.[1]?.body).toBe(JSON.stringify({ page: 2, pageSize: 20 }))
  })

  it('exposes paginated audit records', async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ code: 0, result: { list: [], page: 1, total: 0 } }))
    const client = createGooseClient({ baseURL: 'https://forum.example', fetch: fetchMock })
    await client.admin.audit.records({ page: 3, pageSize: 20 })
    expect(fetchMock.mock.calls[0]?.[0]).toBe('https://forum.example/api/admin/opt-record-page')
    expect(fetchMock.mock.calls[0]?.[1]?.body).toBe(JSON.stringify({ page: 3, pageSize: 20 }))
  })

  it('exposes site and chrome settings contracts', async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ code: 0, result: {} }))
    const client = createGooseClient({ baseURL: 'https://forum.example', fetch: fetchMock })
    await client.admin.settings.saveSite({ siteName:'GooseForum',siteUrl:'https://forum.example',siteLogo:'',siteEmail:'',siteDescription:'',siteKeywords:'' })
    await client.admin.settings.saveChrome({ header:[],mainMenu:[],resources:[],sidebarGroups:[] })
    expect(fetchMock.mock.calls.map(([url])=>url)).toEqual(['https://forum.example/api/admin/save-site-settings','https://forum.example/api/admin/save-site-chrome'])
  })

  it('sets JSON accept headers for GET domain APIs', async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ code: 0, result: [] }))
    const client = createGooseClient({ fetch: fetchMock })

    await client.api.accessGroups.list()

    expect(new Headers(fetchMock.mock.calls[0]?.[1]?.headers).get('Accept')).toBe('application/json')
  })

  it('lists and revokes OIDC grants through the user API', async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ code: 0, result: [] }))
    const client = createGooseClient({ baseURL: 'https://forum.example', fetch: fetchMock })

    await client.api.users.oidcGrants()
    await client.api.users.revokeOIDCGrant('gf_wiki')

    expect(fetchMock.mock.calls[0]?.[0]).toBe('https://forum.example/api/oidc/grants')
    expect(fetchMock.mock.calls[1]?.[0]).toBe('https://forum.example/api/oidc/grants/revoke')
    expect(fetchMock.mock.calls[1]?.[1]?.body).toBe(JSON.stringify({ clientId: 'gf_wiki' }))
  })

  it('uses raw JSON contracts for OIDC consent endpoints', async () => {
    const fetchMock = vi.fn(async (input: string | URL, init?: RequestInit) => {
      if (init?.method === 'POST') return jsonResponse({ redirect_url: 'https://client.example/callback' })
      return jsonResponse({
        client: { id: 'client-id', name: 'Example', public: true },
        scopes: ['openid'],
        expires_at: '2026-09-14T12:00:00Z',
      })
    })
    const client = createGooseClient({ baseURL: 'https://forum.example', fetch: fetchMock })

    await expect(client.api.oidc.consentDetails('interaction-id')).resolves.toMatchObject({
      client: { id: 'client-id' },
    })
    await expect(client.api.oidc.consentDecision('interaction-id', 'approve')).resolves.toEqual({
      redirect_url: 'https://client.example/callback',
    })

    expect(fetchMock.mock.calls[0]?.[0]).toBe('https://forum.example/oauth2/consent/details?interaction=interaction-id')
    expect(new Headers(fetchMock.mock.calls[0]?.[1]?.headers).get('Accept')).toBe('application/json')
    expect(fetchMock.mock.calls[1]?.[0]).toBe('https://forum.example/oauth2/consent')
    expect(fetchMock.mock.calls[1]?.[1]?.body).toBe(JSON.stringify({ interaction: 'interaction-id', decision: 'approve' }))
  })

  it('rejects malformed API envelopes', async () => {
    const client = createGooseClient({ fetch: async () => jsonResponse(null) })

    await expect(client.api.notifications.unread()).rejects.toBeInstanceOf(GooseProtocolError)
  })
})
