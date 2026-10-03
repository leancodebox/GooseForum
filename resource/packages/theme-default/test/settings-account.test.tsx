import { afterEach, expect, it, vi } from 'vitest'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { GooseSiteApi } from '@gooseforum/client'
import { GooseI18nProvider } from '@gooseforum/runtime/i18n'
import { GooseRuntimeProvider, type GooseRuntime } from '@gooseforum/runtime'
import { SessionSettings } from '../src/site/settings/settings-sessions'
import { AccountSettings } from '../src/site/settings/settings-account'

afterEach(cleanup)

it('lists sessions, revokes other devices, and redirects after password change', async () => {
  const authSessions = vi.fn().mockResolvedValue([
    { id: 1, authMethod: 'password', clientIp: '127.0.0.1', userAgent: 'Current browser', createdAt: '', lastSeenAt: '2026-09-28T00:00:00Z', current: true },
    { id: 2, authMethod: 'oauth', oauthProvider: 'github', clientIp: '127.0.0.2', userAgent: 'Other browser', createdAt: '', lastSeenAt: '2026-09-27T00:00:00Z', current: false },
  ])
  const revokeOtherAuthSessions = vi.fn().mockResolvedValue(true)
  const changePassword = vi.fn().mockResolvedValue(undefined)
  const redirect = vi.fn()
  const runtime = {
    api: { users: { authSessions, revokeOtherAuthSessions, revokeAuthSession: vi.fn(), changePassword } } as unknown as GooseSiteApi,
    locale: 'en', redirect, navigate: vi.fn(),
  } as unknown as GooseRuntime
  const user = userEvent.setup()

  const devices = render(<GooseI18nProvider locale="en"><GooseRuntimeProvider runtime={runtime}>
    <SessionSettings showStatus={vi.fn()} showError={vi.fn()} />
  </GooseRuntimeProvider></GooseI18nProvider>)

  expect(await screen.findByText('Other browser')).toBeTruthy()
  await user.click(screen.getByRole('button', { name: 'Sign out other devices' }))
  await waitFor(() => expect(revokeOtherAuthSessions).toHaveBeenCalledOnce())
  expect(screen.queryByText('Other browser')).toBeNull()

  devices.unmount()
  render(<GooseI18nProvider locale="en"><GooseRuntimeProvider runtime={runtime}>
    <AccountSettings showError={vi.fn()} />
  </GooseRuntimeProvider></GooseI18nProvider>)
  expect(authSessions).toHaveBeenCalledOnce()
  expect(screen.queryByText('Current browser')).toBeNull()

  await user.type(screen.getByLabelText('Current password'), 'old-password')
  await user.type(screen.getByLabelText('New password'), 'new-password')
  await user.type(screen.getByLabelText('Confirm password'), 'new-password')
  await user.click(screen.getByRole('button', { name: 'Change password' }))
  await waitFor(() => expect(changePassword).toHaveBeenCalledWith('old-password', 'new-password'))
  expect(redirect).toHaveBeenCalledWith('/login')
})
