import { afterEach, expect, it, vi } from 'vitest'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { GooseAdminApi } from '@gooseforum/client'
import messages from '../messages/en-users'
import { UserMFAReset } from './user-mfa-reset'
afterEach(cleanup)

it('requires a reason and explicit confirmation before resetting MFA', async () => {
  const resetMFA = vi.fn().mockResolvedValue(undefined)
  const api = { users: { mfaStatus: vi.fn().mockResolvedValue({ enabled: true }), resetMFA } } as unknown as GooseAdminApi
  render(<UserMFAReset userId={7} username="alice" api={api} text={key => messages[key]} disabled={false} />)
  const user = userEvent.setup()
  await waitFor(() => expect((screen.getByRole('button', { name: messages.mfaReset }) as HTMLButtonElement).disabled).toBe(false))
  await user.click(screen.getByRole('button', { name: messages.mfaReset }))
  const buttons = screen.getAllByRole('button', { name: messages.mfaReset })
  expect((buttons.at(-1) as HTMLButtonElement).disabled).toBe(true)
  expect(resetMFA).not.toHaveBeenCalled()
  await user.type(screen.getByLabelText(messages.mfaResetReason), 'Identity verified')
  await user.click(buttons.at(-1)!)
  await waitFor(() => expect(resetMFA).toHaveBeenCalledWith(7, 'Identity verified'))
  expect(await screen.findByText(messages.mfaNotEnabled)).toBeTruthy()
})
