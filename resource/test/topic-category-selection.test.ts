import { describe, expect, it } from 'vitest'
import { addTopicCategory, isTopicCategoryAdditionDisabled, type TopicCategoryChoice } from '../src/runtime/topic-category-selection'

const categories: TopicCategoryChoice[] = [
  { id: 1, isRestricted: false, canCreate: true },
  { id: 2, isRestricted: false, canCreate: true },
  { id: 3, isRestricted: false, canCreate: true },
  { id: 4, isRestricted: false, canCreate: true },
  { id: 9, isRestricted: true, canCreate: true },
  { id: 10, isRestricted: true, canCreate: false },
]

describe('topic category selection', () => {
  it('keeps public categories multi-select up to the limit', () => {
    expect(addTopicCategory([1, 2], categories[2], categories)).toEqual([1, 2, 3])
    expect(addTopicCategory([1, 2, 3], categories[3], categories)).toEqual([1, 2, 3])
    expect(isTopicCategoryAdditionDisabled([1, 2, 3], categories[3], categories)).toBe(true)
  })

  it('replaces public categories when a restricted category is selected', () => {
    expect(addTopicCategory([1, 2], categories[4], categories)).toEqual([9])
    expect(isTopicCategoryAdditionDisabled([1, 2, 3], categories[4], categories)).toBe(false)
  })

  it('replaces a restricted selection when switching back to public', () => {
    expect(addTopicCategory([9], categories[0], categories)).toEqual([1])
    expect(isTopicCategoryAdditionDisabled([9], categories[0], categories)).toBe(false)
  })

  it('never selects a category without create capability', () => {
    expect(addTopicCategory([1], categories[5], categories)).toEqual([1])
    expect(isTopicCategoryAdditionDisabled([1], categories[5], categories)).toBe(true)
  })
})
