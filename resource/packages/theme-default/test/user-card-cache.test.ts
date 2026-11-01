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
