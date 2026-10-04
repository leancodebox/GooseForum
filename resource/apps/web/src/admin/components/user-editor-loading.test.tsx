import { afterEach, expect, it, vi } from 'vitest'
import { act, cleanup, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { AdminUser, GooseAdminApi } from '@gooseforum/client'
import messages from '../messages/en-users'
import { UsersManagementPage } from '../pages/users-management-page'

afterEach(cleanup)
vi.stubGlobal('ResizeObserver', class { observe() {} unobserve() {} disconnect() {} })

const member: AdminUser = { userId: 7, username: 'alice', email: 'alice@example.com', restrictionStatus: 'normal', validate: 1, prestige: 0, roleId: 3, roleList: [{ name: 'Moderator', value: 3 }], createTime: '2026-10-03' }

function mount(roles: ReturnType<typeof vi.fn>, badgeOptions: ReturnType<typeof vi.fn>, users = [member]) {
  const edit = vi.fn().mockResolvedValue(undefined)
  const saveBadges = vi.fn().mockResolvedValue(undefined)
  const api = { users: { list: vi.fn().mockResolvedValue({ list: users, total: users.length }), roles, badgeOptions, edit, saveBadges, restrictionHistory: vi.fn().mockResolvedValue({ list: [], total: 0 }) } } as unknown as GooseAdminApi
  render(<UsersManagementPage api={api} text={key => messages[key]} />)
  return { edit, saveBadges }
}

it('shows load errors, preserves the existing role, and never clears badges on a failed load', async () => {
  const { edit, saveBadges } = mount(vi.fn().mockRejectedValue(new Error('Roles offline')), vi.fn().mockRejectedValue(new Error('Badges offline')))
  const actor = userEvent.setup()
  await screen.findAllByText('alice')
  await actor.click(screen.getAllByRole('button', { name: 'Edit user' })[0])
  expect(await screen.findByText('Roles offline')).toBeTruthy()
  const role = screen.getByLabelText<HTMLButtonElement>('Role')
  expect(role.disabled).toBe(true)
  expect(role.textContent).toContain('Moderator')
  await actor.click(screen.getByRole('tab', { name: 'Badges' }))
  expect(await screen.findByText('Badges offline')).toBeTruthy()
  expect(screen.queryByText('No manual badges available')).toBeNull()
  await actor.click(screen.getByRole('button', { name: 'Save changes' }))
  await waitFor(() => expect(edit).toHaveBeenCalledWith(expect.objectContaining({ userId: 7, roleId: 3 })))
  expect(saveBadges).not.toHaveBeenCalled()
})

it('does not write unchanged badges when saving account settings', async () => {
  const { edit, saveBadges } = mount(vi.fn().mockResolvedValue([{ name: 'Moderator', value: 3 }]), vi.fn().mockResolvedValue({ options: [{ code: 'helper', name: 'Helpful' }], active: [{ code: 'helper', name: 'Helpful', source: 'manual' }] }))
  const actor = userEvent.setup()
  await screen.findAllByText('alice')
  await actor.click(screen.getAllByRole('button', { name: 'Edit user' })[0])
  await actor.click(screen.getByRole('tab', { name: 'Badges' }))
  expect((await screen.findByRole('button', { name: 'Helpful' })).getAttribute('aria-pressed')).toBe('true')
  await actor.click(screen.getByRole('button', { name: 'Save changes' }))
  await waitFor(() => expect(edit).toHaveBeenCalledOnce())
  await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull())
  expect(saveBadges).not.toHaveBeenCalled()
})

it('retries each failed query independently and restores role and badge controls', async () => {
  const roles = vi.fn().mockRejectedValueOnce(new Error('Roles offline')).mockResolvedValue([{ name: 'Moderator', value: 3 }])
  const badgeOptions = vi.fn().mockRejectedValueOnce(new Error('Badges offline')).mockResolvedValue({ options: [{ code: 'helper', name: 'Helpful' }, ...Array.from({ length: 7 }, (_, index) => ({ code: `badge-${index}`, name: `Badge ${index}` }))], active: [{ code: 'helper', name: 'Helpful', source: 'manual' }] })
  const { edit, saveBadges } = mount(roles, badgeOptions)
  const actor = userEvent.setup()
  await screen.findAllByText('alice')
  await actor.click(screen.getAllByRole('button', { name: 'Edit user' })[0])
  const roleError = await screen.findByText('Roles offline')
  await actor.click(within(roleError.closest('[role="alert"]') as HTMLElement).getByRole('button', { name: 'Retry' }))
  await waitFor(() => expect(screen.getByLabelText<HTMLButtonElement>('Role').disabled).toBe(false))
  expect(roles).toHaveBeenCalledTimes(2)
  expect(badgeOptions).toHaveBeenCalledTimes(1)
  await actor.click(screen.getByRole('tab', { name: 'Badges' }))
  const badgeError = await screen.findByText('Badges offline')
  await actor.click(within(badgeError.closest('[role="alert"]') as HTMLElement).getByRole('button', { name: 'Retry' }))
  expect((await screen.findByRole('button', { name: 'Helpful' })).getAttribute('data-current')).toBe('true')
  expect(badgeOptions).toHaveBeenCalledTimes(2)
  expect(screen.getAllByRole('button', { name: /^Badge \d$/ })).toHaveLength(7)
  await actor.click(screen.getByRole('button', { name: 'Badge 0' }))
  expect(edit).not.toHaveBeenCalled()
  await actor.click(screen.getByRole('button', { name: 'Save changes' }))
  await waitFor(() => expect(saveBadges).toHaveBeenCalledWith(7, ['helper', 'badge-0']))
})

it('ignores late retry results after opening another user', async () => {
  let finishRoles!: (value: unknown) => void
  let finishBadges!: (value: unknown) => void
  const roles = vi.fn().mockRejectedValueOnce(new Error('Roles offline')).mockImplementationOnce(() => new Promise(resolve => { finishRoles = resolve })).mockResolvedValue([{ name: 'Member', value: 4 }])
  const badgeOptions = vi.fn().mockRejectedValueOnce(new Error('Badges offline')).mockImplementationOnce(() => new Promise(resolve => { finishBadges = resolve })).mockResolvedValue({ options: [{ code: 'new', name: 'New badge' }], active: [] })
  mount(roles, badgeOptions, [member, { ...member, userId: 8, username: 'bob', roleId: 4, roleList: [{ name: 'Member', value: 4 }] }])
  const actor = userEvent.setup()
  await screen.findAllByText('alice')
  await actor.click(screen.getAllByRole('button', { name: 'Edit user' })[0])
  const roleError = await screen.findByText('Roles offline')
  await actor.click(within(roleError.closest('[role="alert"]') as HTMLElement).getByRole('button', { name: 'Retry' }))
  await actor.click(screen.getByRole('tab', { name: 'Badges' }))
  const badgeError = await screen.findByText('Badges offline')
  await actor.click(within(badgeError.closest('[role="alert"]') as HTMLElement).getByRole('button', { name: 'Retry' }))
  await actor.click(screen.getByRole('button', { name: 'Cancel' }))
  await actor.click(screen.getAllByRole('button', { name: 'Edit user' })[1])
  await actor.click(screen.getByRole('tab', { name: 'Badges' }))
  await screen.findByRole('button', { name: 'New badge' })
  await act(async () => {
    finishRoles([{ name: 'Old role', value: 3 }])
    finishBadges({ options: [{ code: 'old', name: 'Old badge' }], active: [] })
    await Promise.resolve()
  })
  expect(screen.queryByRole('button', { name: 'Old badge' })).toBeNull()
  expect(screen.getByRole('button', { name: 'New badge' })).toBeTruthy()
  await actor.click(screen.getByRole('tab', { name: 'Account' }))
  expect(screen.getByLabelText('Role').textContent).toContain('Member')
})

it('preserves account edits, the deadline, and selected badges while switching tabs', async () => {
  const until = new Date(2030, 0, 2, 10, 30).toISOString()
  const { edit, saveBadges } = mount(vi.fn().mockResolvedValue([{ name: 'Moderator', value: 3 }]), vi.fn().mockResolvedValue({ options: [{ code: 'helper', name: 'Helpful' }], active: [] }), [{ ...member, restrictionStatus: 'banned', restrictionReason: 'Initial reason', restrictionUntil: until }])
  const actor = userEvent.setup()
  await screen.findAllByText('alice')
  await actor.click(screen.getAllByRole('button', { name: 'Edit user' })[0])
  await actor.clear(screen.getByLabelText('Restriction reason'))
  await actor.type(screen.getByLabelText('Restriction reason'), 'Reviewed reason')
  await actor.click(screen.getByRole('tab', { name: 'Badges' }))
  await actor.click(await screen.findByRole('button', { name: 'Helpful' }))
  await actor.click(screen.getByRole('tab', { name: 'Account' }))
  expect(screen.getByLabelText<HTMLTextAreaElement>('Restriction reason').value).toBe('Reviewed reason')
  expect(screen.getByLabelText('Restricted until').textContent).toContain('2030')
  await actor.click(screen.getByRole('tab', { name: 'Badges' }))
  expect(screen.getByRole('button', { name: 'Helpful' }).getAttribute('aria-pressed')).toBe('true')
  await actor.click(screen.getByRole('tab', { name: 'History' }))
  expect(edit).not.toHaveBeenCalled()
  await actor.click(screen.getByRole('button', { name: 'Save changes' }))
  await waitFor(() => expect(edit).toHaveBeenCalledWith(expect.objectContaining({ restrictionReason: 'Reviewed reason', restrictionUntil: until, roleId: 3 })))
  expect(saveBadges).toHaveBeenCalledWith(7, ['helper'])
})

it('returns to Account and blocks all API writes when a hidden restriction reason is blank', async () => {
  const { edit, saveBadges } = mount(vi.fn().mockResolvedValue([]), vi.fn().mockResolvedValue({ options: [], active: [] }), [{ ...member, restrictionStatus: 'banned', restrictionReason: '   ' }])
  const actor = userEvent.setup()
  await screen.findAllByText('alice')
  await actor.click(screen.getAllByRole('button', { name: 'Edit user' })[0])
  await actor.click(screen.getByRole('tab', { name: 'Badges' }))
  await actor.click(screen.getByRole('button', { name: 'Save changes' }))
  expect(await screen.findByRole('alert')).toHaveProperty('textContent', 'Enter a reason for the restriction.')
  expect(screen.getByRole('tab', { name: 'Account' }).getAttribute('aria-selected')).toBe('true')
  expect(edit).not.toHaveBeenCalled()
  expect(saveBadges).not.toHaveBeenCalled()
})
