import { describe, expect, it } from 'vitest'
import { errorMessage } from './error-message'

describe('errorMessage', () => {
  it('keeps useful Error messages', () => {
    expect(errorMessage(new Error('server rejected request'), 'fallback')).toBe('server rejected request')
  })

  it('uses the fallback for non-Error rejection values', () => {
    expect(errorMessage({ message: 'hidden object' }, 'Unable to load')).toBe('Unable to load')
    expect(errorMessage(null, 'Unable to load')).toBe('Unable to load')
  })
})
