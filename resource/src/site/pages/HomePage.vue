<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Bell, Mail, Plus, UsersRound } from '@lucide/vue'
import EmptyState from '@/site/components/EmptyState.vue'
import TopicListFooter from '@/site/components/TopicListFooter.vue'
import TopicListModeSwitch from '@/site/components/TopicListModeSwitch.vue'
import TopicList from '@/site/components/TopicList.vue'
import { useTopicList } from '@/site/composables/useTopicList'
import type { HomeProps, LayoutPayload } from '@gooseforum/client'

const page = defineProps<{
  layout: LayoutPayload
  props: HomeProps
  pageUrl: string
}>()
const { t } = useI18n()
const { topics, pagination, hasTopics, listMode, setListMode, loadingMore, loadError, loadMoreSentinel, loadMore } = useTopicList(page)

const announcementReadStorageKey = 'goose:announcement:last-read-published-at'
const announcementReminderWindow = 7 * 24 * 60 * 60 * 1000
const announcementUnread = ref(shouldRemindAnnouncement())
const showPinnedLabels = computed(() => page.props.sort === '' || page.props.sort === 'latest')

watch(
  () => [page.props.announcement.enabled, page.props.announcement.publishedAt] as const,
  () => refreshAnnouncementReminder(),
)

function sortTabLabel(key: string, fallback?: string) {
  if (key === 'latest') return t('topicList.tabs.latest')
  if (key === 'hot') return t('topicList.tabs.hot')
  if (key === 'popular') return t('topicList.tabs.popular')
  return fallback || key
}

function markAnnouncementRead() {
  announcementUnread.value = false
  const publishedAt = parseAnnouncementTime(page.props.announcement.publishedAt)
  if (!Number.isFinite(publishedAt)) return
  try {
    window.localStorage.setItem(announcementReadStorageKey, String(publishedAt))
  } catch {
    // Storage may be unavailable in private or restricted browsing contexts.
  }
}

function shouldRemindAnnouncement() {
  if (!page.props.announcement.enabled) return false
  const publishedAt = parseAnnouncementTime(page.props.announcement.publishedAt)
  if (!Number.isFinite(publishedAt)) return false
  const age = Date.now() - publishedAt
  if (age < 0 || age > announcementReminderWindow) return false

  try {
    const lastReadAt = Number(window.localStorage.getItem(announcementReadStorageKey) || 0)
    return !Number.isFinite(lastReadAt) || lastReadAt < publishedAt
  } catch {
    return true
  }
}

function parseAnnouncementTime(value?: string) {
  return Date.parse((value || '').replace(' ', 'T'))
}

function refreshAnnouncementReminder() {
  announcementUnread.value = shouldRemindAnnouncement()
}

function syncAnnouncementRead(event: StorageEvent) {
  if (event.key === announcementReadStorageKey) refreshAnnouncementReminder()
}

onMounted(() => window.addEventListener('storage', syncAnnouncementRead))
onBeforeUnmount(() => window.removeEventListener('storage', syncAnnouncementRead))

</script>

<template>
    <div class="pb-12">
      <aside
        v-if="page.layout.viewer.requiresEmailVerification"
        class="gf-email-verification mb-0 border-y border-warning/30 bg-warning/10 sm:-mt-3 sm:mb-3 sm:rounded-b-lg sm:border-x sm:border-t-0"
        :aria-label="t('topicList.emailVerification.title')"
      >
        <div class="flex items-center gap-2 px-3 py-2 text-[13px] leading-5 text-warning sm:px-4 sm:text-sm">
          <Mail class="h-4 w-4 shrink-0 text-warning" />
          <div class="min-w-0 flex-1">
            <span class="font-semibold text-warning">{{ t('topicList.emailVerification.title') }}</span>
            <span class="mx-1 text-warning">·</span>
            <span>{{ t('topicList.emailVerification.description') }}</span>
          </div>
          <a href="/settings" class="shrink-0 font-semibold text-warning hover:text-warning">
            {{ t('topicList.emailVerification.action') }}
          </a>
        </div>
      </aside>

      <aside
        v-if="page.props.announcement.enabled"
        class="gf-panel gf-announcement-panel mb-0 border-l-2 border-l-primary/45 bg-base-100 px-3 py-2 sm:mb-3 sm:px-4 sm:py-2.5"
        :aria-label="t('topicList.announcement')"
      >
        <div class="flex items-start gap-2 sm:gap-2.5">
          <button
            v-if="announcementUnread"
            type="button"
            class="-mx-1 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded text-primary transition hover:bg-primary/10 focus-visible:bg-primary/10 focus-visible:outline-none"
            :title="t('topicList.markAnnouncementRead')"
            :aria-label="t('topicList.markAnnouncementRead')"
            @click="markAnnouncementRead"
          >
            <Bell class="announcement-unread-bell h-4 w-4" />
          </button>
          <Bell v-else class="mt-1 h-4 w-4 shrink-0 text-primary" aria-hidden="true" />
          <div class="min-w-0 flex-1">
            <div class="gf-prose gf-prose-announcement" v-html="page.props.announcement.html" />
          </div>
        </div>
      </aside>

      <section class="gf-card overflow-hidden">
        <div class="gf-home-topic-toolbar">
          <div class="gf-home-topic-tools">
            <nav class="gf-home-topic-tabs" :aria-label="t('topicList.columns.topic')">
              <a
                v-for="tab in page.props.tabs"
                :key="tab.key"
                :href="tab.url"
                :aria-current="tab.active ? 'page' : undefined"
                class="gf-tab"
                :class="tab.active ? 'gf-tab-active' : 'gf-tab-idle'"
              >
                {{ sortTabLabel(tab.key, tab.label) }}
              </a>
            </nav>
            <TopicListModeSwitch :model-value="listMode" @update:model-value="setListMode" />
          </div>
          <a href="/publish" class="gf-button gf-button-md gf-button-primary shrink-0 whitespace-nowrap px-3 sm:h-8">
            <Plus class="h-4 w-4" />
            {{ t('topicList.newTopic') }}
          </a>
        </div>

        <TopicList :topics="topics" home :show-pinned="showPinnedLabels">
          <template #empty>
            <EmptyState v-if="!hasTopics" :icon="UsersRound" :title="t('topicList.emptyTitle')" :description="t('topicList.emptyDescription')" />
          </template>
        </TopicList>

        <div ref="loadMoreSentinel">
          <TopicListFooter
            :pagination="pagination"
            :mode="listMode"
            :loading-more="loadingMore"
            :has-topics="hasTopics"
            :load-error="loadError"
            @load-more="loadMore"
          />
        </div>
      </section>
    </div>
</template>

<style scoped>
.announcement-unread-bell {
  animation: announcement-bell-ring 2s ease-in-out infinite;
  transform-origin: 50% 15%;
}

@keyframes announcement-bell-ring {
  0%, 30%, 100% {
    transform: rotate(0) scale(1);
  }
  5% {
    transform: rotate(20deg) scale(1.12);
  }
  10% { transform: rotate(-18deg) scale(1.12); }
  15% {
    transform: rotate(14deg) scale(1.08);
  }
  20% { transform: rotate(-10deg) scale(1.05); }
  25% { transform: rotate(5deg) scale(1.02); }
}

@media (prefers-reduced-motion: reduce) {
  .announcement-unread-bell {
    animation: none;
  }
}
</style>
