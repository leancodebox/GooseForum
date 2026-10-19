import { trustLevels, type TrustLimits } from './trust.js'
import { trustQuotaFields, type TrustLimitField } from './trust-limits.js'

export function trustLimitOrderIssue(limits: readonly TrustLimits[]): { field: TrustLimitField; lowerLevel: number; higherLevel: number } | undefined {
  const byLevel = new Map(limits.map(limit => [limit.level, limit]))
  for (let i = 1; i < trustLevels.length; i++) {
    const lower = byLevel.get(trustLevels[i - 1])
    const higher = byLevel.get(trustLevels[i])
    if (!lower || !higher) continue
    if (lower.allowImageUpload && !higher.allowImageUpload) return { field: 'allowImageUpload', lowerLevel: lower.level, higherLevel: higher.level }
    for (const field of trustQuotaFields) {
      const lowerValue = lower[field], higherValue = higher[field]
      if (higherValue !== null && (lowerValue === null || higherValue < lowerValue)) return { field, lowerLevel: lower.level, higherLevel: higher.level }
    }
  }
}
