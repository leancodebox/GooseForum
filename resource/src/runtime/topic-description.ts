import type { TopicPayload } from '@/types/payload'

const fallbackDescriptions = [
  '(｀・ω・´)',
  '( ´ ▽ ` )ﾉ',
  '(ง •̀_•́)ง',
  '(｡･ω･｡)',
  '(￣▽￣)ノ',
  '(っ´ω`)っ',
  '( •̀ ω •́ )✧',
  '(๑•̀ㅂ•́)و✧',
]

export function topicDescription(topic: TopicPayload) {
  const description = topic.description?.trim()
  if (description) return description
  return fallbackDescriptions[Math.abs(topic.id) % fallbackDescriptions.length]
}
