import { computed, readonly, ref } from 'vue'
import { i18n } from './i18n'
import { resolveApiMessage } from './api-message'
import { setUnreadMessagesDocumentTitle } from './document-title'
import { createGooseClient, GooseClientError, type UnreadStatusPayload } from '@gooseforum/client'

const CACHE_KEY = 'goose:unread-status'
const CACHE_TTL = 10_000
const POLL_INTERVAL = 30_000
const client = createGooseClient()

const notifications = ref(false)
const messages = ref(false)
const moderationReports = ref(false)
const latestNotificationType = ref('')
const checked = ref(false)
let inFlight: Promise<UnreadStatusPayload> | null = null
let pollTimer: number | undefined

const notificationMessage = computed(() => {
  if (latestNotificationType.value === 'comment' || latestNotificationType.value === 'topic_post') return i18n.global.t('notifications.newComment')
  if (notifications.value) return i18n.global.t('notifications.newNotification')
  return i18n.global.t('notifications.noUnread')
})

function normalizeStatus(data: Partial<UnreadStatusPayload> | null | undefined): UnreadStatusPayload {
  return {
    notifications: Boolean(data?.notifications),
    messages: Boolean(data?.messages),
    moderationReports: Boolean(data?.moderationReports),
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
  moderationReports.value = status.moderationReports || false
  latestNotificationType.value = status.latestNotificationType || ''
  checked.value = true
  setUnreadMessagesDocumentTitle(status.messages)
  writeCache(status)
}

async function fetchUnreadStatus() {
  try {
    return normalizeStatus(await client.api.notifications.unread())
  } catch (error) {
    if (error instanceof GooseClientError && error.messageCode) {
      throw new Error(resolveApiMessage(error, i18n.global.t('notifications.checkFailed')))
    }
    throw error
  }
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
  const hasInitial = initial !== undefined && initial !== null
  if (hasInitial) applyUnread(initial)
  if (pollTimer !== undefined) return
  if (!hasInitial) {
    const cached = readCache()
    if (cached) applyUnread(cached)
    void refresh(true)
  }
  pollTimer = window.setInterval(() => {
    void refresh(true)
  }, POLL_INTERVAL)
}

function clearNotifications() {
  applyUnread({
    notifications: false,
    messages: messages.value,
    moderationReports: moderationReports.value,
  })
}

function setNotifications(hasUnread: boolean) {
  applyUnread({
    notifications: hasUnread,
    messages: messages.value,
    moderationReports: moderationReports.value,
    latestNotificationType: hasUnread ? latestNotificationType.value : '',
  })
}

function clearMessages() {
  applyUnread({
    notifications: notifications.value,
    messages: false,
    moderationReports: moderationReports.value,
    latestNotificationType: latestNotificationType.value,
  })
}

export function useUnreadStatus() {
  return {
    notifications: readonly(notifications),
    messages: readonly(messages),
    moderationReports: readonly(moderationReports),
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
