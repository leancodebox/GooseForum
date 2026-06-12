<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Loader2, MessageSquare, Plus, UsersRound } from '@lucide/vue'
import { formatNumber, timeAgo } from '@/runtime/format'
import { fetchPage } from '@/runtime/router'
import { topicDescription } from '@/runtime/topic-description'
import { showUserCard } from '@/runtime/user-card-events'
import PageHeader from '@/site/components/PageHeader.vue'
import UserAvatar from '@/site/components/UserAvatar.vue'
import type { CategoryPageProps, LayoutPayload, PagePayload, TopicPayload } from '@/types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: CategoryPageProps
}>()
const { t } = useI18n()

const topics = ref<TopicPayload[]>([])
const pagination = ref<CategoryPageProps['pagination']>(page.props.pagination)
const loadingMore = ref(false)
const loadError = ref('')
const loadMoreSentinel = ref<HTMLElement | null>(null)
let observer: IntersectionObserver | undefined

const hasTopics = computed(() => topics.value.length > 0)

watch(
  () => [page.props.category.id, page.props.sort, page.props.pagination.page, page.props.topics],
  () => {
    topics.value = [...page.props.topics]
    pagination.value = page.props.pagination
    loadError.value = ''
    void nextTick(observeSentinel)
  },
  { immediate: true },
)

async function loadMore() {
  if (loadingMore.value || !pagination.value.hasNext || !pagination.value.nextUrl) return

  loadingMore.value = true
  loadError.value = ''
  try {
    const payload = (await fetchPage(new URL(pagination.value.nextUrl, window.location.origin))) as PagePayload<CategoryPageProps>
    topics.value = mergeTopics(topics.value, payload.props.topics)
    pagination.value = payload.props.pagination
  } catch (error) {
    loadError.value = error instanceof Error ? error.message : t('common.loadFailed')
  } finally {
    loadingMore.value = false
  }
}

function mergeTopics(current: TopicPayload[], incoming: TopicPayload[]) {
  const seen = new Set(current.map((topic) => topic.id))
  return [...current, ...incoming.filter((topic) => !seen.has(topic.id))]
}

function sortTabLabel(key: string, fallback?: string) {
  if (key === 'latest') return t('topicList.tabs.latestReplies')
  if (key === 'new') return t('topicList.tabs.latestPublished')
  if (key === 'hot') return t('topicList.tabs.hot')
  if (key === 'popular') return t('topicList.tabs.popular')
  return fallback || key
}

function observeSentinel() {
  observer?.disconnect()
  if (!loadMoreSentinel.value || !('IntersectionObserver' in window)) return
  observer = new IntersectionObserver(
    (entries) => {
      if (entries.some((entry) => entry.isIntersecting)) void loadMore()
    },
    { rootMargin: '480px 0px' },
  )
  observer.observe(loadMoreSentinel.value)
}

onMounted(observeSentinel)

onBeforeUnmount(() => {
  observer?.disconnect()
})

</script>

<template>
    <div class="pb-12">
      <PageHeader :title="page.props.category.name" :description="page.props.category.description" compact>
        <template #badge>
          <span class="h-1.5 w-1.5 shrink-0 rounded-full" :style="{ backgroundColor: page.props.category.color || 'var(--gf-color-primary)' }" />
          <span class="text-xs font-semibold uppercase text-base-content/75">{{ t('category.label') }}</span>
        </template>
      </PageHeader>

      <section class="gf-card overflow-hidden">
        <div class="gf-home-topic-toolbar">
          <div class="gf-home-topic-tabs">
            <a
              v-for="tab in page.props.tabs"
              :key="tab.key"
              :href="tab.url"
              class="gf-tab"
              :class="tab.active ? 'gf-tab-active' : 'gf-tab-idle'"
            >
              {{ sortTabLabel(tab.key, tab.label) }}
            </a>
          </div>
          <a href="/publish" class="gf-button gf-button-md gf-button-primary shrink-0 whitespace-nowrap px-3 sm:h-8">
            <Plus class="h-4 w-4" />
            {{ t('topicList.newTopic') }}
          </a>
        </div>

        <div class="gf-topic-list-header">
          <div>{{ t('topicList.columns.topic') }}</div>
          <div class="text-center">{{ t('topicList.columns.users') }}</div>
          <div class="text-center">{{ t('topicList.columns.replies') }}</div>
          <div class="text-center">{{ t('topicList.columns.views') }}</div>
          <div class="text-right">{{ t('topicList.columns.activity') }}</div>
        </div>

        <div class="relative bg-base-100">
          <article
            v-for="topic in topics"
            :key="topic.id"
            class="group gf-topic-row"
          >
            <div class="min-w-0">
              <div class="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
                <a :href="topic.url" class="truncate text-[15px] font-semibold leading-snug text-base-content group-hover:text-primary sm:text-base">
                  {{ topic.title }}
                </a>
              </div>
              <p class="mt-1 truncate text-[13px] leading-relaxed text-base-content/55">{{ topicDescription(topic) }}</p>
              <div class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-base-content/55 lg:hidden">
                <div class="flex -space-x-2">
                  <a
                    v-for="participant in topic.participants"
                    :key="participant.id"
                    :href="`/u/${participant.id}`"
                    :title="participant.username"
                    class="rounded-full ring-2 ring-base-100 transition hover:z-10 hover:scale-110"
                    @click="showUserCard(participant, $event)"
                  >
                    <UserAvatar :src="participant.avatarUrl" :alt="participant.username" class="h-6 w-6 rounded-full object-cover" />
                  </a>
                </div>
                <span>{{ timeAgo(topic.lastUpdateTime) }}</span>
                <span class="inline-flex items-center gap-1">
                  <MessageSquare class="h-3.5 w-3.5" /> {{ formatNumber(topic.replyCount) }}
                </span>
              </div>
            </div>
            <div class="hidden justify-center lg:flex">
              <div class="flex -space-x-3">
                <a
                  v-for="participant in topic.participants"
                  :key="participant.id"
                  :href="`/u/${participant.id}`"
                  :title="participant.username"
                  class="rounded-full ring-2 ring-base-100 transition hover:z-10 hover:scale-110"
                  @click="showUserCard(participant, $event)"
                >
                  <UserAvatar :src="participant.avatarUrl" :alt="participant.username" class="h-8 w-8 rounded-full object-cover" />
                </a>
              </div>
            </div>
            <div class="hidden text-center text-sm font-semibold tabular-nums text-base-content/75 lg:block">{{ formatNumber(topic.replyCount) }}</div>
            <div class="hidden text-center text-sm tabular-nums text-base-content/55 lg:block">{{ formatNumber(topic.viewCount) }}</div>
            <div class="hidden text-right text-[13px] font-medium tabular-nums text-base-content/55 lg:block">{{ timeAgo(topic.lastUpdateTime) }}</div>
          </article>

          <div v-if="!hasTopics" class="px-5 py-16 text-center">
            <UsersRound class="mx-auto h-8 w-8 text-base-content/35" />
            <h2 class="mt-3 text-base font-semibold text-base-content">{{ t('topicList.emptyTitle') }}</h2>
            <p class="mt-1 text-sm text-base-content/55">{{ t('topicList.emptyCategoryDescription') }}</p>
          </div>
        </div>

        <div ref="loadMoreSentinel" class="border-t border-line bg-base-200/50 p-3 text-center">
          <button
            v-if="pagination.hasNext"
            type="button"
            class="gf-button gf-button-sm gf-button-ghost gap-2 disabled:cursor-wait"
            :disabled="loadingMore"
            @click="loadMore"
          >
            <Loader2 v-if="loadingMore" class="h-4 w-4 animate-spin" />
            {{ loadingMore ? t('common.loadFailed') : t('common.loadMore') }}
          </button>
          <p v-else-if="hasTopics" class="text-xs font-medium text-base-content/55">{{ t('topicList.allShown') }}</p>
          <p v-if="loadError" class="mt-2 text-xs text-error">{{ t('topicList.autoLoadFailed') }}</p>
          <a v-if="pagination.hasNext" :href="pagination.nextUrl" rel="next" class="sr-only">{{ t('common.nextPage') }}</a>
        </div>
      </section>
    </div>
</template>
