import { readonly, ref } from 'vue'
import { i18n } from './i18n'

const CACHE_KEY = 'goose:last-unread-notification'
const CACHE_TTL = 10_000
const POLL_INTERVAL = 30_000

interface LastUnreadNotification {
  eventType?: string
}

const hasUnread = ref(false)
const message = ref(i18n.global.t('notifications.noUnread'))
const checked = ref(false)
let inFlight: Promise<LastUnreadNotification> | null = null
let pollTimer: number | undefined

function readCache(): LastUnreadNotification | null {
  try {
    const raw = localStorage.getItem(CACHE_KEY)
    if (!raw) return null
    const cached = JSON.parse(raw) as { data?: LastUnreadNotification; timestamp?: number }
    if (!cached.timestamp || Date.now() - cached.timestamp > CACHE_TTL) return null
    return cached.data || null
  } catch {
    return null
  }
}

function writeCache(data: LastUnreadNotification) {
  try {
    localStorage.setItem(CACHE_KEY, JSON.stringify({
      data,
      timestamp: Date.now(),
    }))
  } catch {
    // Ignore storage errors; notification polling should never break navigation.
  }
}

function applyUnread(data: LastUnreadNotification | null) {
  const eventType = data?.eventType || ''
  hasUnread.value = Boolean(eventType)
  message.value = eventType === 'comment'
    ? i18n.global.t('notifications.newComment')
    : eventType
      ? i18n.global.t('notifications.newNotification')
      : i18n.global.t('notifications.noUnread')
  checked.value = true
}

async function fetchLastUnread() {
  const response = await fetch('/api/forum/notification/last-unread', {
    headers: {
      Accept: 'application/json',
    },
  })
  if (!response.ok) throw new Error(`HTTP ${response.status}`)
  const data = await response.json() as { code?: number; result?: LastUnreadNotification; data?: LastUnreadNotification }
  if (data.code !== undefined && data.code !== 0) throw new Error(i18n.global.t('notifications.checkFailed'))
  return data.result ?? data.data ?? {}
}

async function refresh(force = false) {
  if (!force) {
    const cached = readCache()
    if (cached) {
      applyUnread(cached)
      void refresh(true)
      return cached
    }
  }

  if (!inFlight) {
    inFlight = fetchLastUnread()
      .then((data) => {
        writeCache(data)
        applyUnread(data)
        return data
      })
      .finally(() => {
        inFlight = null
      })
  }
  return inFlight
}

function startPolling() {
  if (pollTimer !== undefined) return
  const cached = readCache()
  if (cached) applyUnread(cached)
  void refresh(true)
  pollTimer = window.setInterval(() => {
    void refresh(true)
  }, POLL_INTERVAL)
}

function clearUnread() {
  hasUnread.value = false
  message.value = i18n.global.t('notifications.noUnread')
  checked.value = true
  writeCache({})
}

export function useUnreadNotifications() {
  return {
    hasUnread: readonly(hasUnread),
    message: readonly(message),
    checked: readonly(checked),
    startPolling,
    refresh,
    clearUnread,
  }
}
