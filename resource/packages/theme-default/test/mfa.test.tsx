import { afterEach, expect, it, vi } from 'vitest'
import { act, cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { loginWithPassword, type GooseSiteApi, type LayoutPayload } from '@gooseforum/client'
import { GooseI18nProvider } from '@gooseforum/runtime/i18n'
import { GooseRuntimeProvider, type GooseRuntime } from '@gooseforum/runtime'
import { MFASettings } from '../src/site/settings/settings-mfa'
import { ConnectionsSettings } from '../src/site/settings/settings-connections'
import { LoginPageView } from '../src/site/auth/login-page'

vi.mock('@gooseforum/client', async (original) => ({ ...await original<typeof import('@gooseforum/client')>(), loginWithPassword: vi.fn() }))
afterEach(() => { cleanup(); vi.clearAllMocks() })

function runtimeFor(api: unknown, currentUrl = '/login') {
  return { api: api as GooseSiteApi, locale: 'en', currentUrl, navigate: vi.fn(), redirect: vi.fn(), queueFlash: vi.fn() } as unknown as GooseRuntime
}
function wrap(runtime: GooseRuntime, children: React.ReactNode) {
  return <GooseI18nProvider locale="en"><GooseRuntimeProvider runtime={runtime}>{children}</GooseRuntimeProvider></GooseI18nProvider>
}
const layout = { site: { name: 'GooseForum', logo: '', brandType: 'default' } } as LayoutPayload

it.each([0, 1, 10])('renders the remaining recovery-code count (%i)', async (remainingCodes) => {
  const runtime = runtimeFor({ users: { mfaStatus: vi.fn().mockResolvedValue({ enabled: true, available: true, remainingCodes }) } })
  render(wrap(runtime, <MFASettings showError={vi.fn()} />))
  expect(await screen.findByText(`Enabled · ${remainingCodes} recovery codes remaining`)).toBeTruthy()
  expect(screen.queryByText(/\{count\}/)).toBeNull()
})

it('does not offer setup when the server encryption key is missing', async () => {
  const mfaBegin = vi.fn()
  const runtime = runtimeFor({ users: { mfaStatus: vi.fn().mockResolvedValue({ enabled: false, available: false, remainingCodes: 0 }), mfaBegin } })
  render(wrap(runtime, <MFASettings showError={vi.fn()} onEnabled={vi.fn()} />))
  expect(await screen.findByText('Two-factor authentication is unavailable. Contact the administrator.')).toBeTruthy()
  expect(screen.queryByRole('button', { name: 'Set up authenticator' })).toBeNull()
  expect(mfaBegin).not.toHaveBeenCalled()
})

it('shows a retryable error when loading MFA settings fails', async () => {
  const mfaStatus = vi.fn().mockRejectedValueOnce(new Error('offline')).mockResolvedValueOnce({ enabled: false, available: true, remainingCodes: 0 })
  const runtime = runtimeFor({ users: { mfaStatus } })
  const enabled = vi.fn()
  const user = userEvent.setup()
  render(wrap(runtime, <MFASettings showError={vi.fn()} onEnabled={enabled} />))
  expect(await screen.findByRole('alert')).toBeTruthy()
  expect(screen.queryByLabelText('Account password')).toBeNull()
  await user.click(screen.getByRole('button', { name: 'Retry' }))
  expect(await screen.findByLabelText('Account password')).toBeTruthy()
  expect(mfaStatus).toHaveBeenCalledTimes(2)
  expect(enabled).toHaveBeenCalledWith(false)
})

it('confirms setup before showing recovery codes and leaves them available to save', async () => {
  const mfaBegin = vi.fn().mockResolvedValue({ secret: 'ABCDEF', qr: 'data:image/png;base64,AA==', uri: 'otpauth://totp/example' })
  const mfaChange = vi.fn().mockResolvedValue({ recoveryCodes: ['AAAA-BBBB-CCCC-DDDD', 'EEEE-FFFF-GGGG-HHHH'] })
  const runtime = runtimeFor({ users: { mfaStatus: vi.fn().mockResolvedValue({ enabled: false, available: true, remainingCodes: 0 }), mfaBegin, mfaChange } })
  const user = userEvent.setup()
  render(wrap(runtime, <MFASettings showError={vi.fn()} onEnabled={vi.fn()} />))
  await user.type(await screen.findByLabelText('Account password'), 'password123')
  await user.click(screen.getByRole('button', { name: 'Set up authenticator' }))
  expect(await screen.findByRole('img', { name: 'Authenticator setup QR code' })).toBeTruthy()
  expect(screen.queryByLabelText('Account password')).toBeNull()
  expect(mfaChange).not.toHaveBeenCalled()
  await user.type(screen.getByLabelText('Authenticator or recovery code'), '123456')
  await user.click(screen.getByRole('button', { name: 'Enable two-factor authentication' }))
  await waitFor(() => expect(mfaChange).toHaveBeenCalledWith('password123', '123456', 'enable'))
  const recovery = await screen.findByText(/AAAA-BBBB-CCCC-DDDD/)
  expect(recovery.textContent).toBe('AAAA-BBBB-CCCC-DDDD\nEEEE-FFFF-GGGG-HHHH')
  expect(screen.getByRole('button', { name: 'Download recovery codes' })).toBeTruthy()
  expect(runtime.redirect).not.toHaveBeenCalled()
})

it('clears setup credentials when authenticator binding is cancelled', async () => {
  const runtime = runtimeFor({ users: {
    mfaStatus: vi.fn().mockResolvedValue({ enabled: false, available: true, remainingCodes: 0 }),
    mfaBegin: vi.fn().mockResolvedValue({ secret: 'ABCDEF', uri: 'otpauth://totp/example' }),
  } })
  const user = userEvent.setup()
  render(wrap(runtime, <MFASettings showError={vi.fn()} />))
  await user.type(await screen.findByLabelText('Account password'), 'password123')
  await user.click(screen.getByRole('button', { name: 'Set up authenticator' }))
  await user.type(await screen.findByLabelText('Authenticator or recovery code'), '123456')
  await user.click(screen.getByRole('button', { name: 'Cancel setup' }))
  expect(screen.getByLabelText('Account password')).toHaveProperty('value', '')
  expect(screen.queryByRole('img', { name: 'Authenticator setup QR code' })).toBeNull()
  expect(screen.queryByLabelText('Authenticator or recovery code')).toBeNull()
})

it('holds password login at the second factor until verification succeeds', async () => {
  vi.mocked(loginWithPassword).mockResolvedValue({ mfaRequired: true })
  const mfaLogin = vi.fn().mockResolvedValue({ redirect: '/' })
  const runtime = runtimeFor({ auth: { captcha: vi.fn().mockResolvedValue({ captchaId: 'id', captchaImg: 'data:image/png;base64,AA==' }), mfaLogin } })
  const user = userEvent.setup()
  render(wrap(runtime, <LoginPageView layout={layout} page={{ initialMode: 'login', redirectUrl: '/topic/1', oauthProviders: [] }} />))
  await user.type(screen.getByLabelText('Username or email'), 'alice')
  await user.type(screen.getByLabelText('Password'), 'password123')
  await user.type(screen.getByLabelText('Captcha'), 'captcha')
  await user.click(screen.getByRole('button', { name: 'Log in' }))
  expect(await screen.findByLabelText('Authenticator or recovery code')).toBeTruthy()
  expect(runtime.navigate).not.toHaveBeenCalled()
  await user.type(screen.getByLabelText('Authenticator or recovery code'), '123456')
  await user.click(screen.getByRole('button', { name: 'Verify' }))
  await waitFor(() => expect(mfaLogin).toHaveBeenCalledWith('123456'))
  expect(runtime.navigate).toHaveBeenCalledWith('/topic/1', { replace: true })
})

it('continues the OAuth challenge to its server-validated redirect', async () => {
  const mfaLogin = vi.fn().mockResolvedValue({ redirect: '/settings?tab=account' })
  const runtime = runtimeFor({ auth: { captcha: vi.fn().mockResolvedValue({ captchaId: 'id', captchaImg: '' }), mfaLogin } }, '/login?mfa=1')
  const user = userEvent.setup()
  render(wrap(runtime, <LoginPageView layout={layout} page={{ initialMode: 'login', redirectUrl: '/', oauthProviders: [] }} />))
  await user.type(screen.getByLabelText('Authenticator or recovery code'), 'RECOVERY-CODE')
  await user.click(screen.getByRole('button', { name: 'Verify' }))
  await waitFor(() => expect(runtime.navigate).toHaveBeenCalledWith('/settings?tab=account', { replace: true }))
})

it('uses the password login destination after leaving an OAuth challenge', async () => {
  vi.mocked(loginWithPassword).mockResolvedValue({ mfaRequired: true })
  const mfaLogin = vi.fn().mockResolvedValue({ redirect: '/' })
  const runtime = runtimeFor({ auth: { captcha: vi.fn().mockResolvedValue({ captchaId: 'id', captchaImg: 'data:image/png;base64,AA==' }), mfaLogin } }, '/login?mfa=1')
  const user = userEvent.setup()
  render(wrap(runtime, <LoginPageView layout={layout} page={{ initialMode: 'login', redirectUrl: '/topic/1', oauthProviders: [] }} />))
  await user.click(screen.getByRole('button', { name: 'Back to login' }))
  await user.type(screen.getByLabelText('Username or email'), 'alice')
  await user.type(screen.getByLabelText('Password'), 'password123')
  await user.type(await screen.findByLabelText('Captcha'), 'captcha')
  await user.click(screen.getByRole('button', { name: 'Log in' }))
  await user.type(await screen.findByLabelText('Authenticator or recovery code'), '123456')
  await user.click(screen.getByRole('button', { name: 'Verify' }))
  await waitFor(() => expect(runtime.navigate).toHaveBeenCalledWith('/topic/1', { replace: true }))
})

it('verifies the second factor through POST before starting an OAuth binding', async () => {
  const prepareOAuthBind = vi.fn().mockResolvedValue({ redirect: 'https://provider.example/authorize?state=signed' })
  const runtime = runtimeFor({ users: {
    mfaStatus: vi.fn().mockResolvedValue({ enabled: true, available: true, remainingCodes: 10 }),
    oauthBindings: vi.fn().mockResolvedValue([{ key: 'github', displayName: 'GitHub', enabled: true, bound: false }]),
    oidcGrants: vi.fn().mockResolvedValue([]), prepareOAuthBind,
  } }, '/settings?tab=binding')
  const user = userEvent.setup()
  render(wrap(runtime, <ConnectionsSettings section="binding" showError={vi.fn()} showStatus={vi.fn()} />))
  await user.type(await screen.findByLabelText('Authenticator or recovery code'), 'RECOVERY-CODE')
  await user.click(await screen.findByRole('button', { name: 'Connect' }))
  await waitFor(() => expect(prepareOAuthBind).toHaveBeenCalledWith('github', 'RECOVERY-CODE'))
  expect(runtime.redirect).toHaveBeenCalledWith('https://provider.example/authorize?state=signed')
})

it('locks all providers and refresh while an OAuth security operation is in progress', async () => {
  let finish!: (value: { redirect: string }) => void
  const prepareOAuthBind = vi.fn().mockImplementation(() => new Promise(resolve => { finish = resolve }))
  const runtime = runtimeFor({ users: {
    mfaStatus: vi.fn().mockResolvedValue({ enabled: true, available: true, remainingCodes: 10 }),
    oauthBindings: vi.fn().mockResolvedValue([{ key: 'github', displayName: 'GitHub', enabled: true, bound: false }, { key: 'google', displayName: 'Google', enabled: true, bound: true }]),
    oidcGrants: vi.fn().mockResolvedValue([]), prepareOAuthBind, unbindOAuth: vi.fn(),
  } }, '/settings?tab=binding')
  const user = userEvent.setup()
  render(wrap(runtime, <ConnectionsSettings section="binding" showError={vi.fn()} showStatus={vi.fn()} />))
  await user.type(await screen.findByLabelText('Authenticator or recovery code'), 'RECOVERY-CODE')
  await user.click(screen.getByRole('button', { name: 'Connect' }))
  await waitFor(() => expect(prepareOAuthBind).toHaveBeenCalledOnce())
  expect(screen.getByRole('button', { name: 'Disconnect' }).hasAttribute('disabled')).toBe(true)
  expect(screen.getByRole('button', { name: 'Refresh' }).hasAttribute('disabled')).toBe(true)
  await user.click(screen.getByRole('button', { name: 'Disconnect' }))
  expect(runtime.api.users.unbindOAuth).not.toHaveBeenCalled()
  await act(() => { finish({ redirect: 'https://provider.example/authorize' }); return Promise.resolve() })
  expect(runtime.redirect).toHaveBeenCalledWith('https://provider.example/authorize')
})

it('blocks OAuth changes until MFA status can be loaded and retried', async () => {
  const prepareOAuthBind = vi.fn()
  const mfaStatus = vi.fn().mockRejectedValueOnce(new Error('offline')).mockResolvedValueOnce({ enabled: true, available: true, remainingCodes: 10 })
  const runtime = runtimeFor({ users: {
    mfaStatus, prepareOAuthBind,
    oauthBindings: vi.fn().mockResolvedValue([{ key: 'github', displayName: 'GitHub', enabled: true, bound: false }]),
    oidcGrants: vi.fn().mockResolvedValue([]),
  } })
  const user = userEvent.setup()
  render(wrap(runtime, <ConnectionsSettings section="binding" showError={vi.fn()} showStatus={vi.fn()} />))
  expect(await screen.findByRole('alert')).toBeTruthy()
  const connect = screen.getByRole('button', { name: 'Connect' })
  expect(connect.hasAttribute('disabled')).toBe(true)
  await user.click(connect)
  expect(prepareOAuthBind).not.toHaveBeenCalled()
  await user.click(screen.getByRole('button', { name: 'Retry' }))
  await user.type(await screen.findByLabelText('Authenticator or recovery code'), '123456')
  expect(screen.getByRole('button', { name: 'Connect' }).hasAttribute('disabled')).toBe(false)
  expect(mfaStatus).toHaveBeenCalledTimes(2)
})

it('does not report errors from bindings after leaving the settings page', async () => {
  let fail!: (reason: Error) => void
  const showError = vi.fn()
  const runtime = runtimeFor({ users: {
    mfaStatus: vi.fn().mockResolvedValue({ enabled: false, available: true, remainingCodes: 0 }),
    oauthBindings: vi.fn().mockImplementation(() => new Promise((_resolve, reject) => { fail = reject })),
    oidcGrants: vi.fn().mockResolvedValue([]),
  } })
  const mounted = render(wrap(runtime, <ConnectionsSettings section="binding" showError={showError} showStatus={vi.fn()} />))
  await waitFor(() => expect(runtime.api.users.oauthBindings).toHaveBeenCalledOnce())
  mounted.unmount()
  await act(() => { fail(new Error('offline')); return Promise.resolve() })
  expect(showError).not.toHaveBeenCalled()
})
