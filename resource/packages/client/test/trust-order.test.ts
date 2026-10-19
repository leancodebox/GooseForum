import { describe, expect, it } from 'vitest'
import { trustLevels, type TrustLimits } from '../src/trust.js'
import { trustQuotaFields } from '../src/trust-limits.js'
import { trustLimitOrderIssue } from '../src/trust-validation.js'

function limits(): TrustLimits[] {
  return trustLevels.map(level => ({ level, allowImageUpload: true, maxDailyTopics: null, maxDailyReplies: null, maxDailyUploads: null, maxImagesPerPost: null, maxExternalLinksPerPost: null, maxMentionsPerPost: null }))
}

describe('trust limit order', () => {
  for (const field of trustQuotaFields) {
    for (const [lower, higher, invalid] of [[10, 9, true], [null, 10, true], [1, 0, true], [0, 0, false], [0, 1, false], [10, null, false], [null, null, false]] as const) {
      it(`${field}: ${lower} -> ${higher}`, () => {
        const config = limits()
        config[0][field] = lower; config[1][field] = higher
        const issue = trustLimitOrderIssue(config.reverse())
        expect(Boolean(issue)).toBe(invalid)
        if (invalid) expect(issue).toEqual({ field, lowerLevel: 0, higherLevel: 100 })
      })
    }
  }
  it.each([1, 2, 3])('rejects upload permission removal at index %i', index => {
    const config = limits(); config[index].allowImageUpload = false
    expect(trustLimitOrderIssue(config)).toEqual({ field: 'allowImageUpload', lowerLevel: trustLevels[index - 1], higherLevel: trustLevels[index] })
  })
  it('accepts upload permission expansion', () => {
    const config = limits(); config[0].allowImageUpload = false
    expect(trustLimitOrderIssue(config)).toBeUndefined()
  })
})
