export interface TopicCategoryChoice {
  id: number
  isRestricted: boolean
  canCreate: boolean
}

export function addTopicCategory(
  selected: number[],
  category: TopicCategoryChoice,
  categories: TopicCategoryChoice[],
  maxCategories = 3,
) {
  if (selected.includes(category.id) || !category.canCreate) return [...selected]
  const hasRestrictedSelection = categories.some((item) => item.isRestricted && selected.includes(item.id))
  if (category.isRestricted || hasRestrictedSelection) return [category.id]
  if (selected.length >= maxCategories) return [...selected]
  return [...selected, category.id]
}

export function isTopicCategoryAdditionDisabled(
  selected: number[],
  category: TopicCategoryChoice,
  categories: TopicCategoryChoice[],
  maxCategories = 3,
) {
  if (selected.includes(category.id)) return false
  if (!category.canCreate) return true
  if (category.isRestricted) return false
  const hasRestrictedSelection = categories.some((item) => item.isRestricted && selected.includes(item.id))
  return !hasRestrictedSelection && selected.length >= maxCategories
}
