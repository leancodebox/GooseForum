import { describe, expect, it } from 'vitest'
import { chooseUserCardSide } from '@/runtime/user-card'

describe('chooseUserCardSide', () => {
  it('uses the space below when the card fits', () => {
    expect(chooseUserCardSide({ top: 100, bottom: 140 }, 800)).toBe('bottom')
  })

  it('uses the space above near the bottom of the viewport', () => {
    expect(chooseUserCardSide({ top: 700, bottom: 740 }, 800)).toBe('top')
  })

  it('uses the larger side when the card fits neither side', () => {
    expect(chooseUserCardSide({ top: 260, bottom: 300 }, 500)).toBe('top')
    expect(chooseUserCardSide({ top: 200, bottom: 240 }, 500)).toBe('bottom')
  })
})
