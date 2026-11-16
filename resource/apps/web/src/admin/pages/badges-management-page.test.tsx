import { afterEach, expect, it, vi } from 'vitest'
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { AdminBadge, GooseAdminApi } from '@gooseforum/client'
import messages from '../messages/zh-assets'
import { BadgesManagementPage } from './badges-management-page'

vi.stubGlobal('ResizeObserver', class { observe() {} unobserve() {} disconnect() {} })

const badge: AdminBadge = { code: 'helper', type: 'custom', grantMode: 'manual', name: '热心伙伴', description: '帮助社区成员', iconUrl: '/static/badges/contributor.svg', color: 'blue', level: 'bronze', isEnabled: true, isWearable: true, sortOrder: 1000, grantedCount: 2, canDelete: false }
const text = (key: keyof typeof messages) => messages[key]
function setup(items: AdminBadge[] = [badge]) {
  const api = {
    assets: {
      badges: vi.fn().mockResolvedValue(items), saveBadge: vi.fn().mockResolvedValue(undefined), deleteBadge: vi.fn().mockResolvedValue(undefined),
      badgeRecipients: vi.fn().mockResolvedValue({ list: [{ userId: 7, username: 'alice', avatarUrl: '', source: 'manual', reason: '提供帮助', grantedBy: 1, grantedByName: 'admin', grantedAt: '2026-11-17T02:00:00Z' }], total: 1, hasNext: false, page: 1, size: 20 }),
      updateBadgeRecipient: vi.fn().mockResolvedValue(undefined),
    },
    pages: { uploadImage: vi.fn().mockResolvedValue({ url: '/uploads/new-badge.png' }) },
    users: { list: vi.fn().mockResolvedValue({ list: [{ userId: 9, username: 'bob' }], total: 1, page: 1, size: 20 }) },
  }
  render(<BadgesManagementPage api={api as unknown as GooseAdminApi} text={text} canManageRecipients />)
  return { api, user: userEvent.setup() }
}
afterEach(() => { cleanup(); vi.restoreAllMocks() })

it('creates a manual badge from a selected icon without requiring a code and preserves sort zero', async () => {
  const { api, user } = setup()
  await user.click(await screen.findByRole('button', { name: messages.newBadge }))
  await user.type(screen.getByRole('textbox', { name: messages.name }), ' 新徽章 ')
  await user.click(screen.getByRole('button', { name: '王冠' }))
  await user.click(screen.getByText(messages.advancedSettings))
  expect(screen.getByRole<HTMLInputElement>('textbox', { name: messages.code }).readOnly).toBe(true)
  fireEvent.change(screen.getByRole('spinbutton', { name: messages.sort }), { target: { value: '0' } })
  await user.click(screen.getByRole('button', { name: messages.save }))
  await waitFor(() => expect(api.assets.saveBadge).toHaveBeenCalledWith(expect.objectContaining({ code: '', name: '新徽章', grantMode: 'manual', iconUrl: '/static/badges/king.svg', sortOrder: 0 })))
})

it('requires an explicit acknowledgement before disabling a badge already held by members', async () => {
  const { api, user } = setup()
  await user.click(await screen.findByRole('button', { name: '编辑 热心伙伴' }))
  await user.click(screen.getByRole('switch', { name: messages.enabled }))
  await user.click(screen.getByRole('button', { name: messages.save }))
  expect(api.assets.saveBadge).not.toHaveBeenCalled()
  await user.click(screen.getByRole('checkbox', { name: messages.disableConfirm.replace('{count}', '2') }))
  await user.click(screen.getByRole('button', { name: messages.save }))
  await waitFor(() => expect(api.assets.saveBadge).toHaveBeenCalledWith(expect.objectContaining({ isEnabled: false })))
})

it('corrects unsupported custom automatic badges and displays the actual rule for system badges', async () => {
  const { api, user } = setup([{ ...badge, grantMode: 'auto' }])
  await user.click(await screen.findByRole('button', { name: '编辑 热心伙伴' }))
  expect(screen.getByText(messages.customAutoCorrection)).toBeTruthy()
  await user.click(screen.getByRole('button', { name: messages.save }))
  await waitFor(() => expect(api.assets.saveBadge).toHaveBeenCalledWith(expect.objectContaining({ grantMode: 'manual' })))
  cleanup()
  const next = setup([{ ...badge, code: 'writer_10', type: 'system', grantMode: 'auto', grantRule: { metric: 'topics', threshold: 10 } }])
  await next.user.click(await screen.findByRole('button', { name: '编辑 热心伙伴' }))
  expect(screen.getByText('累计发布 10 篇主题后自动授予。')).toBeTruthy()
})

it('uploads a selected image and prevents saving a known broken image URL', async () => {
  const { api, user } = setup()
  await user.click(await screen.findByRole('button', { name: '编辑 热心伙伴' }))
  await user.click(screen.getByRole('button', { name: messages.uploadImage }))
  const file = new File(['png'], 'badge.png', { type: 'image/png' })
  await user.upload(screen.getByLabelText(messages.uploadImage, { selector: 'input' }), file)
  await waitFor(() => expect(api.pages.uploadImage).toHaveBeenCalledWith(file))
  await user.click(screen.getByRole('button', { name: messages.save }))
  await waitFor(() => expect(api.assets.saveBadge).toHaveBeenCalledWith(expect.objectContaining({ iconUrl: '/uploads/new-badge.png' })))
  await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull())
  await user.click(screen.getByRole('button', { name: '编辑 热心伙伴' }))
  await user.click(screen.getByRole('button', { name: messages.imageLink }))
  await user.clear(screen.getByRole('textbox', { name: messages.imageLink }))
  await user.type(screen.getByRole('textbox', { name: messages.imageLink }), 'https://example.test/broken.png')
  fireEvent.error(within(screen.getByRole('complementary', { name: messages.livePreview })).getAllByRole('img')[0])
  await user.click(screen.getByRole('button', { name: messages.save }))
  expect(api.assets.saveBadge).toHaveBeenCalledTimes(1)
  expect(screen.getAllByText(messages.badgeImageFailed).length).toBeGreaterThan(0)
  const imageUrl = screen.getByRole('textbox', { name: messages.imageLink })
  await user.clear(imageUrl)
  await user.type(imageUrl, 'https://example.test/replacement.png')
  await user.clear(imageUrl)
  await user.type(imageUrl, 'https://example.test/broken.png')
  fireEvent.load(within(screen.getByRole('complementary', { name: messages.livePreview })).getAllByRole('img')[0])
  await user.click(screen.getByRole('button', { name: messages.save }))
  await waitFor(() => expect(api.assets.saveBadge).toHaveBeenCalledTimes(2))
  expect(api.assets.saveBadge).toHaveBeenLastCalledWith(expect.objectContaining({ iconUrl: 'https://example.test/broken.png' }))
})

it('grants to a selected member and revokes with a required reason using individual operations', async () => {
  const { api, user } = setup()
  await user.click(await screen.findByRole('button', { name: '获奖成员 2' }))
  await screen.findByText('alice')
  await user.click(screen.getByRole('button', { name: messages.grantBadge }))
  let dialog = screen.getAllByRole('dialog').at(-1)!
  await user.type(within(dialog).getByRole('textbox', { name: messages.selectMember }), 'bob')
  await user.click(within(dialog).getByRole('button', { name: messages.search }))
  await user.click(await within(dialog).findByRole('button', { name: 'bob #9' }))
  await user.type(within(dialog).getByRole('textbox', { name: messages.grantReason }), ' 解答问题 ')
  await user.click(within(dialog).getByRole('button', { name: messages.grantBadge }))
  await waitFor(() => expect(api.assets.updateBadgeRecipient).toHaveBeenCalledWith({ code: 'helper', userId: 9, action: 'grant', reason: '解答问题' }))
  await waitFor(() => expect(screen.getAllByRole('dialog')).toHaveLength(1))
  await user.click(screen.getByRole('button', { name: messages.revokeBadge }))
  dialog = screen.getAllByRole('dialog').at(-1)!
  await user.click(within(dialog).getByRole('button', { name: messages.revokeBadge }))
  expect(api.assets.updateBadgeRecipient).toHaveBeenCalledTimes(1)
  await user.type(within(dialog).getByRole('textbox', { name: messages.revokeReason }), '误授予')
  await user.click(within(dialog).getByRole('button', { name: messages.revokeBadge }))
  await waitFor(() => expect(api.assets.updateBadgeRecipient).toHaveBeenLastCalledWith({ code: 'helper', userId: 7, action: 'revoke', reason: '误授予' }))
})

it('does not expose recipient management without user management permission', async () => {
  const badgeRecipients = vi.fn()
  const api = { assets: { badges: vi.fn().mockResolvedValue([badge]), badgeRecipients } } as unknown as GooseAdminApi
  render(<BadgesManagementPage api={api} text={text} />)
  const button = await screen.findByRole('button', { name: '获奖成员 2' })
  expect(button.hasAttribute('disabled')).toBe(true)
  expect(button.getAttribute('title')).toBe(messages.requiresUserManagement)
  await userEvent.click(button)
  expect(badgeRecipients).not.toHaveBeenCalled()
  expect(screen.queryByRole('dialog')).toBeNull()
})

it('uses hasNext for recipient navigation and omits totals for searched results', async () => {
  const { api, user } = setup()
  const recipient = { userId: 7, username: 'alice', avatarUrl: '', source: 'manual', reason: '提供帮助', grantedBy: 1, grantedByName: 'admin', grantedAt: '2026-11-17T02:00:00Z' }
  api.assets.badgeRecipients
    .mockResolvedValueOnce({ list: [recipient], total: 1000, hasNext: false, page: 1, size: 20 })
    .mockResolvedValueOnce({ list: [recipient], hasNext: true, page: 1, size: 20 })
    .mockResolvedValueOnce({ list: [{ ...recipient, userId: 8, username: 'alice-two' }], hasNext: false, page: 2, size: 20 })
  await user.click(await screen.findByRole('button', { name: '获奖成员 2' }))
  const dialog = screen.getByRole('dialog')
  await within(dialog).findByText('当前共 1000 位获奖成员')
  expect(within(dialog).getByRole('button', { name: messages.next }).hasAttribute('disabled')).toBe(true)
  await user.type(within(dialog).getByRole('textbox', { name: messages.recipientSearch }), 'alice')
  await user.click(within(dialog).getByRole('button', { name: messages.search }))
  await waitFor(() => expect(within(dialog).getByRole('button', { name: messages.next }).hasAttribute('disabled')).toBe(false))
  expect(within(dialog).queryByText(/当前共 .* 位获奖成员/)).toBeNull()
  expect(within(dialog).getByText('第 1 页')).toBeTruthy()
  await user.click(within(dialog).getByRole('button', { name: messages.next }))
  await within(dialog).findByText('alice-two')
  expect(api.assets.badgeRecipients).toHaveBeenLastCalledWith({ code: 'helper', page: 2, pageSize: 20, search: 'alice' })
  expect(within(dialog).getByText('第 2 页')).toBeTruthy()
  expect(within(dialog).getByRole('button', { name: messages.next }).hasAttribute('disabled')).toBe(true)
})
