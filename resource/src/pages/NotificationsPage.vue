<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Bell, CheckCheck, Heart, Info, MessageCircle, UserPlus } from '@lucide/vue'
import { markAllNotificationsRead } from '../runtime/api'
import { formatDateTime } from '../runtime/format'
import { useI18n } from '../runtime/i18n'
import { useUnreadNotifications } from '../runtime/unread-notifications'
import type { LayoutPayload, NotificationPayload, NotificationsPageProps } from '../types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: NotificationsPageProps
}>()

const notifications = ref<NotificationPayload[]>(page.props.notifications.map((item) => ({ ...item })))
const unreadCount = ref(page.props.unreadCount)
const markingAllRead = ref(false)
const actionError = ref('')
const { t } = useI18n()
const unreadNotifications = useUnreadNotifications()

const hasUnread = computed(() => unreadCount.value > 0)

if (page.props.unreadCount === 0) {
  unreadNotifications.clearUnread()
}

watch(
  () => page.props.notifications,
  (items) => {
    notifications.value = items.map((item) => ({ ...item }))
    unreadCount.value = page.props.unreadCount
    actionError.value = ''
    if (page.props.unreadCount === 0) {
      unreadNotifications.clearUnread()
    }
  },
)

function notificationIcon(item: NotificationPayload) {
  if (item.eventType === 'like') return Heart
  if (item.eventType === 'follow') return UserPlus
  if (item.eventType === 'system') return Info
  return MessageCircle
}

function notificationText(item: NotificationPayload) {
  if (item.article) {
    return item.article.title
  }
  return item.content || item.payload.content || t('notifications.fallback')
}

function notificationVerb(item: NotificationPayload) {
  if (item.eventType === 'like') return t('notifications.verb.like')
  if (item.eventType === 'follow') return t('notifications.verb.follow')
  if (item.eventType === 'reply') return t('notifications.verb.reply')
  if (item.eventType === 'comment') return t('notifications.verb.comment')
  return item.title
}

function notificationTone(item: NotificationPayload) {
  if (item.eventType === 'like') return item.isRead ? 'text-gray-400' : 'text-rose-600'
  if (item.eventType === 'follow') return item.isRead ? 'text-gray-400' : 'text-emerald-600'
  if (item.eventType === 'system') return item.isRead ? 'text-gray-400' : 'text-amber-600'
  return item.isRead ? 'text-gray-400' : 'text-blue-600'
}

function actorName(item: NotificationPayload) {
  return item.actor.username || item.payload.actorName || t('notifications.actorFallback')
}

async function markAllRead() {
  if (!hasUnread.value || markingAllRead.value) return

  const previousNotifications = notifications.value.map((item) => ({ ...item }))
  const previousUnreadCount = unreadCount.value
  markingAllRead.value = true
  actionError.value = ''
  notifications.value = notifications.value.map((item) => ({ ...item, isRead: true }))
  unreadCount.value = 0
  try {
    await markAllNotificationsRead()
    unreadNotifications.clearUnread()
  } catch (error) {
    notifications.value = previousNotifications
    unreadCount.value = previousUnreadCount
    actionError.value = error instanceof Error ? error.message : t('notifications.markAllReadFailed')
  } finally {
    markingAllRead.value = false
  }
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
          <p class="mt-0.5 text-xs text-gray-500">{{ t('notifications.summary', { total: props.total }) }}</p>
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
        <div class="hidden grid-cols-[34px_minmax(0,1fr)_116px] gap-3 border-b border-gray-100 bg-gray-50/60 px-3 py-2 text-[11px] font-bold uppercase text-gray-600 md:grid">
          <div />
          <div>{{ t('notifications.table.notification') }}</div>
          <div class="text-right">{{ t('notifications.table.time') }}</div>
        </div>

        <div v-if="notifications.length" class="divide-y divide-gray-100">
          <article
            v-for="item in notifications"
            :key="item.id"
            class="relative grid gap-3 px-3 py-2.5 transition hover:bg-gray-50/70 md:grid-cols-[34px_minmax(0,1fr)_116px] md:items-start"
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
                <a v-if="item.actor.id" :href="`/u/${item.actor.id}`" class="font-semibold text-gray-950 hover:text-blue-600">
                  {{ actorName(item) }}
                </a>
                <span v-else class="font-semibold text-gray-950">{{ item.title }}</span>
                <span class="text-gray-500">{{ item.actor.id ? notificationVerb(item) : '' }}</span>
                <a
                  v-if="item.article"
                  :href="item.article.url"
                  class="min-w-0 max-w-full truncate font-semibold text-blue-600 hover:text-blue-700"
                >
                  {{ notificationText(item) }}
                </a>
                <span v-else-if="item.actor.id" class="font-medium text-gray-700">{{ notificationText(item) }}</span>
                <span v-if="!item.isRead" class="h-1.5 w-1.5 rounded-full bg-blue-500" />
              </div>
              <p v-if="item.content && item.content !== notificationText(item)" class="mt-0.5 line-clamp-1 text-xs text-gray-500">{{ item.content }}</p>
              <time class="mt-1 block text-xs text-gray-400 md:hidden">{{ formatDateTime(item.createdAt) }}</time>
            </div>
            <time class="hidden text-right text-xs font-medium tabular-nums text-gray-400 md:block">{{ formatDateTime(item.createdAt) }}</time>
          </article>
        </div>

        <div v-else class="flex min-h-56 flex-col items-center justify-center px-6 text-center">
          <Bell class="h-8 w-8 text-gray-300" />
          <h2 class="mt-2 text-base font-semibold text-gray-950">{{ t('notifications.emptyTitle') }}</h2>
          <p class="mt-1 text-sm text-gray-500">{{ t('notifications.emptyDescription') }}</p>
        </div>
      </section>
    </main>
</template>
