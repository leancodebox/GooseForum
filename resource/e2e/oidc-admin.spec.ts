import { expect, test } from '@playwright/test'

test('manages OIDC connections, grants, deletion and key recovery', async ({ page }, testInfo) => {
  await page.route('**/__goose_page/**', route => route.fulfill({ json: {
    component: 'admin.shell', props: {}, url: '/admin/settings/oidc-provider?lang=en', version: '1.0', meta: { title: 'OIDC' },
    layout: { site: { name: 'GooseForum', description: '', logo: '', favicon: '', brandType: 'default', brandText: '', brandImage: '' }, viewer: { id: 1, username: 'admin', email: 'admin@example.com', avatarUrl: '', isAuthenticated: true, canAccessAdmin: true, isModerator: false, requiresEmailVerification: false, adminPermissions: [0] }, header: [], sidebar: { activeKey: 'topics', categories: [] }, footer: { links: [], primary: [] }, unread: { notifications: false, messages: false }, theme: { enabled: false, current: 'gf-light', themeColor: '#fbfdff' } },
  } }))
  let available = true
  let deleted = false
  const revoked: unknown[] = []
  let reset = false
  const client = { clientId: 'gf_example', name: 'Example application', public: false, enabled: true, scopes: ['openid', 'profile', 'email', 'offline_access'], grantTypes: ['authorization_code', 'refresh_token'], redirectUris: ['https://example.com/auth/callback'], tokenEndpointAuthMethod: 'client_secret_basic', requirePkce: true }
  await page.route('**/api/admin/oidc-provider', route => route.fulfill({ json: { code: 0, result: { enabled: true, available, ...(available ? { issuer: 'https://forum.example.com/oauth2' } : { error: 'Unable to decrypt stored signing key' }) } } }))
  await page.route('**/api/admin/oidc-clients', route => route.fulfill({ json: { code: 0, result: deleted ? [] : [client] } }))
  await page.route('**/api/admin/oidc-clients/grants**', route => {
    const after = new URL(route.request().url()).searchParams.get('after')
    return route.fulfill({ json: { code: 0, result: { items: [{ userId: after ? '21' : '1', username: after ? 'bob' : 'alice', scopes: ['openid', 'profile'], createdAt: '2026-10-04T00:00:00Z', updatedAt: '2026-10-04T01:00:00Z' }], hasMore: !after } } })
  })
  await page.route('**/api/admin/oidc-clients/revoke-grant', route => { revoked.push(route.request().postDataJSON()); return route.fulfill({ json: { code: 0, result: true } }) })
  await page.route('**/api/admin/oidc-clients/delete', route => { deleted = true; return route.fulfill({ json: { code: 0, result: true } }) })
  await page.route('**/api/admin/oidc-provider/reset-signing-key', route => { reset = true; available = true; return route.fulfill({ json: { code: 0, result: { enabled: true, available: true, issuer: 'https://forum.example.com/oauth2' } } }) })
  await page.goto('/admin/settings/oidc-provider?lang=en')
  const endpoints = page.getByRole('button', { name: 'Connection endpoints', exact: true })
  await expect(endpoints).toHaveAttribute('aria-expanded', 'false')
  await expect(page.getByText('https://forum.example.com/oauth2/.well-known/openid-configuration', { exact: true })).toBeHidden()
  await endpoints.click()
  await expect(endpoints).toHaveAttribute('aria-expanded', 'true')
  await expect(page.getByText('https://forum.example.com/oauth2/.well-known/openid-configuration', { exact: true })).toBeVisible()
  expect(await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 1)).toBe(false)
  await testInfo.attach('oidc-provider', { body: await page.screenshot({ path: `/tmp/gooseforum-oidc-${testInfo.project.name}.png`, fullPage: true }), contentType: 'image/png' })
  await page.getByRole('button', { name: 'User grants', exact: true }).click()
  const grants = page.getByRole('dialog', { name: 'Example application · User grants' })
  await expect(grants.getByText('alice', { exact: true })).toBeVisible()
  await grants.getByRole('button', { name: 'Next page' }).click()
  await expect(grants.getByText('bob', { exact: true })).toBeVisible()
  await grants.getByRole('button', { name: 'Previous page' }).click()
  await expect(grants.getByText('alice', { exact: true })).toBeVisible()
  await testInfo.attach('oidc-grants', { body: await page.screenshot({ path: `/tmp/gooseforum-oidc-grants-${testInfo.project.name}.png` }), contentType: 'image/png' })
  expect(await grants.evaluate(el => el.scrollWidth > el.clientWidth + 1)).toBe(false)
  await grants.getByRole('button', { name: 'Revoke grant', exact: true }).click()
  await page.getByRole('dialog', { name: 'Revoke grant', exact: true }).getByRole('button', { name: 'Confirm' }).click()
  await expect(page.getByRole('dialog', { name: 'Revoke grant', exact: true })).toBeHidden()
  expect(revoked).toEqual([{ clientId: 'gf_example', userId: '1' }])
  await grants.getByRole('button', { name: 'Revoke all grants', exact: true }).click()
  await page.getByRole('dialog', { name: 'Revoke all grants', exact: true }).getByRole('button', { name: 'Confirm' }).click()
  await expect(page.getByRole('dialog', { name: 'Revoke all grants', exact: true })).toBeHidden()
  expect(revoked[1]).toEqual({ clientId: 'gf_example' })
  await grants.getByRole('button', { name: 'Close', exact: true }).click()
  await page.getByRole('button', { name: 'Delete client', exact: true }).click()
  await page.getByRole('dialog', { name: 'Delete client', exact: true }).getByRole('button', { name: 'Confirm' }).click()
  await expect(page.getByText('No OIDC clients', { exact: true })).toBeVisible()
  available = false
  await page.getByRole('button').filter({ has: page.locator('svg.lucide-refresh-cw') }).click()
  await page.getByRole('button', { name: 'Reset signing key', exact: true }).click()
  await page.getByRole('dialog', { name: 'Reset signing key', exact: true }).getByRole('button', { name: 'Confirm' }).click()
  await expect(page.getByText('Running', { exact: true })).toBeVisible()
  expect(reset).toBe(true)
})
