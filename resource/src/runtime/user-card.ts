import type { UserBadgePayload, UserCardPayload } from '@gooseforum/client'

export interface UserCardTarget {
  id: number
  username: string
  avatarUrl: string
  wornBadge?: UserBadgePayload | null
}

export const userCardCache = new Map<number, UserCardPayload>()

export type UserCardSide = 'top' | 'bottom'

export function chooseUserCardSide(
  bounds: Pick<DOMRect, 'top' | 'bottom'>,
  viewportHeight: number,
  estimatedCardHeight = 340,
): UserCardSide {
  const clearance = 22
  const spaceBelow = viewportHeight - bounds.bottom - clearance
  const spaceAbove = bounds.top - clearance
  return spaceBelow >= estimatedCardHeight || spaceBelow >= spaceAbove ? 'bottom' : 'top'
}
