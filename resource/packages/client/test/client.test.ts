import { describe, expect, it, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import {
  createGooseClient,
  GooseClientError,
  GooseProtocolError,
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
    expect(pageComponents).toHaveLength(20)
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

  it('sets JSON accept headers for GET domain APIs', async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ code: 0, result: [] }))
    const client = createGooseClient({ fetch: fetchMock })

    await client.api.accessGroups.list()

    expect(new Headers(fetchMock.mock.calls[0]?.[1]?.headers).get('Accept')).toBe('application/json')
  })

  it('rejects malformed API envelopes', async () => {
    const client = createGooseClient({ fetch: async () => jsonResponse(null) })

    await expect(client.api.notifications.unread()).rejects.toBeInstanceOf(GooseProtocolError)
  })
})
