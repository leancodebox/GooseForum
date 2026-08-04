import { pageComponents, type AnyPagePayload } from './contracts/page-components.js'
import type { PagePayload } from './contracts/payload.js'
import type { GooseHttpClient } from './http/client.js'
import { GooseClientError, GooseProtocolError } from './http/error.js'

export interface PageClient {
  fetch(path: string | URL, init?: RequestInit): Promise<AnyPagePayload>
  parse(value: unknown): AnyPagePayload
}

function payloadMajor(version: string) {
  const major = Number.parseInt(version.split('.')[0] || '', 10)
  return Number.isFinite(major) ? major : undefined
}

export function createPageClient(http: GooseHttpClient): PageClient {
  const supportedComponents = new Set<string>(pageComponents)

  function parse(value: unknown) {
    if (!value || typeof value !== 'object') {
      throw new GooseProtocolError('GooseForum page payload must be an object')
    }
    const payload = value as Partial<PagePayload>
    if (!payload.component || typeof payload.component !== 'string') {
      throw new GooseProtocolError('GooseForum page payload is missing component')
    }
    if (!supportedComponents.has(payload.component)) {
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
    return payload as AnyPagePayload
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
