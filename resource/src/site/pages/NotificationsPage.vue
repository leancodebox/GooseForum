<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { Award, Bell, Check, CheckCheck, Heart, Info, MessageCircle, UserPlus } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { fetchNotifications, markAllNotificationsRead, markNotificationRead } from '@/runtime/api'
import { formatDateTime } from '@/runtime/format'
import { useUnreadStatus } from '@/runtime/unread-status'
import type { LayoutPayload, NotificationFilter, NotificationPayload, NotificationsPageProps } from '@/types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: NotificationsPageProps
}>()

interface NotificationListState {
  items: NotificationPayload[]
  nextCursor: number
  hasNext: boolean
  loading: boolean
  loaded: boolean
}

const pageSize = 20
const activeFilter = ref<NotificationFilter>('all')
const unreadCount = ref(page.props.unreadCount)
const markingAllRead = ref(false)
const actionError = ref('')
const loadMoreEl = ref<HTMLElement | null>(null)
const { t } = useI18n()
const unreadStatus = useUnreadStatus()
let loadObserver: IntersectionObserver | undefined

const lists = reactive<Record<NotificationFilter, NotificationListState>>({
  all: {
    items: copyNotifications(page.props.notifications),
    nextCursor: initialNextCursor(),
    hasNext: page.props.pagination.hasNext,
    loading: false,
    loaded: true,
  },
  unread: {
    items: [],
    nextCursor: 0,
    hasNext: true,
    loading: false,
    loaded: false,
  },
})

const hasUnread = computed(() => unreadCount.value > 0)
const activeList = computed(() => lists[activeFilter.value])
const notifications = computed(() => activeList.value.items)
const loadedCount = computed(() => activeList.value.items.length)
const emptyTitle = computed(() => activeFilter.value === 'unread' ? t('notifications.unreadEmptyTitle') : t('notifications.emptyTitle'))
const emptyDescription = computed(() => activeFilter.value === 'unread' ? t('notifications.unreadEmptyDescription') : t('notifications.emptyDescription'))

if (page.props.unreadCount === 0) {
  unreadStatus.clearNotifications()
}

watch(
  () => page.props.notifications,
  () => {
    lists.all.items = copyNotifications(page.props.notifications)
    lists.all.nextCursor = initialNextCursor()
    lists.all.hasNext = page.props.pagination.hasNext
    lists.all.loaded = true
    lists.all.loading = false
    lists.unread.items = []
    lists.unread.nextCursor = 0
    lists.unread.hasNext = true
    lists.unread.loaded = false
    lists.unread.loading = false
    unreadCount.value = page.props.unreadCount
    actionError.value = ''
    if (page.props.unreadCount === 0) {
      unreadStatus.clearNotifications()
    }
    void nextTick(observeLoadMore)
  },
)

watch(activeFilter, async (filter) => {
  actionError.value = ''
  if (!lists[filter].loaded) {
    await loadNotifications(filter, true)
  }
  await nextTick()
  observeLoadMore()
})

onMounted(() => {
  observeLoadMore()
})

onBeforeUnmount(() => {
  loadObserver?.disconnect()
})

function copyNotifications(items: NotificationPayload[]) {
  return items.map((item) => ({ ...item }))
}

function initialNextCursor() {
  if (!page.props.pagination.hasNext) return 0
  const url = page.props.pagination.nextUrl
  if (!url) {
    const last = lastNotification(page.props.notifications)
    return last?.id ?? 0
  }
  try {
    const origin = typeof window !== 'undefined' ? window.location.origin : 'http://localhost'
    return Number(new URL(url, origin).searchParams.get('cursor')) || 0
  } catch {
    const last = lastNotification(page.props.notifications)
    return last?.id ?? 0
  }
}

function lastNotification(items: NotificationPayload[]) {
  return items.length > 0 ? items[items.length - 1] : undefined
}

function observeLoadMore() {
  loadObserver?.disconnect()
  if (!loadMoreEl.value || !activeList.value.hasNext || !('IntersectionObserver' in window)) return
  loadObserver = new IntersectionObserver((entries) => {
    if (entries[0]?.isIntersecting) {
      void loadNotifications(activeFilter.value)
    }
  }, { rootMargin: '160px 0px' })
  loadObserver.observe(loadMoreEl.value)
}

async function loadNotifications(filter: NotificationFilter, reset = false) {
  const list = lists[filter]
  if (list.loading) return
  if (!reset && !list.hasNext) return

  list.loading = true
  actionError.value = ''
  try {
    const payload = await fetchNotifications(filter, reset ? 0 : list.nextCursor, pageSize)
    const nextItems = payload.items.map((item) => ({ ...item }))
    list.items = reset ? nextItems : mergeNotifications(list.items, nextItems)
    list.nextCursor = payload.nextCursor
    list.hasNext = payload.hasNext
    list.loaded = true
    unreadCount.value = payload.unreadCount
    unreadStatus.setNotifications(payload.unreadCount > 0)
  } catch (error) {
    actionError.value = error instanceof Error ? error.message : t('api.notificationsLoadFailed')
  } finally {
    list.loading = false
    await nextTick()
    observeLoadMore()
  }
}

function mergeNotifications(current: NotificationPayload[], incoming: NotificationPayload[]) {
  const seen = new Set(current.map((item) => item.id))
  const merged = [...current]
  for (const item of incoming) {
    if (seen.has(item.id)) continue
    seen.add(item.id)
    merged.push(item)
  }
  return merged
}

function setActiveFilter(filter: NotificationFilter) {
  activeFilter.value = filter
}

function notificationIcon(item: NotificationPayload) {
  if (item.eventType === 'like') return Heart
  if (item.eventType === 'follow') return UserPlus
  if (item.eventType === 'badge') return Award
  if (item.eventType === 'system') return Info
  return MessageCircle
}

function notificationText(item: NotificationPayload) {
  const templateText = notificationTemplateText(item)
  if (item.eventType === 'badge') {
    return templateText || (item.payload.metadata?.badgeName
      ? t('notifications.badgeEarned', { badge: item.payload.metadata.badgeName })
      : item.content || item.payload.content || t('notifications.fallback'))
  }
  if (item.eventType === 'follow') {
    return templateText || item.content || item.payload.content || t('notifications.followDescription', { actor: actorName(item) })
  }
  if (item.article) {
    return item.article.title
  }
  if (templateText) return templateText
  return item.content || item.payload.content || t('notifications.fallback')
}

function notificationTemplateText(item: NotificationPayload) {
  switch (item.payload.templateKey) {
    case 'notifications.templates.comment':
      return t('notifications.templates.comment')
    case 'notifications.templates.reply':
      return t('notifications.templates.reply')
    case 'notifications.templates.articleComment':
      return t('notifications.templates.articleComment')
    case 'notifications.templates.follow':
      return t('notifications.templates.follow')
    case 'notifications.templates.badge':
      return t('notifications.templates.badge', { badge: item.payload.templateParams?.badgeName || item.payload.metadata?.badgeName || '' })
    default:
      return ''
  }
}

function notificationVerb(item: NotificationPayload) {
  const templateText = notificationTemplateText(item)
  if (templateText && item.eventType !== 'badge') return templateText
  if (item.eventType === 'like') return t('notifications.verb.like')
  if (item.eventType === 'follow') return t('notifications.verb.follow')
  if (item.eventType === 'badge') return ''
  if (item.eventType === 'reply') return t('notifications.verb.reply')
  if (item.eventType === 'comment' || item.eventType === 'article_comment') return t('notifications.verb.comment')
  return item.title
}

function notificationTone(item: NotificationPayload) {
  if (item.eventType === 'like') return item.isRead ? 'text-gray-400' : 'text-rose-600'
  if (item.eventType === 'follow') return item.isRead ? 'text-gray-400' : 'text-emerald-600'
  if (item.eventType === 'badge') return item.isRead ? 'text-gray-400' : 'text-amber-600'
  if (item.eventType === 'system') return item.isRead ? 'text-gray-400' : 'text-amber-600'
  return item.isRead ? 'text-gray-400' : 'text-blue-600'
}

function actorName(item: NotificationPayload) {
  if (item.eventType === 'badge') return notificationTemplateText(item) || item.title || t('notifications.actorFallback')
  return item.actor.username || item.payload.actorName || item.payload.metadata?.followerName || t('notifications.actorFallback')
}

function actorURL(item: NotificationPayload) {
  return item.actor.id ? `/u/${item.actor.id}` : ''
}

function targetURL(item: NotificationPayload) {
  if (item.article) return item.article.url
  if (item.eventType === 'badge') return item.payload.metadata?.profileUrl || actorURL(item)
  if (item.eventType === 'follow') return actorURL(item)
  return ''
}

async function markAllRead() {
  if (!hasUnread.value || markingAllRead.value) return

  const previousAll = copyNotifications(lists.all.items)
  const previousUnread = copyNotifications(lists.unread.items)
  const previousUnreadState = {
    nextCursor: lists.unread.nextCursor,
    hasNext: lists.unread.hasNext,
    loaded: lists.unread.loaded,
  }
  const previousUnreadCount = unreadCount.value
  markingAllRead.value = true
  actionError.value = ''
  lists.all.items = lists.all.items.map((item) => ({ ...item, isRead: true }))
  lists.unread.items = []
  lists.unread.nextCursor = 0
  lists.unread.hasNext = false
  lists.unread.loaded = true
  unreadCount.value = 0
  try {
    await markAllNotificationsRead()
    unreadStatus.clearNotifications()
  } catch (error) {
    lists.all.items = previousAll
    lists.unread.items = previousUnread
    lists.unread.nextCursor = previousUnreadState.nextCursor
    lists.unread.hasNext = previousUnreadState.hasNext
    lists.unread.loaded = previousUnreadState.loaded
    unreadCount.value = previousUnreadCount
    actionError.value = error instanceof Error ? error.message : t('notifications.markAllReadFailed')
  } finally {
    markingAllRead.value = false
  }
}

function markItemRead(item: NotificationPayload) {
  if (item.isRead) return

  const previousAll = copyNotifications(lists.all.items)
  const previousUnread = copyNotifications(lists.unread.items)
  const previousUnreadCount = unreadCount.value
  const notificationId = item.id

  lists.all.items = lists.all.items.map((notification) => notification.id === notificationId ? { ...notification, isRead: true } : notification)
  lists.unread.items = lists.unread.items.filter((notification) => notification.id !== notificationId)
  unreadCount.value = Math.max(unreadCount.value - 1, 0)
  unreadStatus.setNotifications(unreadCount.value > 0)

  void markNotificationRead(notificationId).catch(() => {
    lists.all.items = previousAll
    lists.unread.items = previousUnread
    unreadCount.value = previousUnreadCount
    void unreadStatus.refresh(true).catch(() => undefined)
  })
}

function markItemReadAndNavigate(item: NotificationPayload) {
  markItemRead(item)
}
</script>

<template>
  <main class="min-w-0 pb-8">
    <header class="mb-3 flex flex-col gap-2 border-b border-gray-200/70 pb-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="min-w-0">
        <div class="flex min-w-0 items-center gap-2">
          <h1 class="text-xl font-bold text-gray-950">{{ t('notifications.title') }}</h1>
          <span
            v-if="unreadCount"
            class="inline-flex h-5 items-center rounded-full bg-blue-50 px-2 text-xs font-semibold tabular-nums text-blue-700"
          >
            {{ t('notifications.unread', { count: unreadCount }) }}
          </span>
        </div>
        <p class="mt-0.5 text-xs text-gray-500">{{ t('notifications.summary', { total: loadedCount }) }}</p>
        <p v-if="actionError" class="mt-1 text-xs text-red-600">{{ actionError }}</p>
      </div>
      <button
        type="button"
        class="inline-flex h-8 w-fit items-center gap-1.5 rounded-md border border-gray-200 bg-white px-2.5 text-xs font-semibold text-gray-600 hover:bg-gray-50 hover:text-gray-900 disabled:cursor-not-allowed disabled:opacity-45"
        :disabled="!hasUnread || markingAllRead"
        @click="markAllRead"
      >
        <CheckCheck class="h-4 w-4" />
        {{ markingAllRead ? t('common.loading') : t('notifications.markAllRead') }}
      </button>
    </header>

    <section class="overflow-hidden rounded-lg border border-gray-200/70 bg-white shadow-[0_2px_8px_rgba(0,0,0,0.02)]">
      <div class="flex items-center gap-1 border-b border-gray-100 bg-gray-50/60 p-2">
        <button
          type="button"
          class="inline-flex h-8 items-center rounded-md px-3 text-sm font-semibold transition"
          :class="activeFilter === 'all' ? 'bg-white text-gray-950 shadow-sm ring-1 ring-gray-200' : 'text-gray-500 hover:bg-white/70 hover:text-gray-800'"
          @click="setActiveFilter('all')"
        >
          {{ t('notifications.tabs.all') }}
        </button>
        <button
          type="button"
          class="inline-flex h-8 items-center gap-1.5 rounded-md px-3 text-sm font-semibold transition"
          :class="activeFilter === 'unread' ? 'bg-white text-gray-950 shadow-sm ring-1 ring-gray-200' : 'text-gray-500 hover:bg-white/70 hover:text-gray-800'"
          @click="setActiveFilter('unread')"
        >
          {{ t('notifications.tabs.unread') }}
          <span v-if="unreadCount" class="rounded-full bg-blue-50 px-1.5 text-[11px] font-bold tabular-nums text-blue-700">{{ unreadCount }}</span>
        </button>
      </div>

      <div class="hidden grid-cols-[34px_minmax(0,1fr)_116px] gap-3 border-b border-gray-100 bg-gray-50/60 px-3 py-2 text-[11px] font-bold uppercase text-gray-600 md:grid">
        <div />
        <div>{{ t('notifications.table.notification') }}</div>
        <div class="text-right">{{ t('notifications.table.time') }}</div>
      </div>

      <div v-if="notifications.length" class="divide-y divide-gray-100">
        <article
          v-for="item in notifications"
          :key="item.id"
          class="relative grid gap-3 px-3 py-2.5 transition hover:bg-gray-50/70 md:grid-cols-[34px_minmax(0,1fr)_116px_40px] md:items-start"
          :class="{ 'bg-blue-50/20 before:absolute before:inset-y-0 before:left-0 before:w-0.5 before:bg-blue-500': !item.isRead }"
        >
          <div
            class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md"
            :class="item.isRead ? 'bg-gray-50 text-gray-400' : 'bg-white shadow-sm ring-1 ring-gray-100'"
          >
            <component :is="notificationIcon(item)" class="h-4 w-4" :class="notificationTone(item)" />
          </div>
          <div class="min-w-0">
            <div class="flex min-w-0 flex-wrap items-center gap-x-1.5 gap-y-1 text-sm leading-5">
              <a v-if="actorURL(item) && item.eventType !== 'badge'" :href="actorURL(item)" class="font-semibold text-gray-950 hover:text-blue-600" @click="markItemReadAndNavigate(item)">
                {{ actorName(item) }}
              </a>
              <span v-else class="font-semibold text-gray-950">{{ item.eventType === 'follow' ? actorName(item) : item.title }}</span>
              <span class="text-gray-500">{{ item.actor.id || item.eventType === 'follow' ? notificationVerb(item) : '' }}</span>
              <a
                v-if="item.article"
                :href="item.article.url"
                class="min-w-0 max-w-full truncate font-semibold text-blue-600 hover:text-blue-700"
                @click="markItemReadAndNavigate(item)"
              >
                {{ notificationText(item) }}
              </a>
              <a
                v-else-if="item.eventType === 'badge' && targetURL(item)"
                :href="targetURL(item)"
                class="font-semibold text-amber-700 hover:text-amber-800"
                @click="markItemReadAndNavigate(item)"
              >
                {{ notificationText(item) }}
              </a>
              <a
                v-else-if="item.eventType === 'follow' && actorURL(item)"
                :href="actorURL(item)"
                class="font-semibold text-emerald-700 hover:text-emerald-800"
                @click="markItemReadAndNavigate(item)"
              >
                {{ t('notifications.viewProfile') }}
              </a>
              <span v-else-if="item.actor.id || item.eventType === 'follow'" class="font-medium text-gray-700">{{ notificationText(item) }}</span>
              <span v-if="!item.isRead" class="h-1.5 w-1.5 rounded-full bg-blue-500" />
            </div>
            <p v-if="item.content && item.content !== notificationText(item)" class="mt-0.5 line-clamp-1 text-xs text-gray-500">{{ item.content }}</p>
            <time class="mt-1 block text-xs text-gray-400 md:hidden">{{ formatDateTime(item.createdAt) }}</time>
          </div>
          <time class="hidden text-right text-xs font-medium tabular-nums text-gray-400 md:block">{{ formatDateTime(item.createdAt) }}</time>
          <button
            v-if="!item.isRead"
            type="button"
            class="absolute right-2 top-2 inline-flex h-7 w-7 items-center justify-center rounded-md text-gray-400 hover:bg-white hover:text-blue-600 md:static"
            :title="t('notifications.markRead')"
            :aria-label="t('notifications.markRead')"
            @click.stop="markItemRead(item)"
          >
            <Check class="h-4 w-4" />
          </button>
        </article>
      </div>

      <div v-else-if="activeList.loading" class="flex min-h-56 flex-col items-center justify-center px-6 text-center">
        <Bell class="h-8 w-8 text-gray-300" />
        <h2 class="mt-2 text-base font-semibold text-gray-950">{{ t('notifications.loadingMore') }}</h2>
      </div>

      <div v-else class="flex min-h-56 flex-col items-center justify-center px-6 text-center">
        <Bell class="h-8 w-8 text-gray-300" />
        <h2 class="mt-2 text-base font-semibold text-gray-950">{{ emptyTitle }}</h2>
        <p class="mt-1 text-sm text-gray-500">{{ emptyDescription }}</p>
      </div>

      <div ref="loadMoreEl" class="border-t border-gray-100 px-4 py-3 text-center text-xs font-semibold text-gray-500">
        <button
          v-if="activeList.hasNext"
          type="button"
          class="inline-flex h-8 items-center rounded-md px-3 transition hover:bg-gray-50 hover:text-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
          :disabled="activeList.loading"
          @click="loadNotifications(activeFilter)"
        >
          {{ activeList.loading ? t('notifications.loadingMore') : t('notifications.loadMore') }}
        </button>
        <span v-else-if="notifications.length">{{ t('notifications.noMore') }}</span>
      </div>
    </section>
  </main>
</template>
