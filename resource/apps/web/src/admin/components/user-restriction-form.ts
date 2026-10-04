import type { AdminUser } from '@gooseforum/client'

export interface RestrictionForm { status: 'normal' | 'suspended' | 'banned'; until: string; reason: string }

function localDate(value?: string | null) {
  if (!value) return ''
  const date = new Date(value)
  if (!Number.isFinite(date.getTime())) return ''
  const offset = date.getTimezoneOffset() * 60_000
  return new Date(date.getTime() - offset).toISOString().slice(0, 16)
}

export function restrictionForm(user: AdminUser | null): RestrictionForm {
  return { status: user?.restrictionStatus || 'normal', until: localDate(user?.restrictionUntil), reason: user?.restrictionReason || '' }
}
