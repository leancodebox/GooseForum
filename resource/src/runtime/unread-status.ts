import { computed, readonly, ref } from 'vue'
import { i18n } from './i18n'
import { setUnreadMessagesDocumentTitle } from './document-title'
import type { UnreadStatusPayload } from '@/types/payload'

const CACHE_KEY = 'goose:unread-status'
const CACHE_TTL = 10_000
const POLL_INTERVAL = 30_000

const notifications = ref(false)
const messages = ref(false)
const latestNotificationType = ref('')
const checked = ref(false)
let inFlight: Promise<UnreadStatusPayload> | null = null
let pollTimer: number | undefined

const notificationMessage = computed(() => {
  if (latestNotificationType.value === 'comment') return i18n.global.t('notifications.newComment')
  if (notifications.value) return i18n.global.t('notifications.newNotification')
  return i18n.global.t('notifications.noUnread')
})

function normalizeStatus(data: Partial<UnreadStatusPayload> | null | undefined): UnreadStatusPayload {
  return {
    notifications: Boolean(data?.notifications),
    messages: Boolean(data?.messages),
    latestNotificationType: data?.latestNotificationType || '',
  }
}

function readCache(): UnreadStatusPayload | null {
  try {
    const raw = localStorage.getItem(CACHE_KEY)
    if (!raw) return null
    const cached = JSON.parse(raw) as { data?: UnreadStatusPayload; timestamp?: number }
    if (!cached.timestamp || Date.now() - cached.timestamp > CACHE_TTL) return null
    return normalizeStatus(cached.data)
  } catch {
    return null
  }
}

function writeCache(data: UnreadStatusPayload) {
  try {
    localStorage.setItem(CACHE_KEY, JSON.stringify({
      data,
      timestamp: Date.now(),
    }))
  } catch {
    // Ignore storage errors; unread polling should never break navigation.
  }
}

function applyUnread(data: Partial<UnreadStatusPayload> | null | undefined) {
  const status = normalizeStatus(data)
  notifications.value = status.notifications
  messages.value = status.messages
  latestNotificationType.value = status.latestNotificationType || ''
  checked.value = true
  setUnreadMessagesDocumentTitle(status.messages)
  writeCache(status)
}

async function fetchUnreadStatus() {
  const response = await fetch('/api/forum/unread-status', {
    headers: {
      Accept: 'application/json',
    },
  })
  if (!response.ok) throw new Error(`HTTP ${response.status}`)
  const data = await response.json() as { code?: number; result?: UnreadStatusPayload; data?: UnreadStatusPayload }
  if (data.code !== undefined && data.code !== 0) throw new Error(i18n.global.t('notifications.checkFailed'))
  return normalizeStatus(data.result ?? data.data)
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
    inFlight = fetchUnreadStatus()
      .then((data) => {
        applyUnread(data)
        return data
      })
      .finally(() => {
        inFlight = null
      })
  }
  return inFlight
}

function startPolling(initial?: Partial<UnreadStatusPayload>) {
  if (initial) applyUnread(initial)
  if (pollTimer !== undefined) return
  const cached = readCache()
  if (cached) applyUnread(cached)
  void refresh(true)
  pollTimer = window.setInterval(() => {
    void refresh(true)
  }, POLL_INTERVAL)
}

function clearNotifications() {
  applyUnread({
    notifications: false,
    messages: messages.value,
  })
}

function setNotifications(hasUnread: boolean) {
  applyUnread({
    notifications: hasUnread,
    messages: messages.value,
    latestNotificationType: hasUnread ? latestNotificationType.value : '',
  })
}

function clearMessages() {
  applyUnread({
    notifications: notifications.value,
    messages: false,
    latestNotificationType: latestNotificationType.value,
  })
}

export function useUnreadStatus() {
  return {
    notifications: readonly(notifications),
    messages: readonly(messages),
    latestNotificationType: readonly(latestNotificationType),
    checked: readonly(checked),
    notificationMessage,
    startPolling,
    refresh,
    applyUnread,
    clearNotifications,
    setNotifications,
    clearMessages,
  }
}
