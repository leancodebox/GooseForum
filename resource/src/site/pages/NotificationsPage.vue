<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { Award, Bell, Check, CheckCheck, Info, MessageCircle, UserPlus } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { fetchNotifications, markAllNotificationsRead, markNotificationRead } from '@/runtime/api'
import { formatDateTime } from '@/runtime/format'
import { useUnreadStatus } from '@/runtime/unread-status'
import EmptyState from '@/site/components/EmptyState.vue'
import PageHeader from '@/site/components/PageHeader.vue'
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

function notificationTitleText(item: NotificationPayload) {
  return item.title || t('notifications.fallback')
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
  if (item.eventType === 'follow') return t('notifications.verb.follow')
  if (item.eventType === 'badge') return ''
  if (item.eventType === 'reply') return t('notifications.verb.reply')
  if (item.eventType === 'comment' || item.eventType === 'article_comment') return t('notifications.verb.comment')
  return notificationTitleText(item)
}

function notificationTone(item: NotificationPayload) {
  if (item.eventType === 'follow') return item.isRead ? 'text-base-content/55' : 'text-success'
  if (item.eventType === 'badge') return item.isRead ? 'text-base-content/55' : 'text-warning'
  if (item.eventType === 'system') return item.isRead ? 'text-base-content/55' : 'text-warning'
  return item.isRead ? 'text-base-content/55' : 'text-primary'
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
    <PageHeader :title="t('notifications.title')" :description="t('notifications.summary', { total: loadedCount })" compact>
      <template #badge>
        <span v-if="unreadCount" class="gf-badge gf-badge-info h-5 tabular-nums">
          {{ t('notifications.unread', { count: unreadCount }) }}
        </span>
      </template>
      <template #meta>
        <p v-if="actionError" class="mt-1 text-xs text-error">{{ actionError }}</p>
      </template>
      <template #actions>
        <button
          type="button"
          class="gf-button gf-button-sm gf-button-secondary w-fit text-xs disabled:opacity-45"
          :disabled="!hasUnread || markingAllRead"
          @click="markAllRead"
        >
          <CheckCheck class="h-4 w-4" />
          {{ markingAllRead ? t('common.loading') : t('notifications.markAllRead') }}
        </button>
      </template>
    </PageHeader>

    <section class="gf-card overflow-hidden">
      <div class="flex items-center gap-1 border-b border-line bg-base-200/60 p-2">
        <button
          type="button"
          class="gf-tab"
          :class="activeFilter === 'all' ? 'bg-base-100 text-base-content shadow-sm ring-1 ring-line' : 'text-base-content/55 hover:bg-base-100/70 hover:text-base-content'"
          @click="setActiveFilter('all')"
        >
          {{ t('notifications.tabs.all') }}
        </button>
        <button
          type="button"
          class="gf-tab gap-1.5"
          :class="activeFilter === 'unread' ? 'bg-base-100 text-base-content shadow-sm ring-1 ring-line' : 'text-base-content/55 hover:bg-base-100/70 hover:text-base-content'"
          @click="setActiveFilter('unread')"
        >
          {{ t('notifications.tabs.unread') }}
          <span v-if="unreadCount" class="gf-badge gf-badge-info px-1.5 text-[11px] font-bold tabular-nums">{{ unreadCount }}</span>
        </button>
      </div>

      <div class="hidden grid-cols-[34px_minmax(0,1fr)_116px] gap-3 border-b border-line bg-base-200/60 px-3 py-2 text-[11px] font-bold uppercase text-base-content/75 md:grid">
        <div />
        <div>{{ t('notifications.table.notification') }}</div>
        <div class="text-right">{{ t('notifications.table.time') }}</div>
      </div>

      <div v-if="notifications.length" class="divide-y divide-line">
        <article
          v-for="item in notifications"
          :key="item.id"
          class="relative grid grid-cols-[34px_minmax(0,1fr)] gap-3 px-3 py-2.5 transition hover:bg-base-200/70 md:grid-cols-[34px_minmax(0,1fr)_116px_40px] md:items-start"
          :class="{ 'bg-info/10 before:absolute before:inset-y-0 before:left-0 before:w-0.5 before:bg-primary': !item.isRead }"
        >
          <div
            class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md"
            :class="item.isRead ? 'bg-base-200 text-base-content/55' : 'bg-base-100 shadow-sm ring-1 ring-line'"
          >
            <component :is="notificationIcon(item)" class="h-4 w-4" :class="notificationTone(item)" />
          </div>
          <div class="min-w-0">
            <div class="flex min-w-0 items-center gap-1.5 text-sm leading-5">
              <a v-if="actorURL(item) && item.eventType !== 'badge'" :href="actorURL(item)" class="max-w-[42%] shrink-0 truncate font-semibold text-base-content hover:text-primary" @click="markItemReadAndNavigate(item)">
                {{ actorName(item) }}
              </a>
              <span v-else class="max-w-[42%] shrink-0 truncate font-semibold text-base-content">{{ item.eventType === 'follow' ? actorName(item) : notificationTitleText(item) }}</span>
              <span class="shrink-0 text-base-content/55">{{ item.actor.id || item.eventType === 'follow' ? notificationVerb(item) : '' }}</span>
              <a
                v-if="item.article"
                :href="item.article.url"
                class="min-w-0 max-w-full truncate font-semibold text-primary hover:text-primary"
                @click="markItemReadAndNavigate(item)"
              >
                {{ notificationText(item) }}
              </a>
              <a
                v-else-if="item.eventType === 'badge' && targetURL(item)"
                :href="targetURL(item)"
                class="font-semibold text-warning hover:text-warning"
                @click="markItemReadAndNavigate(item)"
              >
                {{ notificationText(item) }}
              </a>
              <a
                v-else-if="item.eventType === 'follow' && actorURL(item)"
                :href="actorURL(item)"
                class="font-semibold text-success hover:text-success/90"
                @click="markItemReadAndNavigate(item)"
              >
                {{ t('notifications.viewProfile') }}
              </a>
              <span v-else-if="item.actor.id || item.eventType === 'follow'" class="font-medium text-base-content/75">{{ notificationText(item) }}</span>
              <span v-if="!item.isRead" class="h-1.5 w-1.5 rounded-full bg-primary" />
            </div>
            <p v-if="item.content && item.content !== notificationText(item)" class="mt-0.5 line-clamp-1 text-xs text-base-content/55">{{ item.content }}</p>
            <time class="mt-1 block text-xs text-base-content/55 md:hidden">{{ formatDateTime(item.createdAt) }}</time>
          </div>
          <time class="hidden text-right text-xs font-medium tabular-nums text-base-content/55 md:block">{{ formatDateTime(item.createdAt) }}</time>
          <button
            v-if="!item.isRead"
            type="button"
            class="absolute right-2 top-2 inline-flex h-7 w-7 items-center justify-center rounded-md text-icon-muted hover:bg-base-100 hover:text-primary md:static"
            :title="t('notifications.markRead')"
            :aria-label="t('notifications.markRead')"
            @click.stop="markItemRead(item)"
          >
            <Check class="h-4 w-4" />
          </button>
        </article>
      </div>

      <EmptyState v-else-if="activeList.loading" :icon="Bell" :title="t('notifications.loadingMore')" loading />

      <EmptyState v-else :icon="Bell" :title="emptyTitle" :description="emptyDescription" />

      <div ref="loadMoreEl" class="border-t border-line px-4 py-3 text-center text-xs font-semibold text-base-content/55">
        <button
          v-if="activeList.hasNext"
          type="button"
          class="gf-button gf-button-sm gf-button-ghost"
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
