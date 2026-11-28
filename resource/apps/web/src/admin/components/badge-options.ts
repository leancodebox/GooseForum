import type { AdminBadge } from '@gooseforum/client'
import type { AssetText, AssetTextKey } from '../assets-i18n'

export type BadgeText = AssetText

export const badgeColors = [
  'blue', 'emerald', 'teal', 'sky', 'cyan', 'rose', 'violet',
  'purple', 'fuchsia', 'indigo', 'amber', 'orange', 'yellow', 'slate',
] as const

export const badgeColorValues: Record<(typeof badgeColors)[number], string> = {
  blue: '#2563eb', emerald: '#059669', teal: '#0d9488', sky: '#0284c7',
  cyan: '#0891b2', rose: '#e11d48', violet: '#7c3aed', purple: '#9333ea',
  fuchsia: '#c026d3', indigo: '#4f46e5', amber: '#d97706', orange: '#ea580c',
  yellow: '#ca8a04', slate: '#64748b',
}

export const badgeLevels = ['bronze', 'silver', 'gold', 'special'] as const
export const builtinBadgeIcons = [
  'first-post', 'first-comment', 'first-like-given', 'first-follower',
  'writer-10', 'commenter-50', 'liked-10', 'popular-100', 'social-10',
  'early-member', 'contributor', 'moderator', 'sponsor', 'king', 'robot',
] as const

export function badgeLevelLabel(level: string, text: BadgeText) {
  return badgeLevels.includes(level as (typeof badgeLevels)[number])
    ? text(`level_${level}` as AssetTextKey) : level
}

export function badgeRuleLabel(badge: AdminBadge, text: BadgeText) {
  if (badge.grantMode !== 'auto') return text('manualGrantHint')
  if (!badge.grantRule) return text('unsupportedAutomatic')
  return text(`rule_${badge.grantRule.metric}`, { count: badge.grantRule.threshold })
}

export const newBadge: AdminBadge = {
  code: '', type: 'custom', grantMode: 'manual', name: '', description: '',
  iconType: 'asset', iconKey: '', iconUrl: '/static/badges/contributor.svg',
  color: 'blue', level: 'bronze', isEnabled: true, isWearable: true, sortOrder: 1000,
}
