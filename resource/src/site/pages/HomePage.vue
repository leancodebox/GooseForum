<script setup lang="ts">
import { computed, nextTick, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Bell, Mail, Plus, UsersRound } from '@lucide/vue'
import { fetchPage } from '@/runtime/router'
import EmptyState from '@/site/components/EmptyState.vue'
import TopicListFooter from '@/site/components/TopicListFooter.vue'
import TopicListModeSwitch from '@/site/components/TopicListModeSwitch.vue'
import TopicList from '@/site/components/TopicList.vue'
import { useTopicListMode } from '@/site/composables/useTopicListMode'
import type { HomeProps, LayoutPayload, PagePayload, TopicPayload } from '@/types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: HomeProps
  pageUrl: string
}>()
const { t } = useI18n()
const { mode: listMode, setMode: setListMode } = useTopicListMode()

const topics = ref<TopicPayload[]>([])
const pagination = ref<HomeProps['pagination']>(page.props.pagination)
const loadingMore = ref(false)
const loadError = ref('')
const loadMoreSentinel = ref<HTMLElement | null>(null)
let observer: IntersectionObserver | undefined

const hasTopics = computed(() => topics.value.length > 0)
const showPinnedLabels = computed(() => page.props.sort === '' || page.props.sort === 'latest')
const isWaterfallMode = computed(() => listMode.value === 'waterfall')

watch(
  () => page.pageUrl,
  () => {
    topics.value = [...page.props.topics]
    pagination.value = page.props.pagination
    loadError.value = ''
    void nextTick(observeSentinel)
  },
  { immediate: true },
)

watch(
  () => page.props.topics,
  (incoming) => {
    const unseenByID = new Map(incoming.map((topic) => [topic.id, topic.unseen]))
    topics.value = topics.value.map((topic) =>
      unseenByID.has(topic.id) ? { ...topic, unseen: unseenByID.get(topic.id) } : topic,
    )
  },
)

watch(listMode, (mode) => {
  if (mode === 'pagination') {
    observer?.disconnect()
    topics.value = [...page.props.topics]
    pagination.value = page.props.pagination
    return
  }
  void nextTick(observeSentinel)
})

async function loadMore() {
  if (!isWaterfallMode.value || loadingMore.value || !pagination.value.hasNext || !pagination.value.nextUrl) return

  loadingMore.value = true
  loadError.value = ''
  try {
    const payload = (await fetchPage(new URL(pagination.value.nextUrl, window.location.origin))) as PagePayload<HomeProps>
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
  if (key === 'latest') return t('topicList.tabs.latest')
  if (key === 'hot') return t('topicList.tabs.hot')
  if (key === 'popular') return t('topicList.tabs.popular')
  return fallback || key
}

function observeSentinel() {
  observer?.disconnect()
  if (!isWaterfallMode.value || !loadMoreSentinel.value || !('IntersectionObserver' in window)) return
  observer = new IntersectionObserver(
    (entries) => {
      if (entries.some((entry) => entry.isIntersecting)) void loadMore()
    },
    { rootMargin: '480px 0px' },
  )
  observer.observe(loadMoreSentinel.value)
}

onMounted(observeSentinel)
onActivated(() => {
  void nextTick(observeSentinel)
})
onDeactivated(() => {
  observer?.disconnect()
})

onBeforeUnmount(() => {
  observer?.disconnect()
})

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
          <Bell class="mt-1 h-4 w-4 shrink-0 text-primary" aria-hidden="true" />
          <div class="min-w-0 flex-1">
            <div class="gf-prose gf-prose-announcement" v-html="page.props.announcement.html" />
          </div>
        </div>
      </aside>

      <section class="gf-card overflow-hidden">
        <div class="gf-home-topic-toolbar">
          <div class="gf-home-topic-tools">
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
