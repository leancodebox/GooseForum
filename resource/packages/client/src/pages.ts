import { pageComponents, type AnyPagePayload } from './contracts/page-components.js'
import type { PagePayload } from './contracts/payload.js'
import type { GooseHttpClient } from './http/client.js'
import { GooseClientError, GooseProtocolError } from './http/error.js'

export interface PageClient<TPage extends PagePayload = AnyPagePayload> {
  fetch(path: string | URL, init?: RequestInit): Promise<TPage>
  parse(value: unknown): TPage
}

export interface PageClientOptions {
  /** Additional component names supplied by extensions or a newer server. */
  components?: readonly string[]
  /** Accept unknown components as generic PagePayload values. Defaults to false. */
  allowUnknownComponents?: boolean
  /** Optional application-specific payload validation executed after envelope validation. */
  validate?: (payload: PagePayload) => void
}

function payloadMajor(version: string) {
  const major = Number.parseInt(version.split('.')[0] || '', 10)
  return Number.isFinite(major) ? major : undefined
}

export function createPageClient<TPage extends PagePayload = AnyPagePayload>(
  http: GooseHttpClient,
  options: PageClientOptions = {},
): PageClient<TPage> {
  const supportedComponents = new Set<string>([...pageComponents, ...(options.components || [])])

  function parse(value: unknown) {
    if (!value || typeof value !== 'object') {
      throw new GooseProtocolError('GooseForum page payload must be an object')
    }
    const payload = value as Partial<PagePayload>
    if (!payload.component || typeof payload.component !== 'string') {
      throw new GooseProtocolError('GooseForum page payload is missing component')
    }
    if (!supportedComponents.has(payload.component) && !options.allowUnknownComponents) {
      throw new GooseProtocolError(`Unsupported GooseForum page component ${payload.component}`)
    }
    if (!payload.version || typeof payload.version !== 'string') {
      throw new GooseProtocolError('GooseForum page payload is missing version')
    }
    const major = payloadMajor(payload.version)
    if (major !== http.payloadMajorVersion) {
      throw new GooseProtocolError(
        `Unsupported GooseForum payload version ${payload.version}; expected ${http.payloadMajorVersion}.x`,
      )
    }
    if (!payload.props || typeof payload.props !== 'object') {
      throw new GooseProtocolError('GooseForum page payload is missing props')
    }
    if (!payload.meta || typeof payload.meta !== 'object' || typeof payload.meta.title !== 'string') {
      throw new GooseProtocolError('GooseForum page payload is missing meta')
    }
    if (!payload.layout || typeof payload.layout !== 'object') {
      throw new GooseProtocolError('GooseForum page payload is missing layout')
    }
    if (typeof payload.url !== 'string') {
      throw new GooseProtocolError('GooseForum page payload is missing url')
    }
    options.validate?.(payload as PagePayload)
    return payload as TPage
  }

  async function fetchPage(path: string | URL, init: RequestInit = {}) {
    const headers = new Headers(init.headers)
    headers.set('Accept', 'application/json')
    headers.set('X-Goose-Page', 'true')
    const response = await http.fetch(path, { ...init, headers })
    if (!response.ok && response.status !== 404) {
      throw new GooseClientError(`GooseForum page request failed with HTTP ${response.status}`, {
        status: response.status,
      })
    }
    const value = await response.json().catch((cause) => {
      throw new GooseProtocolError('GooseForum page returned invalid JSON', {
        status: response.status,
        cause,
      })
    })
    return parse(value)
  }

  return { fetch: fetchPage, parse }
}
