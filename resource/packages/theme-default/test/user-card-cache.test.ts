import { expect, it, vi } from 'vitest'
import type { GooseSiteApi, UserCardPayload } from '@gooseforum/client'
import { getCachedUserCard, invalidateUserCard, loadUserCard } from '../src/site/users/user-card-cache'

it('coalesces requests and does not cache an invalidated in-flight response', async () => {
  let resolveOld!: (value: UserCardPayload) => void
  const old = new Promise<UserCardPayload>(resolve => { resolveOld = resolve })
  const fresh = { userId: 876543, profileCoverUrl: '/new.webp' } as UserCardPayload
  const card = vi.fn().mockReturnValueOnce(old).mockResolvedValueOnce(fresh)
  const api = { users: { card } } as unknown as GooseSiteApi
  const first = loadUserCard(api, fresh.userId)
  expect(loadUserCard(api, fresh.userId)).toBe(first)
  invalidateUserCard(fresh.userId)
  const second = loadUserCard(api, fresh.userId)
  resolveOld({ ...fresh, profileCoverUrl: '/old.webp' })
  await first
  expect(getCachedUserCard(fresh.userId)).toBe(fresh)
  expect(await second).toBe(fresh)
  expect(card).toHaveBeenCalledTimes(2)
  invalidateUserCard(fresh.userId)
})

it('bounds the card cache and keeps recently used cards', async () => {
  const api = { users: { card: vi.fn(async (userId: number) => ({ userId } as UserCardPayload)) } } as unknown as GooseSiteApi
  for (let userId = 1; userId <= 128; userId++) await loadUserCard(api, userId)
  expect(getCachedUserCard(1)?.userId).toBe(1)
  await loadUserCard(api, 129)
  expect(getCachedUserCard(1)?.userId).toBe(1)
  expect(getCachedUserCard(2)).toBeUndefined()
  expect(getCachedUserCard(129)?.userId).toBe(129)
  invalidateUserCard(1)
  invalidateUserCard(129)
})
