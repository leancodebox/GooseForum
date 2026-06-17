import { ref } from 'vue'

export type TopicListMode = 'waterfall' | 'pagination'

const storageKey = 'goose:topic-list-mode'

function initialTopicListMode(): TopicListMode {
  if (typeof window === 'undefined') return 'waterfall'
  try {
    return window.localStorage.getItem(storageKey) === 'pagination' ? 'pagination' : 'waterfall'
  } catch {
    return 'waterfall'
  }
}

export function useTopicListMode() {
  const mode = ref<TopicListMode>(initialTopicListMode())

  function setMode(nextMode: TopicListMode) {
    mode.value = nextMode
    try {
      window.localStorage.setItem(storageKey, nextMode)
    } catch {
      // Preference persistence is best-effort; the current page can still switch modes.
    }
  }

  return {
    mode,
    setMode,
  }
}
