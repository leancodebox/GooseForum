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

  it('reads and validates the embedded initial payload', () => {
    const client = createGooseClient({ fetch: async () => jsonResponse({}) })
    const root = {
      querySelector: () => ({ textContent: JSON.stringify(pagePayload()) }),
    } as unknown as ParentNode

    expect(client.readInitialPayload(root).component).toBe('home.index')
  })

  it('keeps the public page component list unique', () => {
    expect(pageComponents).toHaveLength(17)
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
})
