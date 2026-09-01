import { GooseClientError, GooseProtocolError } from './error.js'

export interface ApiEnvelope<T> {
  code?: number
  message?: string
  messageCode?: string
  params?: Record<string, unknown>
  result?: T
  data?: T
}

export interface GooseApiResult<T> {
  value: T
  message?: string
  messageCode?: string
  params?: Record<string, unknown>
}

export interface GooseClientOptions {
  baseURL?: string
  fetch?: typeof globalThis.fetch
  payloadMajorVersion?: number
}

export interface GooseRequestInit extends Omit<RequestInit, 'body'> {
  body?: BodyInit | null
  json?: unknown
  query?: URLSearchParams | Record<string, string | number | boolean | null | undefined>
}

export interface GooseHttpClient {
  readonly payloadMajorVersion: number
  fetch(input: string | URL, init?: RequestInit): Promise<Response>
  request<T>(path: string | URL, init?: GooseRequestInit): Promise<T>
  requestWithMeta<T>(path: string | URL, init?: GooseRequestInit): Promise<GooseApiResult<T>>
  resolve(path: string | URL): string
}

function defaultFetch(input: string | URL | Request, init?: RequestInit) {
  if (typeof globalThis.fetch !== 'function') {
    throw new GooseClientError('No fetch implementation is available')
  }
  return globalThis.fetch(input, init)
}

function appendQuery(url: string, query: GooseRequestInit['query']) {
  if (!query) return url
  const absolute = /^https?:\/\//i.test(url)
  const target = absolute ? new URL(url) : new URL(url, 'http://gooseforum.local')
  const entries = query instanceof URLSearchParams ? query.entries() : Object.entries(query)
  for (const [key, value] of entries) {
    if (value !== undefined && value !== null) target.searchParams.set(key, String(value))
  }
  if (absolute) return target.toString()
  return `${target.pathname}${target.search}${target.hash}`
}

async function readEnvelope<T>(response: Response): Promise<GooseApiResult<T>> {
  const envelope = await response.json().catch((cause) => {
    throw new GooseProtocolError('GooseForum API returned invalid JSON', {
      status: response.status,
      cause,
    })
  }) as ApiEnvelope<T> | null

  if (!envelope || typeof envelope !== 'object' || Array.isArray(envelope)) {
    throw new GooseProtocolError('GooseForum API returned an invalid response envelope', {
      status: response.status,
    })
  }

  if (envelope.code !== undefined && envelope.code !== 0) {
    throw new GooseClientError(envelope.message || envelope.messageCode || 'GooseForum API request failed', {
      status: response.status,
      code: envelope.code,
      messageCode: envelope.messageCode,
      params: envelope.params,
    })
  }
  if (!response.ok) {
    throw new GooseClientError(`GooseForum API request failed with HTTP ${response.status}`, {
      status: response.status,
      messageCode: envelope.messageCode,
      params: envelope.params,
    })
  }
  return {
    value: (envelope.result ?? envelope.data) as T,
    message: envelope.message,
    messageCode: envelope.messageCode,
    params: envelope.params,
  }
}

export function createHttpClient(options: GooseClientOptions = {}): GooseHttpClient {
  const fetchImpl = options.fetch || defaultFetch
  const baseURL = options.baseURL?.replace(/\/$/, '')

  function resolve(path: string | URL) {
    const value = path.toString()
    if (/^https?:\/\//i.test(value) || !baseURL) return value
    return new URL(value, `${baseURL}/`).toString()
  }

  async function requestWithMeta<T>(path: string | URL, init: GooseRequestInit = {}) {
    const { json, query, headers: inputHeaders, ...requestInit } = init
    const headers = new Headers(inputHeaders)
    if (!headers.has('Accept')) headers.set('Accept', 'application/json')
    let body = requestInit.body
    if (json !== undefined) {
      headers.set('Accept', 'application/json')
      headers.set('Content-Type', 'application/json')
      body = JSON.stringify(json)
    }
    const response = await fetchImpl(appendQuery(resolve(path), query), {
      ...requestInit,
      headers,
      body,
    })
    return readEnvelope<T>(response)
  }

  async function request<T>(path: string | URL, init: GooseRequestInit = {}) {
    return (await requestWithMeta<T>(path, init)).value
  }

  return {
    payloadMajorVersion: options.payloadMajorVersion ?? 1,
    fetch: (input, init) => fetchImpl(resolve(input), init),
    request,
    requestWithMeta,
    resolve,
  }
}
