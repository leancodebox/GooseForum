import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  createOIDCClient,
  getOIDCClients,
  getOIDCProviderStatus,
  rotateOIDCClientSecret,
  rotateOIDCSigningKey,
  saveOIDCProviderSettings,
  updateOIDCClient,
} from '@/admin/runtime/api'
import type { OIDCClient, OIDCClientInput } from '@/admin/types'

const client: OIDCClient = {
  clientId: 'gf_client',
  name: 'Internal Wiki',
  redirectUris: ['https://wiki.example.com/oauth/callback'],
  scopes: ['openid', 'profile'],
  grantTypes: ['authorization_code'],
  tokenEndpointAuthMethod: 'client_secret_basic',
  requirePkce: true,
  public: false,
  enabled: true,
}

const input: OIDCClientInput = {
  name: client.name,
  redirectUris: client.redirectUris,
  scopes: client.scopes,
  grantTypes: client.grantTypes,
  tokenEndpointAuthMethod: client.tokenEndpointAuthMethod,
  requirePkce: client.requirePkce,
  public: client.public,
  enabled: client.enabled,
}

function mockResult(result: unknown) {
  return vi.fn().mockImplementation(async () => new Response(JSON.stringify({ code: 0, result }), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  }))
}

afterEach(() => vi.unstubAllGlobals())

describe('OIDC client management API', () => {
  it('loads, toggles, and rotates the provider runtime', async () => {
    const status = { enabled: true, available: true, issuer: 'https://forum.example/oauth2' }
    const fetchMock = mockResult(status)
    vi.stubGlobal('fetch', fetchMock)

    await getOIDCProviderStatus()
    await saveOIDCProviderSettings(true)
    await rotateOIDCSigningKey()

    expect(fetchMock.mock.calls[0][0]).toBe('/api/admin/oidc-provider')
    expect(JSON.parse(fetchMock.mock.calls[1][1].body)).toEqual({ enabled: true })
    expect(fetchMock.mock.calls[2][0]).toBe('/api/admin/oidc-provider/rotate-signing-key')
  })

  it('lists registered clients', async () => {
    const fetchMock = mockResult([client])
    vi.stubGlobal('fetch', fetchMock)

    await expect(getOIDCClients()).resolves.toEqual([client])
    expect(fetchMock).toHaveBeenCalledWith('/api/admin/oidc-clients', {
      headers: { Accept: 'application/json' },
    })
  })

  it('creates a client with the controller request shape', async () => {
    const fetchMock = mockResult({ client, clientSecret: 'only-once' })
    vi.stubGlobal('fetch', fetchMock)

    await expect(createOIDCClient(input)).resolves.toMatchObject({ clientSecret: 'only-once' })
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual(input)
    expect(fetchMock.mock.calls[0][0]).toBe('/api/admin/oidc-clients/create')
  })

  it('updates a client and rotates its secret', async () => {
    const fetchMock = mockResult(client)
    vi.stubGlobal('fetch', fetchMock)

    await updateOIDCClient(client)
    await rotateOIDCClientSecret(client.clientId)

    expect(fetchMock.mock.calls[0][0]).toBe('/api/admin/oidc-clients/update')
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual(client)
    expect(fetchMock.mock.calls[1][0]).toBe('/api/admin/oidc-clients/rotate-secret')
    expect(JSON.parse(fetchMock.mock.calls[1][1].body)).toEqual({ clientId: client.clientId })
  })
})
